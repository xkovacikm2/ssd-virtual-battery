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

        # Create or update reading using model logic
        result = VirtualBatteryReading.create_from_profile_data(
          date: date,
          profile_data: profile_data,
          current_charge: current_charge
        )
        reading = result[:reading]
        current_charge = result[:current_charge]

        puts "  ✓ Created reading for #{date}: charge=#{reading.current_charge} kWh, " \
             "exported=#{reading.exported_to_battery}, imported_battery=#{reading.imported_from_battery}, " \
             "imported_grid=#{reading.imported_from_grid}"

      rescue StandardError => e
        puts "  ✗ Error processing #{date}: #{e.message}"
        raise
      end
    end

    puts "Virtual battery data collection completed."
  end
end
