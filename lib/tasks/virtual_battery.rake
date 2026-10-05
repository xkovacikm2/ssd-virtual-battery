namespace :virtual_battery do
  desc "Collect virtual battery data from SSD API for all missing days"
  task collect_data: :environment do
    puts "Starting virtual battery data collection..."
    saved = VirtualBatteryReading.sync_missing_readings!
    puts "Virtual battery data collection completed. Saved #{saved} reading(s)."
  end
end
