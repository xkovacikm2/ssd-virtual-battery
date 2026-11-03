# SSD Virtual Battery Dashboard

A Go and Lit dashboard for daily electricity imports and exports from the SSD virtual battery service. PostgreSQL stores one reading per date. The dashboard displays year-to-date totals, a projection chart, and daily, weekly, and monthly readings.

## Development

Open the repository in its VS Code devcontainer. It provides Go, Node.js, PostgreSQL 16, and Playwright. For a local setup, install Go (see `go.mod`), Node.js LTS, and PostgreSQL 16, then create development, test, and end-to-end databases.

Set `DATABASE_URL` to the development PostgreSQL connection string. The devcontainer sets it automatically. Set `IMS_SSD_USERNAME`, `IMS_SSD_PASSWORD`, and `IMS_SSD_POINT_OF_DELIVERY_ID` for SSD access. For example, copy `.env.template` to `.env`, fill in the values, then explicitly load it in your shell:

```bash
set -a
source .env
set +a
```

The Go application does not automatically load `.env`. Never commit credentials.

Install and build the frontend before building or starting Go, because the binary embeds the built files:

```bash
npm --prefix web ci
npm --prefix web run build
go run ./cmd/ssdvb serve
```

Open http://localhost:3000. For frontend development, run `npm --prefix web run dev` in a second terminal and open http://localhost:5173; Vite proxies API requests to the Go server on port 3000.

## Data And Migrations

Opening the dashboard synchronizes missing days through yesterday, in UTC. If SSD is unavailable, the dashboard still shows existing data. To synchronize manually:

```bash
go run ./cmd/ssdvb sync
```

`serve` runs migrations on startup; `go run ./cmd/ssdvb migrate` runs them separately. Back up the database before upgrading from Rails: migration 00002 removes Rails metadata and any old Solid Queue tables and is **irreversible**. Restore the backup to roll back to Rails. The migration preserves `virtual_battery_readings` and fails rather than cascading into unrelated dependent objects.

Daily incoming and outgoing measurements are stored as `imported_from_grid` and `exported_to_grid`, respectively. The year-to-date balance is `min(total exported, 6000) - total imported`, in kWh. The dashboard has no authentication; protect access at the deployment boundary if needed.

## Tests

Set `TEST_DATABASE_URL` and `E2E_DATABASE_URL` to separate disposable PostgreSQL databases. The Go tests create and drop isolated schemas; the browser suite deletes readings in its end-to-end database.

```bash
npm --prefix web run build
go test ./...
npm --prefix web test
npm --prefix e2e ci
cd e2e && npx playwright test
```

The browser suite uses a fake SSD server. For a first local run, install Chromium with `npx playwright install --with-deps chromium` from `e2e/`. Use test-driven development: write a failing test, implement the behavior, and rerun the focused test.

## Configuration

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string, required |
| `PORT` | HTTP port, defaults to 3000 |
| `IMS_SSD_USERNAME` | SSD account username |
| `IMS_SSD_PASSWORD` | SSD account password |
| `IMS_SSD_POINT_OF_DELIVERY_ID` | SSD point of delivery ID |
| `IMS_SSD_BASE_URL` | Optional SSD API URL override, primarily for testing |

SSD credentials are only needed for synchronization; missing credentials do not prevent the dashboard from serving existing data.