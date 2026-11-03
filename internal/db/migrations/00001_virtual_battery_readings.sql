-- +goose Up
-- IF NOT EXISTS lets this adopt the table Rails already created in production.
CREATE TABLE IF NOT EXISTS virtual_battery_readings (
  id bigserial PRIMARY KEY,
  created_at timestamp(6) NOT NULL DEFAULT now(),
  date date NOT NULL,
  exported_to_grid numeric(10,2) DEFAULT 0.0,
  imported_from_grid numeric(10,2) DEFAULT 0.0,
  updated_at timestamp(6) NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS index_virtual_battery_readings_on_date ON virtual_battery_readings (date);
ALTER TABLE virtual_battery_readings ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE virtual_battery_readings ALTER COLUMN updated_at SET DEFAULT now();

-- +goose Down
DROP TABLE IF EXISTS virtual_battery_readings;
