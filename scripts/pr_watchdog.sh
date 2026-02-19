#!/usr/bin/env bash
set -euo pipefail

# Lightweight placeholder watchdog runner.
# Prevents cron failures when this script path is invoked.

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

echo "[pr_watchdog] script found at $0"
echo "[pr_watchdog] repo: $REPO_ROOT"
echo "[pr_watchdog] no default PR configured; exiting cleanly"
