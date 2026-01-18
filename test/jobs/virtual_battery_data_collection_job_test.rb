require "test_helper"

class VirtualBatteryDataCollectionJobTest < ActiveJob::TestCase
  test "should create a reading for today if it doesn't exist" do
    VirtualBatteryReading.where(date: Date.current).destroy_all
    
    assert_difference "VirtualBatteryReading.count", 1 do
      VirtualBatteryDataCollectionJob.perform_now
    end
    
    reading = VirtualBatteryReading.find_by(date: Date.current)
    assert_not_nil reading
    assert reading.current_charge >= 0
    assert reading.exported_to_battery >= 0
    assert reading.imported_from_battery >= 0
    assert reading.imported_from_grid >= 0
  end

  test "should not create duplicate reading for today" do
    VirtualBatteryReading.create!(
      date: Date.current,
      current_charge: 50.0,
      exported_to_battery: 10.0,
      imported_from_battery: 8.0,
      imported_from_grid: 5.0
    )
    
    assert_no_difference "VirtualBatteryReading.count" do
      VirtualBatteryDataCollectionJob.perform_now
    end
  end
end
