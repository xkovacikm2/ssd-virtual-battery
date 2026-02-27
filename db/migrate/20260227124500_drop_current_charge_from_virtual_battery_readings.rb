class DropCurrentChargeFromVirtualBatteryReadings < ActiveRecord::Migration[8.1]
  def up
    remove_column :virtual_battery_readings, :current_charge, :decimal
  end

  def down
    add_column :virtual_battery_readings, :current_charge, :decimal, precision: 10, scale: 2, default: 0.0

    execute <<~SQL
      WITH ordered AS (
        SELECT
          id,
          SUM(COALESCE(exported_to_grid, 0) - COALESCE(imported_from_grid, 0))
            OVER (ORDER BY date, id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running_charge
        FROM virtual_battery_readings
      )
      UPDATE virtual_battery_readings v
      SET current_charge = ordered.running_charge
      FROM ordered
      WHERE v.id = ordered.id
    SQL
  end
end
