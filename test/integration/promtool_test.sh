#!/bin/sh
# Manual exposition check: pipes a live scrape through promtool.
#   ./promtool_test.sh [host:port]
set -eu
ADDR=${1:-127.0.0.1:9167}
command -v promtool >/dev/null 2>&1 || { echo "promtool not in PATH" >&2; exit 1; }
curl -fsS "http://$ADDR/metrics" | promtool check metrics
echo "promtool: metrics OK"
