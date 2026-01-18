class VirtualBatteryDataCollectionJob < ApplicationJob
  queue_as :default

  def perform
    # This job will be called by the background service to collect data
    # For now, it creates a sample reading for today if one doesn't exist
    
    today = Date.current
    
    reading = VirtualBatteryReading.find_or_initialize_by(date: today)
    
    unless reading.persisted?
      # In a real implementation, these values would be fetched from the SSD provider's API
      # For demonstration purposes, we'll use sample data
      reading.assign_attributes(
        current_charge: rand(0.0..100.0).round(2),
        exported_to_battery: rand(0.0..50.0).round(2),
        imported_from_battery: rand(0.0..50.0).round(2),
        imported_from_grid: rand(0.0..30.0).round(2)
      )
      
      reading.save!
      Rails.logger.info "Created VirtualBatteryReading for #{today}: #{reading.inspect}"
    else
      Rails.logger.info "VirtualBatteryReading for #{today} already exists"
    end
  rescue StandardError => e
    Rails.logger.error "Failed to collect virtual battery data: #{e.message}"
    raise
  end
end
