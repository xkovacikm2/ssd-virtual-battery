require "test_helper"

class DashboardControllerTest < ActionDispatch::IntegrationTest
  test "should get index" do
    get root_url
    assert_response :success
    assert_select "h1", "Prehľad virtuálnej batérie"
  end

  test "should display year to date summary" do
    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 20.0,
      imported_from_grid: 10.0
    )

    get root_url
    assert_response :success
    assert_select "h2", "Ročný súhrn (#{Date.current.year})"
    assert_select ".metric-label", text: "Kumulatívne odoslané do siete (aktuálny rok)"
    assert_select ".metric-label", text: "Kumulatívne prijaté zo siete (aktuálny rok)"
    assert_select ".metric-label", text: "Kumulatívne saldo siete (aktuálny rok)"
    assert_select ".metric-value", text: /20\.00/
    assert_select ".metric-value", text: /10\.00/
  end

  test "should handle no data gracefully" do
    VirtualBatteryReading.destroy_all

    get root_url
    assert_response :success
    assert_select ".no-data"
  end
end
