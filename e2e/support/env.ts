export const appPort = Number(process.env.E2E_APP_PORT ?? 3100);
export const ssdPort = Number(process.env.FAKE_SSD_PORT ?? 4010);
export const ssdUrl = `http://127.0.0.1:${ssdPort}`;
export const databaseUrl =
  process.env.E2E_DATABASE_URL ??
  `postgres://postgres:postgres@${process.env.DATABASE_HOST ?? "localhost"}:5432/ssd_virtual_battery_e2e`;
