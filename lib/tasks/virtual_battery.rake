namespace :virtual_battery do
  desc "Collect virtual battery data from SSD API for all missing days"
  task collect_data: :environment do
    puts "Starting virtual battery data collection..."

    # Determine start date: day after last reading, or first day of current year
    last_reading = VirtualBatteryReading.order(:date).last
    start_date = if last_reading
                   last_reading.date + 1.day
    else
                   Date.current.beginning_of_year
    end

    end_date = Date.current - 1.day # yesterday

    if start_date > end_date
      puts "No new days to process. Last reading is from #{last_reading&.date || 'N/A'}."
      next
    end

    puts "Processing data from #{start_date} to #{end_date}..."

    # Initialize SSD API client
    ssd_client = SsdApiClient.new

    # Get the current charge from the last reading (or 0 if none)
    current_charge = last_reading&.current_charge || 0.0

    # Process each day
    (start_date..end_date).each do |date|
      puts "Fetching data for #{date}..."

      begin
        # Fetch profile data for the day
        profile_data = ssd_client.fetch_profile_data_for_date(date)

        # Sum up all 15-minute intervals for the day
        # incoming = consumption from grid/battery (actualConsumption)
        # outgoing = production/export to battery (actualSupply)
        total_incoming = profile_data.sum { |row| row[:incoming].to_f/4 } # convert from 15-min to hourly
        total_outgoing = profile_data.sum { |row| row[:outgoing].to_f/4 } # convert from 15-min to hourly

        # Calculate battery transactions:
        # - outgoing goes INTO the virtual battery (exported_to_battery)
        # - incoming comes FROM the virtual battery first, then grid if depleted
        exported_to_battery = total_outgoing
        current_charge += exported_to_battery

        # Try to satisfy incoming demand from battery first
        imported_from_battery = 0.0
        imported_from_grid = 0.0

        if total_incoming <= current_charge
          # Battery can cover all consumption
          imported_from_battery = total_incoming
          current_charge -= total_incoming
        else
          # Battery partially covers consumption, rest from grid
          imported_from_battery = current_charge
          imported_from_grid = total_incoming - current_charge
          current_charge = 0.0
        end

        # Create or update the reading for this day
        reading = VirtualBatteryReading.find_or_initialize_by(date: date)
        reading.assign_attributes(
          current_charge: current_charge.round(2),
          exported_to_battery: exported_to_battery.round(2),
          imported_from_battery: imported_from_battery.round(2),
          imported_from_grid: imported_from_grid.round(2)
        )
        reading.save!

        puts "  ✓ Created reading for #{date}: charge=#{current_charge.round(2)} kWh, " \
             "exported=#{exported_to_battery.round(2)}, imported_battery=#{imported_from_battery.round(2)}, " \
             "imported_grid=#{imported_from_grid.round(2)}"

      rescue StandardError => e
        puts "  ✗ Error processing #{date}: #{e.message}"
        raise
      end
    end

    puts "Virtual battery data collection completed."
  end
end
