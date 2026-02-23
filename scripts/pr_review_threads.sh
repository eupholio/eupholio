#!/usr/bin/env bash
set -euo pipefail

# List unresolved review threads for PRs.
# Default: only actionable unresolved threads (non-outdated).
# Use --include-outdated to also show outdated unresolved threads.

REPO="${REPO:-eupholio/eupholio}"
STATE="open"
LIMIT="${LIMIT:-50}"
INCLUDE_OUTDATED=0
MAX_THREAD_PAGES="${MAX_THREAD_PAGES:-20}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo) REPO="$2"; shift 2 ;;
    --state) STATE="$2"; shift 2 ;;
    --limit) LIMIT="$2"; shift 2 ;;
    --include-outdated) INCLUDE_OUTDATED=1; shift ;;
    --max-thread-pages) MAX_THREAD_PAGES="$2"; shift 2 ;;
    -h|--help)
      cat <<'EOF'
Usage: scripts/pr_review_threads.sh [options]

Options:
  --repo <owner/name>      GitHub repo (default: eupholio/eupholio)
  --state <open|merged|closed|all>
  --limit <n>              max PRs to scan (default: 50)
  --include-outdated       include unresolved outdated threads
  --max-thread-pages <n>   max GraphQL pages per PR for review threads (default: 20)
  -h, --help               show this help
EOF
      exit 0 ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

if ! [[ "$REPO" =~ ^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$ ]]; then
  echo "Invalid REPO format: '$REPO'. Expected 'owner/name'." >&2
  exit 1
fi
if ! [[ "$LIMIT" =~ ^[1-9][0-9]*$ ]]; then
  echo "Invalid LIMIT: '$LIMIT'. Expected positive integer." >&2
  exit 1
fi
if ! [[ "$MAX_THREAD_PAGES" =~ ^[1-9][0-9]*$ ]]; then
  echo "Invalid MAX_THREAD_PAGES: '$MAX_THREAD_PAGES'. Expected positive integer." >&2
  exit 1
fi
case "$STATE" in
  open|merged|closed|all) ;;
  *)
    echo "Invalid STATE: '$STATE'. Expected one of: open|merged|closed|all." >&2
    exit 1
    ;;
esac

for cmd in gh jq; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "Missing required command: $cmd" >&2; exit 1; }
done
gh auth status >/dev/null 2>&1 || { echo "gh not authenticated" >&2; exit 1; }

OWNER="${REPO%%/*}"
NAME="${REPO##*/}"

if [[ "$STATE" == "all" ]]; then
  states=(open merged closed)
else
  states=("$STATE")
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
: > "$tmp"

THREAD_QUERY='query($owner:String!, $repo:String!, $num:Int!, $after:String) {
  repository(owner:$owner,name:$repo){
    pullRequest(number:$num){
      number
      title
      state
      url
      reviewThreads(first:100, after:$after){
        nodes{
          isResolved
          isOutdated
          comments(last:1){
            nodes{ author{login} path body url }
          }
        }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}'

append_pr_threads() {
  local prn="$1"
  local after=""
  local pages=0

  while :; do
    pages=$((pages + 1))
    if (( pages > MAX_THREAD_PAGES )); then
      echo "WARNING: PR #$prn exceeded MAX_THREAD_PAGES=$MAX_THREAD_PAGES, stopping pagination." >&2
      return 0
    fi

    local raw
    if ! raw=$(gh api graphql -f query="$THREAD_QUERY" -F owner="$OWNER" -F repo="$NAME" -F num="$prn" -F after="$after" 2>/dev/null); then
      echo "WARNING: failed to fetch reviewThreads for PR #$prn" >&2
      return 0
    fi

    if ! jq -e '.data.repository.pullRequest != null' >/dev/null <<<"$raw"; then
      echo "WARNING: unexpected GraphQL payload for PR #$prn" >&2
      return 0
    fi

    jq --argjson include_outdated "$INCLUDE_OUTDATED" -c '
      .data.repository.pullRequest as $pr
      | ($pr.reviewThreads.nodes // [])[]
      | select(.isResolved == false)
      | select(($include_outdated == 1) or (.isOutdated == false))
      | {
          number: $pr.number,
          title: $pr.title,
          state: $pr.state,
          pr_url: $pr.url,
          outdated: .isOutdated,
          author: (.comments.nodes[0].author.login // "unknown"),
          path: (.comments.nodes[0].path // ""),
          body: ((.comments.nodes[0].body // "") | gsub("\n"; " ") | .[0:220]),
          thread_url: (.comments.nodes[0].url // "")
        }
    ' <<<"$raw" >> "$tmp"

    local has_next
    has_next=$(jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.hasNextPage // false' <<<"$raw")
    if [[ "$has_next" != "true" ]]; then
      break
    fi

    after=$(jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.endCursor // empty' <<<"$raw")
    [[ -z "$after" ]] && break
  done
}

for st in "${states[@]}"; do
  pr_list=$(gh pr list --repo "$REPO" --state "$st" --limit "$LIMIT" --json number --jq '.[].number' 2>/dev/null || true)
  [[ -z "$pr_list" ]] && continue

  while IFS= read -r prn; do
    [[ -z "$prn" ]] && continue
    append_pr_threads "$prn"
  done <<< "$pr_list"
done

if [[ ! -s "$tmp" ]]; then
  echo "No unresolved review threads found (state=$STATE, include_outdated=$INCLUDE_OUTDATED)."
  exit 0
fi

jq -sr '
  sort_by(.number)
  | group_by(.number)
  | .[]
  | "PR #\(.[0].number) [\(.[0].state)] \(.[0].title)\n\(.[0].pr_url)\n"
    + (map("  - [\(.author)] \(.path)\(if .outdated then " (outdated)" else "" end)\n    \(.body)\n    \(.thread_url)") | join("\n"))
' "$tmp"
