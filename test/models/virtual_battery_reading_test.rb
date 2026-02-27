require "test_helper"

class VirtualBatteryReadingTest < ActiveSupport::TestCase
  test "should be valid with all required attributes" do
    reading = VirtualBatteryReading.new(
      date: Date.current,
      exported_to_grid: 10.0,
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
      exported_to_grid: 10.0,
      imported_from_grid: 5.0
    )

    duplicate = VirtualBatteryReading.new(
      date: date
    )
    assert_not duplicate.valid?
    assert_includes duplicate.errors[:date], "has already been taken"
  end

  test "create_from_profile_data maps incoming to imported_from_grid" do
    profile_data = [
      { incoming: 10.0, outgoing: 0.0 },
      { incoming: 10.0, outgoing: 0.0 },
      { incoming: 10.0, outgoing: 0.0 },
      { incoming: 10.0, outgoing: 0.0 }
    ]

    result = VirtualBatteryReading.create_from_profile_data(
      date: Date.current,
      profile_data: profile_data
    )

    reading = result[:reading]
    assert_equal(0.0, reading.exported_to_grid)
    assert_equal(10.0, reading.imported_from_grid)
  end

  test "create_from_profile_data maps outgoing to exported_to_grid" do
    profile_data = [
      { incoming: 0.0, outgoing: 40.0 },
      { incoming: 0.0, outgoing: 40.0 },
      { incoming: 0.0, outgoing: 40.0 },
      { incoming: 0.0, outgoing: 40.0 }
    ]

    result = VirtualBatteryReading.create_from_profile_data(
      date: Date.current,
      profile_data: profile_data
    )

    reading = result[:reading]
    assert_equal(40.0, reading.exported_to_grid)
    assert_equal(0.0, reading.imported_from_grid)
  end

  test "year_to_date_summary should calculate correct totals" do
    VirtualBatteryReading.create!(
      date: Date.current - 2.days,
      exported_to_grid: 10.0,
      imported_from_grid: 5.0
    )

    VirtualBatteryReading.create!(
      date: Date.current - 1.day,
      exported_to_grid: 10.0,
      imported_from_grid: 5.0
    )

    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 10.0,
      imported_from_grid: 5.0
    )

    summary = VirtualBatteryReading.year_to_date_summary

    assert_equal 30.0, summary[:total_exported_to_grid]
    assert_equal 15.0, summary[:total_imported_from_grid]
    assert_equal 15.0, summary[:net_grid_flow]
  end

  test "year_to_date_summary should cap export at 6000 for net_grid_flow" do
    VirtualBatteryReading.create!(
      date: Date.current - 1.day,
      exported_to_grid: 3500.0,
      imported_from_grid: 400.0
    )

    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 3200.0,
      imported_from_grid: 700.0
    )

    summary = VirtualBatteryReading.year_to_date_summary

    assert_equal 6700.0, summary[:total_exported_to_grid]
    assert_equal 1100.0, summary[:total_imported_from_grid]
    assert_equal 4900.0, summary[:net_grid_flow]
  end
end
