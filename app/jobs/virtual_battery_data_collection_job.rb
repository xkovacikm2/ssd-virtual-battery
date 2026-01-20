class VirtualBatteryDataCollectionJob < ApplicationJob
  queue_as :default

  def perform
    last_reading = VirtualBatteryReading.order(:date).last
    current_charge = last_reading&.current_charge || 0.0

    ssd_client = SsdApiClient.new
    profile_data = ssd_client.fetch_profile_data_for_date

    # Create or update reading using model logic
    result = VirtualBatteryReading.create_from_profile_data(
      date: date,
      profile_data: profile_data,
      current_charge: current_charge
    )
    reading = result[:reading]

    Rails.logger.info "Created reading for #{date}: charge=#{reading.current_charge} kWh, " \
                      "exported=#{reading.exported_to_battery}, imported_battery=#{reading.imported_from_battery}, " \
                      "imported_grid=#{reading.imported_from_grid}"
  rescue StandardError => e
    Rails.logger.error "Failed to collect virtual battery data: #{e.message}"
    raise
  end
end
