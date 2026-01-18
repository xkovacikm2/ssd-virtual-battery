namespace :virtual_battery do
  desc "Collect virtual battery data"
  task collect_data: :environment do
    puts "Starting virtual battery data collection..."
    VirtualBatteryDataCollectionJob.perform_now
    puts "Virtual battery data collection completed."
  end

  desc "Seed sample data for testing (creates readings for the past 30 days)"
  task seed_sample_data: :environment do
    puts "Seeding sample virtual battery data..."
    
    30.downto(0) do |days_ago|
      date = Date.current - days_ago.days
      
      reading = VirtualBatteryReading.find_or_initialize_by(date: date)
      
      unless reading.persisted?
        reading.assign_attributes(
          current_charge: rand(50.0..100.0).round(2),
          exported_to_battery: rand(5.0..50.0).round(2),
          imported_from_battery: rand(5.0..50.0).round(2),
          imported_from_grid: rand(2.0..30.0).round(2)
        )
        
        reading.save!
        puts "Created reading for #{date}"
      else
        puts "Reading for #{date} already exists"
      end
    end
    
    puts "Sample data seeding completed."
  end
end
