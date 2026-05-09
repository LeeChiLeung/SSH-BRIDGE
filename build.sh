#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ $# -eq 0 ]]; then
  exec "$ROOT_DIR/scripts/release.sh" all
fi

exec "$ROOT_DIR/scripts/release.sh" "$@"
