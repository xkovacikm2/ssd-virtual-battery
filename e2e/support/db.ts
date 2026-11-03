import pg from "pg";
import { databaseUrl } from "./env";

export type Reading = { date: string; exported: number; imported: number };

const pool = new pg.Pool({ connectionString: databaseUrl, max: 2 });

export async function resetReadings(readings: Reading[] = []): Promise<void> {
  const client = await pool.connect();
  try {
    await client.query("BEGIN");
    await client.query("DELETE FROM virtual_battery_readings");
    for (const r of readings) {
      await client.query(
        `INSERT INTO virtual_battery_readings (date, exported_to_grid, imported_from_grid, created_at, updated_at)
         VALUES ($1, $2, $3, now(), now())`,
        [r.date, r.exported, r.imported],
      );
    }
    await client.query("COMMIT");
  } catch (e) {
    await client.query("ROLLBACK");
    throw e;
  } finally {
    client.release();
  }
}

export async function storedReadings(): Promise<Reading[]> {
  const { rows } = await pool.query(
    `SELECT to_char(date, 'YYYY-MM-DD') AS date, exported_to_grid::float8 AS exported, imported_from_grid::float8 AS imported
     FROM virtual_battery_readings ORDER BY date`,
  );
  return rows;
}
