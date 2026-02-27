class VirtualBatteryDataCollectionJob < ApplicationJob
  queue_as :default

  def perform
    ssd_client = SsdApiClient.new
    profile_data = ssd_client.fetch_profile_data_for_date

    # Create or update reading using model logic
    result = VirtualBatteryReading.create_from_profile_data(
      date: Date.yesterday,
      profile_data: profile_data
    )
    reading = result[:reading]

    Rails.logger.info "Created reading for yesterday: exported_grid=#{reading.exported_to_grid}, " \
                      "imported_grid=#{reading.imported_from_grid}"
  rescue StandardError => e
    Rails.logger.error "Failed to collect virtual battery data: #{e.message}"
    raise
  end
end
