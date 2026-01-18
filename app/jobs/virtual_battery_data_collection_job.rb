class VirtualBatteryDataCollectionJob < ApplicationJob
  queue_as :default

  def perform
    ssd_client = SsdApiClient.new
    profile_data = ssd_client.fetch_profile_data_for_date  # defaults to yesterday
  rescue StandardError => e
    Rails.logger.error "Failed to collect virtual battery data: #{e.message}"
    raise
  end
end
