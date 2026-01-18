require "test_helper"

class VirtualBatteryReadingTest < ActiveSupport::TestCase
  test "should be valid with all required attributes" do
    reading = VirtualBatteryReading.new(
      date: Date.current,
      current_charge: 50.0,
      exported_to_battery: 10.0,
      imported_from_battery: 8.0,
      imported_from_grid: 5.0
    )
    assert reading.valid?
  end

  test "should require date" do
    reading = VirtualBatteryReading.new
    assert_not reading.valid?
    assert_includes reading.errors[:date], "can't be blank"
  end

  test "should enforce unique date" do
    date = Date.current
    VirtualBatteryReading.create!(
      date: date,
      current_charge: 50.0,
      exported_to_battery: 10.0,
      imported_from_battery: 8.0,
      imported_from_grid: 5.0
    )
    
    duplicate = VirtualBatteryReading.new(
      date: date,
      current_charge: 60.0
    )
    assert_not duplicate.valid?
    assert_includes duplicate.errors[:date], "has already been taken"
  end

  test "should validate non-negative values" do
    reading = VirtualBatteryReading.new(
      date: Date.current,
      current_charge: -10.0
    )
    assert_not reading.valid?
    assert_includes reading.errors[:current_charge], "must be greater than or equal to 0"
  end

  test "year_to_date_summary should calculate correct totals" do
    # Create readings for current year
    3.times do |i|
      VirtualBatteryReading.create!(
        date: Date.current - i.days,
        current_charge: 50.0 + i,
        exported_to_battery: 10.0,
        imported_from_battery: 8.0,
        imported_from_grid: 5.0
      )
    end
    
    summary = VirtualBatteryReading.year_to_date_summary
    
    assert_equal 50.0, summary[:current_charge] # Latest reading
    assert_equal 30.0, summary[:total_exported_to_battery]
    assert_equal 24.0, summary[:total_imported_from_battery]
    assert_equal 15.0, summary[:total_imported_from_grid]
  end
end
