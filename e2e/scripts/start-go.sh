#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
npm --prefix web run build
mkdir -p tmp
go build -o tmp/ssdvb-e2e ./cmd/ssdvb
exec tmp/ssdvb-e2e serve
