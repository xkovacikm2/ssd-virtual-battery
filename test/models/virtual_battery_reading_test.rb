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

  test "daily_chart_data returns empty array when no readings exist" do
    VirtualBatteryReading.current_year.destroy_all

    assert_equal [], VirtualBatteryReading.daily_chart_data
  end

  test "daily_chart_data returns one entry per reading in date order" do
    VirtualBatteryReading.current_year.destroy_all
    VirtualBatteryReading.create!(date: Date.current - 2.days, exported_to_grid: 10.0, imported_from_grid: 4.0)
    VirtualBatteryReading.create!(date: Date.current - 1.day,  exported_to_grid: 20.0, imported_from_grid: 6.0)
    VirtualBatteryReading.create!(date: Date.current,          exported_to_grid: 30.0, imported_from_grid: 8.0)

    data = VirtualBatteryReading.daily_chart_data

    assert_equal 3, data.length
    assert_equal (Date.current - 2.days).iso8601, data[0][:date]
    assert_equal (Date.current - 1.day).iso8601,  data[1][:date]
    assert_equal Date.current.iso8601,             data[2][:date]
  end

  test "daily_chart_data accumulates cumulative_exported and cumulative_imported" do
    VirtualBatteryReading.current_year.destroy_all
    VirtualBatteryReading.create!(date: Date.current - 2.days, exported_to_grid: 10.0, imported_from_grid: 4.0)
    VirtualBatteryReading.create!(date: Date.current - 1.day,  exported_to_grid: 20.0, imported_from_grid: 6.0)
    VirtualBatteryReading.create!(date: Date.current,          exported_to_grid: 30.0, imported_from_grid: 8.0)

    data = VirtualBatteryReading.daily_chart_data

    assert_equal 10.0, data[0][:cumulative_exported]
    assert_equal 4.0,  data[0][:cumulative_imported]

    assert_equal 30.0, data[1][:cumulative_exported]
    assert_equal 10.0, data[1][:cumulative_imported]

    assert_equal 60.0, data[2][:cumulative_exported]
    assert_equal 18.0, data[2][:cumulative_imported]
  end

  test "daily_chart_data excludes readings outside current year" do
    VirtualBatteryReading.current_year.destroy_all
    VirtualBatteryReading.create!(date: Date.new(Date.current.year - 1, 12, 31), exported_to_grid: 999.0, imported_from_grid: 999.0)
    VirtualBatteryReading.create!(date: Date.current - 1.day, exported_to_grid: 10.0, imported_from_grid: 5.0)

    data = VirtualBatteryReading.daily_chart_data

    assert_equal 1, data.length
    assert_equal 10.0, data[0][:cumulative_exported]
    assert_equal 5.0,  data[0][:cumulative_imported]
  end

  test "previous_year_daily_data returns empty array when no previous year readings exist" do
    destroy_previous_year_readings

    assert_equal [], VirtualBatteryReading.previous_year_daily_data
  end

  test "previous_year_daily_data returns readings from 1 year ago onwards, not earlier" do
    destroy_previous_year_readings

    previous_year = Date.current.year - 1
    before_cutoff = Date.current - 1.year - 1.day
    at_cutoff     = Date.current - 1.year

    # Only create the record at or after the cutoff when the cutoff is in the previous year
    if before_cutoff.year == previous_year
      VirtualBatteryReading.create!(date: before_cutoff, exported_to_grid: 5.0, imported_from_grid: 2.0)
    end
    VirtualBatteryReading.create!(date: at_cutoff, exported_to_grid: 15.0, imported_from_grid: 7.0) if at_cutoff.year == previous_year
    VirtualBatteryReading.create!(date: Date.new(previous_year, 12, 31), exported_to_grid: 20.0, imported_from_grid: 8.0)

    data = VirtualBatteryReading.previous_year_daily_data

    data.each do |row|
      assert row[:date] >= at_cutoff.iso8601, "Expected no readings before #{at_cutoff}, got #{row[:date]}"
    end
    assert data.any? { |row| row[:date] == Date.new(previous_year, 12, 31).iso8601 }
  end

  test "previous_year_daily_data excludes current year readings" do
    destroy_previous_year_readings

    previous_year = Date.current.year - 1
    VirtualBatteryReading.create!(date: Date.new(previous_year, 12, 31), exported_to_grid: 20.0, imported_from_grid: 8.0)
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 99.0, imported_from_grid: 99.0) unless VirtualBatteryReading.exists?(date: Date.current)

    data = VirtualBatteryReading.previous_year_daily_data

    assert data.none? { |row| row[:date].start_with?(Date.current.year.to_s) }
  end

  private

  def destroy_previous_year_readings
    previous_year = Date.current.year - 1
    VirtualBatteryReading.where(date: Date.new(previous_year).beginning_of_year..Date.new(previous_year).end_of_year).destroy_all
  end
end
