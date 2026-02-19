#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-eupholio/eupholio}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORKSPACE_ROOT="${WORKSPACE_ROOT:-$(cd "$REPO_ROOT/.." && pwd)}"

if ! [[ "$REPO" =~ ^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$ ]]; then
  echo "Invalid REPO format: '$REPO'. Expected 'owner/name'." >&2
  exit 1
fi

OWNER="${REPO%%/*}"
NAME="${REPO##*/}"

if [[ -n "${STATE_FILE:-}" ]]; then
  :
else
  DEFAULT_STATE_DIR_OUT="$WORKSPACE_ROOT/memory"
  DEFAULT_STATE_DIR_IN="$REPO_ROOT/.cache"
  if mkdir -p "$DEFAULT_STATE_DIR_OUT" 2>/dev/null; then
    STATE_FILE="$DEFAULT_STATE_DIR_OUT/pr-watchdog-state.json"
  else
    mkdir -p "$DEFAULT_STATE_DIR_IN"
    STATE_FILE="$DEFAULT_STATE_DIR_IN/pr-watchdog-state.json"
  fi
fi

for cmd in gh jq; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Missing required command: $cmd" >&2
    exit 1
  fi
done

if ! gh auth status >/dev/null 2>&1; then
  echo "gh not authenticated" >&2
  exit 1
fi

mkdir -p "$(dirname "$STATE_FILE")"
if [[ ! -f "$STATE_FILE" ]] || ! jq -e . "$STATE_FILE" >/dev/null 2>&1; then
  echo '{"prs":{}}' > "$STATE_FILE"
fi

WARN=0
warn_once() { WARN=1; }

count_unresolved_threads() {
  local pr_number="$1"
  local after=""
  local total=0

  while :; do
    local page
    if [[ -z "$after" ]]; then
      if ! page=$(gh api graphql \
        -f query='query($owner:String!, $name:String!, $number:Int!){ repository(owner:$owner, name:$name){ pullRequest(number:$number){ reviewThreads(first:100){ nodes{ isResolved } pageInfo{ hasNextPage endCursor } } } } }' \
        -F owner="$OWNER" -F name="$NAME" -F number="$pr_number" 2>/dev/null); then
        return 1
      fi
    else
      if ! page=$(gh api graphql \
        -f query='query($owner:String!, $name:String!, $number:Int!, $after:String){ repository(owner:$owner, name:$name){ pullRequest(number:$number){ reviewThreads(first:100, after:$after){ nodes{ isResolved } pageInfo{ hasNextPage endCursor } } } } }' \
        -F owner="$OWNER" -F name="$NAME" -F number="$pr_number" -F after="$after" 2>/dev/null); then
        return 1
      fi
    fi

    local page_unresolved
    page_unresolved=$(echo "$page" | jq '[.data.repository.pullRequest.reviewThreads.nodes[]? | select(.isResolved==false)] | length')
    total=$((total + page_unresolved))

    local has_next
    has_next=$(echo "$page" | jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.hasNextPage // false')
    if [[ "$has_next" != "true" ]]; then
      break
    fi

    after=$(echo "$page" | jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.endCursor // empty')
    [[ -z "$after" ]] && break
  done

  echo "$total"
}

pr_lines=""
list_ok=1
if ! pr_lines=$(gh pr list --repo "$REPO" --state open --json number,url --jq '.[] | "\(.number)\t\(.url)"' 2>/dev/null); then
  list_ok=0
  warn_once
fi

prev_state="$(cat "$STATE_FILE" 2>/dev/null || echo '{"prs":{}}')"
new_state='{"prs":{}}'
alerts=""

if [[ "$list_ok" -eq 1 ]]; then
  while IFS=$'\t' read -r pr_number pr_url; do
    [[ -z "${pr_number:-}" ]] && continue

    prev_unresolved=$(echo "$prev_state" | jq -r --arg n "$pr_number" '.prs[$n].lastUnresolvedCount // 0')
    prev_failing=$(echo "$prev_state" | jq -c --arg n "$pr_number" '.prs[$n].lastFailingChecks // []')
    prev_sig=$(echo "$prev_state" | jq -r --arg n "$pr_number" '.prs[$n].lastNotifiedSignature // ""')

    failing_ok=1
    failing='[]'
    if ! failing=$(gh pr view "$pr_number" --repo "$REPO" --json statusCheckRollup --jq '
      [.statusCheckRollup[]?
        | select(.conclusion == "FAILURE"
              or .conclusion == "STARTUP_FAILURE"
              or .conclusion == "TIMED_OUT"
              or .conclusion == "CANCELLED"
              or .conclusion == "ACTION_REQUIRED")
        | (.name // .workflowName // "unknown")]
      | unique | sort' 2>/dev/null); then
      warn_once
      failing_ok=0
      failing="$prev_failing"
    fi
    if ! echo "$failing" | jq -e . >/dev/null 2>&1; then
      warn_once
      failing_ok=0
      failing="$prev_failing"
    fi

    unresolved_ok=1
    unresolved="$prev_unresolved"
    if ! unresolved=$(count_unresolved_threads "$pr_number"); then
      warn_once
      unresolved_ok=0
      unresolved="$prev_unresolved"
    fi
    if ! [[ "$unresolved" =~ ^[0-9]+$ ]]; then
      warn_once
      unresolved_ok=0
      unresolved="$prev_unresolved"
    fi

    new_fail_count=$(jq -n --argjson cur "$failing" --argjson prev "$prev_failing" '($cur - $prev) | length' 2>/dev/null || echo 0)

    need_alert=0
    reasons=""
    if [[ "$new_fail_count" -gt 0 ]]; then
      need_alert=1
      reasons="New failing checks: $(echo "$failing" | jq -r 'join(", ")')"
    fi
    if [[ "$unresolved" -gt "$prev_unresolved" ]]; then
      need_alert=1
      [[ -n "$reasons" ]] && reasons+=" / "
      reasons+="Unresolved reviews increased: ${prev_unresolved}->${unresolved}"
    fi

    if ! sig=$(jq -nc --arg n "$pr_number" --argjson f "$failing" --argjson u "$unresolved" '{prNumber:($n|tonumber),failingChecks:$f,unresolvedCount:$u}' 2>/dev/null); then
      warn_once
      sig="$prev_sig"
    fi
    last_sig="$prev_sig"
    if [[ "$need_alert" -eq 1 ]] && [[ "$sig" != "$prev_sig" ]]; then
      alerts+=$'\n- PR #'
      alerts+="$pr_number $pr_url"
      alerts+=$'\n  Reason: '
      alerts+="$reasons"
      last_sig="$sig"
    fi

    # Keep previous signature if this PR had transient fetch failures.
    if [[ "$failing_ok" -eq 0 || "$unresolved_ok" -eq 0 ]]; then
      last_sig="$prev_sig"
    fi

    if ! tmp_state=$(echo "$new_state" | jq --arg n "$pr_number" --argjson u "$unresolved" --argjson f "$failing" --arg sig "$last_sig" '.prs[$n]={lastUnresolvedCount:$u,lastFailingChecks:$f,lastNotifiedSignature:$sig}' 2>/dev/null); then
      warn_once
      continue
    fi
    new_state="$tmp_state"
  done <<< "$pr_lines"

  printf '%s\n' "$new_state" > "${STATE_FILE}.tmp" && mv "${STATE_FILE}.tmp" "$STATE_FILE"
fi

if [[ -n "$alerts" ]]; then
  echo "Action required PRs detected:${alerts}"
elif [[ "$WARN" -eq 1 ]]; then
  echo "PR watchdog had partial errors (continuing)"
else
  echo "NO_REPLY"
fi
