#!/bin/bash
set -euo pipefail

# PR watchdog
# - Detects actionable deltas for one PR (CI failures, new review comments, unresolved threads)
# - Prints NO_REPLY when no change

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

STATE_FILE="${PR_WATCH_STATE_FILE:-/home/kinakao/.openclaw/workspace/memory/pr-watchdog-state.json}"
mkdir -p "$(dirname "$STATE_FILE")"

if ! command -v gh >/dev/null 2>&1; then
  echo "watchdog error: gh not found"
  exit 0
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "watchdog error: jq not found"
  exit 0
fi

if ! gh auth status >/dev/null 2>&1; then
  if [ -f "$STATE_FILE" ] && [ "$(jq -r '.lastError // ""' "$STATE_FILE" 2>/dev/null)" = "gh_auth" ]; then
    echo "NO_REPLY"
    exit 0
  fi
  jq -n '{lastError:"gh_auth",updatedAt:(now|floor)}' > "$STATE_FILE"
  echo "watchdog error: gh auth required"
  exit 0
fi

repo="${PR_WATCH_REPO:-}"
if [ -z "$repo" ]; then
  repo="$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || true)"
fi
if [ -z "$repo" ]; then
  echo "watchdog error: repo not found"
  exit 0
fi

pr="${PR_WATCH_PR:-}"
if [ -z "$pr" ]; then
  pr="$(gh pr list --repo "$repo" --state open --limit 1 --json number -q '.[0].number' 2>/dev/null || true)"
fi
if [ -z "$pr" ] || [ "$pr" = "null" ]; then
  echo "NO_REPLY"
  exit 0
fi

owner="${repo%/*}"
name="${repo#*/}"

checks_json="$(gh pr checks "$pr" --repo "$repo" --json name,state,link 2>/dev/null || echo '[]')"
comments_json="$(gh api "repos/$repo/pulls/$pr/comments?per_page=100" 2>/dev/null || echo '[]')"
threads_json="$(gh api graphql -f query='query($owner:String!,$name:String!,$number:Int!){repository(owner:$owner,name:$name){pullRequest(number:$number){reviewThreads(first:100){nodes{id isResolved}}}}}' -F owner="$owner" -F name="$name" -F number="$pr" 2>/dev/null || echo '{}')"

failed_count="$(jq '[.[] | select(.state != "SUCCESS" and .state != "SKIPPED" and .state != "NEUTRAL")] | length' <<<"$checks_json" 2>/dev/null || echo 0)"
latest_comment_id="$(jq '([.[].id] | max) // 0' <<<"$comments_json" 2>/dev/null || echo 0)"
unresolved_count="$(jq '.data.repository.pullRequest.reviewThreads.nodes | map(select(.isResolved == false)) | length // 0' <<<"$threads_json" 2>/dev/null || echo 0)"

if [ -f "$STATE_FILE" ]; then
  last_failed="$(jq -r '.last.failedCount // 0' "$STATE_FILE" 2>/dev/null || echo 0)"
  last_comment_id="$(jq -r '.last.latestCommentId // 0' "$STATE_FILE" 2>/dev/null || echo 0)"
  last_unresolved="$(jq -r '.last.unresolvedCount // 0' "$STATE_FILE" 2>/dev/null || echo 0)"
else
  last_failed=0
  last_comment_id=0
  last_unresolved=0
fi

new_comments=0
if [ "$latest_comment_id" -gt "$last_comment_id" ]; then
  new_comments="$(jq --argjson last "$last_comment_id" '[.[] | select(.id > $last)] | length' <<<"$comments_json" 2>/dev/null || echo 0)"
fi

has_change=0
[ "$failed_count" -ne "$last_failed" ] && has_change=1
[ "$new_comments" -gt 0 ] && has_change=1
[ "$unresolved_count" -ne "$last_unresolved" ] && has_change=1

pr_url="https://github.com/$repo/pull/$pr"

jq -n \
  --arg repo "$repo" \
  --argjson pr "$pr" \
  --argjson failed "$failed_count" \
  --argjson latest "$latest_comment_id" \
  --argjson unresolved "$unresolved_count" \
  '{lastError:null,repo:$repo,pr:$pr,last:{failedCount:$failed,latestCommentId:$latest,unresolvedCount:$unresolved},updatedAt:(now|floor)}' > "$STATE_FILE"

if [ "$has_change" -eq 0 ]; then
  echo "NO_REPLY"
  exit 0
fi

echo "PR Watchdog: $repo#$pr"
echo "CI failing: $failed_count (prev $last_failed)"
echo "New review comments: $new_comments"
echo "Unresolved threads: $unresolved_count (prev $last_unresolved)"
echo "$pr_url"
