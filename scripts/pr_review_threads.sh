#!/usr/bin/env bash
set -euo pipefail

# List unresolved review threads for PRs.
# Default: only actionable unresolved threads (non-outdated).
# Use --include-outdated to also show outdated unresolved threads.

REPO="${REPO:-eupholio/eupholio}"
STATE="open"
LIMIT="${LIMIT:-50}"
INCLUDE_OUTDATED=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo) REPO="$2"; shift 2 ;;
    --state) STATE="$2"; shift 2 ;;
    --limit) LIMIT="$2"; shift 2 ;;
    --include-outdated) INCLUDE_OUTDATED=1; shift ;;
    -h|--help)
      cat <<'EOF'
Usage: scripts/pr_review_threads.sh [options]

Options:
  --repo <owner/name>      GitHub repo (default: eupholio/eupholio)
  --state <open|merged|closed|all>
  --limit <n>              max PRs to scan (default: 50)
  --include-outdated       include unresolved outdated threads
  -h, --help               show this help
EOF
      exit 0 ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

for cmd in gh jq; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "Missing: $cmd" >&2; exit 1; }
done
gh auth status >/dev/null 2>&1 || { echo "gh not authenticated" >&2; exit 1; }

owner="${REPO%%/*}"
name="${REPO##*/}"

if [[ "$STATE" == "all" ]]; then
  states=(open merged closed)
else
  states=("$STATE")
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
: > "$tmp"

THREAD_QUERY='query($owner:String!, $repo:String!, $num:Int!){
  repository(owner:$owner,name:$repo){
    pullRequest(number:$num){
      number
      title
      state
      url
      reviewThreads(first:100){
        nodes{
          isResolved
          isOutdated
          comments(first:1){
            nodes{ author{login} path body url }
          }
        }
      }
    }
  }
}'

for st in "${states[@]}"; do
  pr_list=$(gh pr list --repo "$REPO" --state "$st" --limit "$LIMIT" --json number --jq '.[].number')
  [[ -z "$pr_list" ]] && continue

  while IFS= read -r prn; do
    [[ -z "$prn" ]] && continue
    raw=$(gh api graphql -f query="$THREAD_QUERY" -F owner="$owner" -F repo="$name" -F num="$prn")
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
  done <<< "$pr_list"
done

if [[ ! -s "$tmp" ]]; then
  echo "No unresolved review threads found (state=$STATE, include_outdated=$INCLUDE_OUTDATED)."
  exit 0
fi

jq -s '
  sort_by(.number)
  | group_by(.number)
  | .[]
  | "PR #\(.[0].number) [\(.[0].state)] \(.[0].title)\n\(.[0].pr_url)\n"
    + (map("  - [\(.author)] \(.path)\(if .outdated then " (outdated)" else "" end)\n    \(.body)\n    \(.thread_url)") | join("\n"))
' "$tmp" | sed 's/^"//; s/"$//; s/\\n/\
/g'
