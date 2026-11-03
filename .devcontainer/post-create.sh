#!/usr/bin/env bash
set -euo pipefail

for db in ssd_virtual_battery_test ssd_virtual_battery_e2e; do
  psql "postgres://postgres:postgres@db:5432/postgres" -tAc "SELECT 1 FROM pg_database WHERE datname='$db'" | grep -q 1 \
    || psql "postgres://postgres:postgres@db:5432/postgres" -c "CREATE DATABASE $db"
done

[[ -f go.mod ]] && go mod download
[[ -f web/package.json ]] && npm --prefix web ci
if [[ -f e2e/package.json ]]; then
  npm --prefix e2e ci
  (cd e2e && npx playwright install --with-deps chromium)
fi
