class CreateVirtualBatteryReadings < ActiveRecord::Migration[8.1]
  def change
    create_table :virtual_battery_readings do |t|
      t.date :date, null: false
      t.decimal :current_charge, precision: 10, scale: 2, default: 0.0
      t.decimal :exported_to_battery, precision: 10, scale: 2, default: 0.0
      t.decimal :imported_from_battery, precision: 10, scale: 2, default: 0.0
      t.decimal :imported_from_grid, precision: 10, scale: 2, default: 0.0

      t.timestamps
    end
    
    add_index :virtual_battery_readings, :date, unique: true
  end
end
