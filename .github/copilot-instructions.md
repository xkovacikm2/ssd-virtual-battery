# Copilot Instructions

SSD Virtual Battery is a Go backend with a Lit and TypeScript frontend. PostgreSQL 16 stores one `virtual_battery_readings` row per date. Go serves the built frontend embedded from `web/dist`.

- Use test-driven development: write a failing test, implement, then rerun the focused test.
- Run `npm --prefix web run build` before `go test ./...` or building the Go binary.
- Run `go test ./...`, `npm --prefix web test`, and `cd e2e && npx playwright test` for full verification. Set `TEST_DATABASE_URL` and `E2E_DATABASE_URL` to separate disposable databases; the e2e suite deletes readings in its database.
- Sync is on demand from the dashboard or `go run ./cmd/ssdvb sync`; it is implemented in `internal/readings` and uses the SSD client in `internal/ssd`.
- Incoming SSD energy maps to `imported_from_grid`, outgoing to `exported_to_grid`. Yearly net flow is `min(total exported, 6000) - total imported` in kWh.
- Use `DATABASE_URL`, `IMS_SSD_USERNAME`, `IMS_SSD_PASSWORD`, and `IMS_SSD_POINT_OF_DELIVERY_ID`. Never commit `.env` or credentials.
- Migrations are embedded in `internal/db/migrations`. Migration 00002 irreversibly removes Rails metadata and Solid Queue tables; back up before upgrading and do not cascade into unrelated data.