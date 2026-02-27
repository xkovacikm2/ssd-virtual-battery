class MigrateVirtualBatteryGridFlowColumns < ActiveRecord::Migration[8.1]
  def up
    add_column :virtual_battery_readings, :exported_to_grid, :decimal, precision: 10, scale: 2, default: 0.0

    execute <<~SQL
      UPDATE virtual_battery_readings
      SET exported_to_grid = COALESCE(exported_to_battery, 0)
    SQL

    remove_column :virtual_battery_readings, :imported_from_battery, :decimal
    remove_column :virtual_battery_readings, :exported_to_battery, :decimal
  end

  def down
    add_column :virtual_battery_readings, :exported_to_battery, :decimal, precision: 10, scale: 2, default: 0.0
    add_column :virtual_battery_readings, :imported_from_battery, :decimal, precision: 10, scale: 2, default: 0.0

    execute <<~SQL
      UPDATE virtual_battery_readings
      SET exported_to_battery = COALESCE(exported_to_grid, 0)
    SQL

    remove_column :virtual_battery_readings, :exported_to_grid, :decimal
  end
end
