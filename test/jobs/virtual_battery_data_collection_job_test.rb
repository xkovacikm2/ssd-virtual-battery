require "test_helper"

class VirtualBatteryDataCollectionJobTest < ActiveJob::TestCase
  FAKE_PROFILE_DATA = [
    { incoming: "3.0", outgoing: "5.0" },
    { incoming: "2.0", outgoing: "4.0" }
  ].freeze

  FakeApiClient = Struct.new(:profile_data) do
    def fetch_profile_data_for_date = profile_data
  end

  setup do
    fake = FakeApiClient.new(FAKE_PROFILE_DATA)
    SsdApiClient.define_singleton_method(:new) { fake }
  end

  teardown do
    SsdApiClient.singleton_class.remove_method(:new)
  end

  test "should create a reading for yesterday if it doesn't exist" do
    VirtualBatteryReading.where(date: Date.yesterday).destroy_all

    assert_difference "VirtualBatteryReading.count", 1 do
      VirtualBatteryDataCollectionJob.perform_now
    end

    reading = VirtualBatteryReading.find_by(date: Date.yesterday)
    assert_not_nil reading
    assert reading.exported_to_grid >= 0
    assert reading.imported_from_grid >= 0
  end

  test "should not create duplicate reading for yesterday" do
    VirtualBatteryReading.where(date: Date.yesterday).destroy_all
    VirtualBatteryReading.create!(
      date: Date.yesterday,
      exported_to_grid: 10.0,
      imported_from_grid: 5.0
    )

    assert_no_difference "VirtualBatteryReading.count" do
      VirtualBatteryDataCollectionJob.perform_now
    end
  end
end
