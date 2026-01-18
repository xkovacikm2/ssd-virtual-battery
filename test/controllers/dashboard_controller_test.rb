require "test_helper"

class DashboardControllerTest < ActionDispatch::IntegrationTest
  test "should get index" do
    get root_url
    assert_response :success
    assert_select "h1", "Virtual Battery Dashboard"
  end

  test "should display year to date summary" do
    VirtualBatteryReading.create!(
      date: Date.current,
      current_charge: 75.5,
      exported_to_battery: 20.0,
      imported_from_battery: 15.0,
      imported_from_grid: 10.0
    )
    
    get root_url
    assert_response :success
    assert_select "h2", "Year to Date Summary (#{Date.current.year})"
  end

  test "should handle no data gracefully" do
    VirtualBatteryReading.destroy_all
    
    get root_url
    assert_response :success
    assert_select ".no-data"
  end
end
