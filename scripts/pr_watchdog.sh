#!/bin/bash
set -euo pipefail

REPO="${REPO:-eupholio/eupholio}"
PR_LIST_LIMIT="${PR_LIST_LIMIT:-200}"
MAX_THREAD_PAGES="${MAX_THREAD_PAGES:-100}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORKSPACE_ROOT="${WORKSPACE_ROOT:-$(cd "$REPO_ROOT/.." && pwd)}"
REVIEW_AUDIT_SCRIPT="${REVIEW_AUDIT_SCRIPT:-$REPO_ROOT/scripts/pr_review_threads.sh}"
REVIEW_AUDIT_LIMIT="${REVIEW_AUDIT_LIMIT:-20}"
REVIEW_AUDIT_MAX_LINES="${REVIEW_AUDIT_MAX_LINES:-40}"

if ! [[ "$REPO" =~ ^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$ ]]; then
  echo "Invalid REPO format: '$REPO'. Expected 'owner/name'." >&2
  exit 1
fi
if ! [[ "$PR_LIST_LIMIT" =~ ^[1-9][0-9]*$ ]]; then
  echo "Invalid PR_LIST_LIMIT: '$PR_LIST_LIMIT'. Expected positive integer." >&2
  exit 1
fi
if ! [[ "$MAX_THREAD_PAGES" =~ ^[0-9]+$ ]] || [[ "$MAX_THREAD_PAGES" -le 0 ]]; then
  echo "Invalid MAX_THREAD_PAGES: '$MAX_THREAD_PAGES'. Expected positive integer." >&2
  exit 1
fi
if ! [[ "$REVIEW_AUDIT_LIMIT" =~ ^[1-9][0-9]*$ ]]; then
  echo "Invalid REVIEW_AUDIT_LIMIT: '$REVIEW_AUDIT_LIMIT'. Expected positive integer." >&2
  exit 1
fi
if ! [[ "$REVIEW_AUDIT_MAX_LINES" =~ ^[1-9][0-9]*$ ]]; then
  echo "Invalid REVIEW_AUDIT_MAX_LINES: '$REVIEW_AUDIT_MAX_LINES'. Expected positive integer." >&2
  exit 1
fi

OWNER="${REPO%%/*}"
NAME="${REPO##*/}"

# STATE_FILE priority:
# 1) explicit STATE_FILE
# 2) $WORKSPACE_ROOT/memory (shared workspace)
# 3) cache fallback ($XDG_CACHE_HOME or $TMPDIR)
if [[ -n "${STATE_FILE:-}" ]]; then
  state_dir="$(dirname "$STATE_FILE")"
  if ! mkdir -p "$state_dir" 2>/dev/null; then
    echo "Failed to create STATE_FILE directory: $state_dir" >&2
    exit 1
  fi
  if ! : >"$state_dir/.pr_watchdog_write_test" 2>/dev/null; then
    echo "STATE_FILE directory is not writable: $state_dir" >&2
    exit 1
  fi
  rm -f "$state_dir/.pr_watchdog_write_test" 2>/dev/null || true
else
  DEFAULT_STATE_DIR_OUT="$WORKSPACE_ROOT/memory"
  if mkdir -p "$DEFAULT_STATE_DIR_OUT" 2>/dev/null && : >"$DEFAULT_STATE_DIR_OUT/.pr_watchdog_write_test" 2>/dev/null; then
    rm -f "$DEFAULT_STATE_DIR_OUT/.pr_watchdog_write_test" 2>/dev/null || true
    STATE_FILE="$DEFAULT_STATE_DIR_OUT/pr-watchdog-state.json"
  else
    CACHE_BASE="${XDG_CACHE_HOME:-${TMPDIR:-/tmp}}"
    DEFAULT_STATE_DIR_FALLBACK="$CACHE_BASE/pr_watchdog"
    if mkdir -p "$DEFAULT_STATE_DIR_FALLBACK" 2>/dev/null && : >"$DEFAULT_STATE_DIR_FALLBACK/.pr_watchdog_write_test" 2>/dev/null; then
      rm -f "$DEFAULT_STATE_DIR_FALLBACK/.pr_watchdog_write_test" 2>/dev/null || true
      STATE_FILE="$DEFAULT_STATE_DIR_FALLBACK/pr-watchdog-state.json"
    else
      echo "Failed to find a writable directory for 'pr_watchdog' state file." >&2
      exit 1
    fi
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

LOCK_FILE="${STATE_FILE}.lock"
LOCK_DIR="${STATE_FILE}.lockdir"
LOCK_MODE=""
LOCK_WAIT_TIMEOUT="${LOCK_WAIT_TIMEOUT:-30}"
cleanup_lock() {
  if [[ "$LOCK_MODE" == "flock" ]]; then
    exec 9>&-
  elif [[ "$LOCK_MODE" == "mkdir" ]]; then
    rmdir "$LOCK_DIR" 2>/dev/null || true
  fi
}
trap cleanup_lock EXIT
# Prefer flock when available; fallback to mkdir lock for environments without flock.
if command -v flock >/dev/null 2>&1; then
  exec 9>"$LOCK_FILE"
  flock -x 9
  LOCK_MODE="flock"
else
  start_time="$(date +%s)"
  while ! mkdir "$LOCK_DIR" 2>/dev/null; do
    now="$(date +%s)"
    elapsed=$((now - start_time))
    if (( elapsed >= LOCK_WAIT_TIMEOUT )); then
      echo "Failed to acquire mkdir lock '$LOCK_DIR' after ${LOCK_WAIT_TIMEOUT}s" >&2
      exit 1
    fi
    sleep 0.1
  done
  LOCK_MODE="mkdir"
fi

if [[ ! -f "$STATE_FILE" ]] || ! jq -e . "$STATE_FILE" >/dev/null 2>&1; then
  echo '{"prs":{}}' > "$STATE_FILE"
fi

PARTIAL_FAILURE=0
set_partial_failure() { PARTIAL_FAILURE=1; }

NO_REPLY_SENTINEL="NO_REPLY"

GQL_THREADS_FIRST100='query($owner:String!, $name:String!, $number:Int!) {
  repository(owner:$owner, name:$name) {
    pullRequest(number:$number) {
      reviewThreads(first:100) {
        nodes { isResolved }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}'

GQL_THREADS_FIRST100_AFTER='query($owner:String!, $name:String!, $number:Int!, $after:String!) {
  repository(owner:$owner, name:$name) {
    pullRequest(number:$number) {
      reviewThreads(first:100, after:$after) {
        nodes { isResolved }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}'

count_unresolved_threads() {
  local pr_number="$1"
  local after=""
  local total=0
  local pages=0

  while :; do
    pages=$((pages + 1))
    if (( pages > MAX_THREAD_PAGES )); then
      return 1
    fi
    local page
    if [[ -z "$after" ]]; then
      if ! page=$(gh api graphql \
        -f query="$GQL_THREADS_FIRST100" \
        -F owner="$OWNER" -F name="$NAME" -F number="$pr_number" 2>/dev/null); then
        return 1
      fi
    else
      if ! page=$(gh api graphql \
        -f query="$GQL_THREADS_FIRST100_AFTER" \
        -F owner="$OWNER" -F name="$NAME" -F number="$pr_number" -F after="$after" 2>/dev/null); then
        return 1
      fi
    fi

    local page_unresolved
    if ! page_unresolved=$(echo "$page" | jq '[.data.repository.pullRequest.reviewThreads.nodes[]? | select(.isResolved==false)] | length' 2>/dev/null); then
      return 1
    fi
    total=$((total + page_unresolved))

    local has_next
    if ! has_next=$(echo "$page" | jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.hasNextPage // false' 2>/dev/null); then
      return 1
    fi
    if [[ "$has_next" != "true" ]]; then
      break
    fi

    if ! after=$(echo "$page" | jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.endCursor // empty' 2>/dev/null); then
      return 1
    fi
    [[ -z "$after" ]] && break
  done

  echo "$total"
}

pr_lines=""
list_ok=1
if ! pr_lines=$(gh pr list --repo "$REPO" --state open --limit "$PR_LIST_LIMIT" --json number,url --jq '.[] | "\(.number)\t\(.url)"' 2>/dev/null); then
  list_ok=0
  set_partial_failure
  echo "WARNING: PR list fetch failed; using potentially stale state from '$STATE_FILE'." >&2
fi

prev_state="$(cat "$STATE_FILE" 2>/dev/null || echo '{"prs":{}}')"
new_state='{"prs":{}}'
alerts=""

if [[ "$list_ok" -eq 1 ]]; then
  if [[ -n "$pr_lines" ]]; then
    while IFS=$'\t' read -r pr_number pr_url; do
    [[ -z "${pr_number:-}" ]] && continue
    if ! [[ "$pr_number" =~ ^[0-9]+$ ]]; then
      set_partial_failure
      echo "WARNING: Skipping PR with invalid number '$pr_number' (url: $pr_url)" >&2
      continue
    fi

    if ! prev_unresolved=$(echo "$prev_state" | jq -r --arg n "$pr_number" '.prs[$n].lastUnresolvedCount // 0' 2>/dev/null); then
      set_partial_failure
      prev_unresolved=0
    fi
    if ! prev_failing=$(echo "$prev_state" | jq -c --arg n "$pr_number" '.prs[$n].lastFailingChecks // []' 2>/dev/null); then
      set_partial_failure
      prev_failing='[]'
    fi
    if ! prev_sig=$(echo "$prev_state" | jq -r --arg n "$pr_number" '.prs[$n].lastNotifiedSignature // ""' 2>/dev/null); then
      set_partial_failure
      prev_sig=""
    fi

    failing_ok=1
    failing='[]'
    if ! failing=$(gh pr view "$pr_number" --repo "$REPO" --json statusCheckRollup --jq '
      [.statusCheckRollup[]?
        # Only completed, actionable failures are included; pending/in-progress
        # checks have null conclusion and are intentionally ignored.
        | select(.conclusion == "FAILURE"
              or .conclusion == "STARTUP_FAILURE"
              or .conclusion == "TIMED_OUT"
              or .conclusion == "ACTION_REQUIRED")
        | (.name // .workflowName // "unknown")]
      | unique | sort' 2>/dev/null); then
      set_partial_failure
      failing_ok=0
      failing="$prev_failing"
    fi
    if ! echo "$failing" | jq -e . >/dev/null 2>&1; then
      set_partial_failure
      failing_ok=0
      failing="$prev_failing"
    fi

    unresolved_ok=1
    unresolved="$prev_unresolved"
    if ! unresolved=$(count_unresolved_threads "$pr_number"); then
      set_partial_failure
      unresolved_ok=0
      unresolved="$prev_unresolved"
    fi
    if ! [[ "$unresolved" =~ ^[0-9]+$ ]]; then
      set_partial_failure
      unresolved_ok=0
      unresolved="$prev_unresolved"
    fi

    new_failing='[]'
    if ! new_failing=$(jq -n --argjson cur "$failing" --argjson prev "$prev_failing" '($cur - $prev) | unique | sort' 2>/dev/null); then
      set_partial_failure
      new_failing='[]'
    fi
    if ! new_fail_count=$(echo "$new_failing" | jq 'length' 2>/dev/null); then
      set_partial_failure
      new_fail_count=0
    fi

    need_alert=0
    reasons=""
    if [[ "$new_fail_count" -gt 0 ]]; then
      need_alert=1
      new_failing_names=$(echo "$new_failing" | jq -r 'join(", ")' 2>/dev/null || printf '')
      if [[ -n "$new_failing_names" ]]; then
        reasons="New failing checks: $new_failing_names"
      else
        reasons="New failing checks detected (unable to list names)"
      fi
    fi
    if [[ "$unresolved" -gt "$prev_unresolved" ]]; then
      need_alert=1
      [[ -n "$reasons" ]] && reasons+=" / "
      reasons+="Unresolved reviews increased: ${prev_unresolved}->${unresolved}"
    fi

    if ! sig=$(jq -nc --arg n "$pr_number" --argjson f "$failing" --argjson u "$unresolved" '{prNumber:($n|tonumber),failingChecks:$f,unresolvedCount:$u}' 2>/dev/null); then
      set_partial_failure
      sig="$prev_sig"
    fi
    norm_sig_ok=0
    norm_prev_sig_ok=0
    if norm_sig_tmp=$(printf '%s' "$sig" | jq -cS '.' 2>/dev/null); then
      norm_sig="$norm_sig_tmp"
      norm_sig_ok=1
    else
      set_partial_failure
      norm_sig="$sig"
    fi
    if norm_prev_sig_tmp=$(printf '%s' "$prev_sig" | jq -cS '.' 2>/dev/null); then
      norm_prev_sig="$norm_prev_sig_tmp"
      norm_prev_sig_ok=1
    else
      set_partial_failure
      norm_prev_sig="$prev_sig"
    fi
    if [[ "$need_alert" -eq 1 ]]; then
      if [[ "$norm_sig_ok" -eq 0 || "$norm_prev_sig_ok" -eq 0 || "$norm_sig" != "$norm_prev_sig" ]]; then
        alerts+=$'\n- PR #'
        alerts+="$pr_number $pr_url"
        alerts+=$'\n  Reason: '
        alerts+="$reasons"
      fi
    fi

    # Keep previous signature only if this PR had transient fetch failures.
    if [[ "$failing_ok" -eq 0 || "$unresolved_ok" -eq 0 ]]; then
      last_sig="$prev_sig"
    else
      last_sig="$sig"
    fi

    if ! tmp_state=$(echo "$new_state" | jq --arg n "$pr_number" --argjson u "$unresolved" --argjson f "$failing" --arg sig "$last_sig" '.prs[$n]={lastUnresolvedCount:$u,lastFailingChecks:$f,lastNotifiedSignature:$sig}' 2>/dev/null); then
      set_partial_failure
      continue
    fi
    new_state="$tmp_state"
    done <<< "$pr_lines"
  fi

  if ! tmp_state_file=$(mktemp "$(dirname "$STATE_FILE")/pr-watchdog-state.XXXXXX"); then
    echo "Failed to create temporary state file in $(dirname "$STATE_FILE")" >&2
    exit 1
  fi
  if ! printf '%s\n' "$new_state" > "$tmp_state_file"; then
    rm -f "$tmp_state_file"
    echo "Failed to write state file: $tmp_state_file" >&2
    exit 1
  fi
  if ! mv "$tmp_state_file" "$STATE_FILE"; then
    rm -f "$tmp_state_file"
    echo "Failed to move state file into place: $STATE_FILE" >&2
    exit 1
  fi
fi

if [[ -n "$alerts" ]]; then
  extra_audit=""
  if [[ -x "$REVIEW_AUDIT_SCRIPT" ]]; then
    if audit_output=$("$REVIEW_AUDIT_SCRIPT" --repo "$REPO" --state open --limit "$REVIEW_AUDIT_LIMIT" 2>/dev/null); then
      if [[ -n "$audit_output" && "$audit_output" != No\ unresolved\ review\ threads* ]]; then
        extra_audit=$'\n\nActionable unresolved review threads (open PRs):\n'
        extra_audit+="$(printf '%s\n' "$audit_output" | sed -n "1,${REVIEW_AUDIT_MAX_LINES}p")"
      fi
    else
      set_partial_failure
    fi
  fi

  echo "Action required PRs detected:${alerts}${extra_audit}"
elif [[ "$PARTIAL_FAILURE" -eq 1 ]]; then
  echo "PR watchdog had partial errors (continuing)"
else
  # Contract: heartbeat/cron wrapper treats this sentinel as "no user-facing update".
  echo "$NO_REPLY_SENTINEL"
fi
