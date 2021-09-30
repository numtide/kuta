#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
export CGO_ENABLED=0
go build -a -tags netgo -ldflags '-w -extldflags "-static"' "$@" .
