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

  test "should display daily readings table on index" do
    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 15.0,
      imported_from_grid: 5.0
    )

    get root_url
    assert_response :success
    assert_select ".readings-table"
    assert_select ".readings-tab--active", text: "Denné"
  end

  test "should get readings with daily tab" do
    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 15.0,
      imported_from_grid: 5.0
    )

    get dashboard_readings_url(tab: "daily")
    assert_response :success
    assert_select "turbo-frame#readings"
    assert_select ".readings-table"
  end

  test "should get readings with weekly tab" do
    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 15.0,
      imported_from_grid: 5.0
    )

    get dashboard_readings_url(tab: "weekly")
    assert_response :success
    assert_select ".readings-table"
    assert_select ".readings-tab--active", text: "Týždenné"
  end

  test "should get readings with monthly tab" do
    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 15.0,
      imported_from_grid: 5.0
    )

    get dashboard_readings_url(tab: "monthly")
    assert_response :success
    assert_select ".readings-table"
    assert_select ".readings-tab--active", text: "Mesačné"
  end

  test "should paginate daily readings" do
    VirtualBatteryReading.create!(
      date: Date.current - 10,
      exported_to_grid: 12.0,
      imported_from_grid: 4.0
    )

    get dashboard_readings_url(tab: "daily", page: 1)
    assert_response :success
    assert_select ".pagination-link", text: "Novšie →"
  end

  test "should default to daily tab for invalid tab param" do
    VirtualBatteryReading.create!(
      date: Date.current,
      exported_to_grid: 10.0,
      imported_from_grid: 3.0
    )

    get dashboard_readings_url(tab: "invalid")
    assert_response :success
    assert_select ".readings-tab--active", text: "Denné"
  end

  # --- Daily view rendering ---

  test "daily view renders all 4 table columns" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 18.5, imported_from_grid: 6.3)

    get dashboard_readings_url(tab: "daily")
    assert_select ".readings-table th", count: 4
    assert_select ".readings-table th", text: "Dátum"
    assert_select ".readings-table th", text: "Odoslané (kWh)"
    assert_select ".readings-table th", text: "Prijaté (kWh)"
    assert_select ".readings-table th", text: "Rozdiel (kWh)"
  end

  test "daily view shows 7 rows for a full week" do
    7.times do |i|
      VirtualBatteryReading.create!(date: Date.current - i, exported_to_grid: 10.0, imported_from_grid: 3.0)
    end

    get dashboard_readings_url(tab: "daily")
    assert_select ".readings-table tbody tr", count: 7
  end

  test "daily view shows correct values for a reading" do
    reading_date = Date.yesterday
    VirtualBatteryReading.create!(date: reading_date, exported_to_grid: 25.50, imported_from_grid: 8.75)

    get dashboard_readings_url(tab: "daily")
    assert_select ".readings-table tbody tr" do
      assert_select "td", text: /#{reading_date.strftime('%d. %m.')}/
      assert_select "td", text: /25\.50/
      assert_select "td", text: /8\.75/
      assert_select "td", text: /16\.75/
    end
  end

  test "daily view shows zero values for dates without readings" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 10.0, imported_from_grid: 5.0)

    get dashboard_readings_url(tab: "daily")
    # 7 rows total but only 1 has data; the rest show 0.00
    assert_select ".readings-table tbody tr", count: 7
    assert_select ".readings-table tbody td", text: "0.00", minimum: 3
  end

  test "daily view shows day of week in Slovak" do
    # Create a reading on a known day
    monday = Date.current.beginning_of_week(:monday)
    VirtualBatteryReading.create!(date: monday, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "daily", page: ((Date.current - monday).to_i / 7))
    assert_select ".readings-table tbody td", text: /Pondelok/
  end

  test "daily view shows no older link on first page with no older data" do
    VirtualBatteryReading.destroy_all
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "daily", page: 0)
    assert_select ".pagination-link", text: "← Staršie", count: 0
  end

  test "daily view shows older and newer pagination links" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 5.0, imported_from_grid: 2.0)
    VirtualBatteryReading.create!(date: Date.current - 14, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "daily", page: 1)
    assert_select ".pagination-link", text: "← Staršie"
    assert_select ".pagination-link", text: "Novšie →"
  end

  # --- Weekly view rendering ---

  test "weekly view renders 4 table columns" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 10.0, imported_from_grid: 4.0)

    get dashboard_readings_url(tab: "weekly")
    assert_select ".readings-table th", count: 4
    assert_select ".readings-table th", text: "Týždeň"
    assert_select ".readings-table th", text: "Odoslané (kWh)"
    assert_select ".readings-table th", text: "Prijaté (kWh)"
    assert_select ".readings-table th", text: "Rozdiel (kWh)"
  end

  test "weekly view aggregates readings by week" do
    monday = Date.current.beginning_of_week(:monday)
    VirtualBatteryReading.create!(date: monday, exported_to_grid: 10.0, imported_from_grid: 3.0)
    VirtualBatteryReading.create!(date: monday + 1, exported_to_grid: 15.0, imported_from_grid: 5.0)

    get dashboard_readings_url(tab: "weekly")
    # Sums: exported 25.00, imported 8.00, diff 17.00
    assert_select ".readings-table tbody td", text: /25\.00/
    assert_select ".readings-table tbody td", text: /8\.00/
    assert_select ".readings-table tbody td", text: /17\.00/
  end

  test "weekly view shows date range for each week" do
    monday = Date.current.beginning_of_week(:monday)
    VirtualBatteryReading.create!(date: monday, exported_to_grid: 5.0, imported_from_grid: 2.0)
    VirtualBatteryReading.create!(date: monday + 3, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "weekly")
    assert_select ".readings-table tbody td", text: /#{monday.strftime('%d. %m.')}.*#{(monday + 3).strftime('%d. %m.')}/
  end

  test "weekly view has no pagination" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "weekly")
    assert_select ".readings-pagination", count: 0
  end

  # --- Monthly view rendering ---

  test "monthly view renders 4 table columns" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 10.0, imported_from_grid: 4.0)

    get dashboard_readings_url(tab: "monthly")
    assert_select ".readings-table th", count: 4
    assert_select ".readings-table th", text: "Mesiac"
    assert_select ".readings-table th", text: "Odoslané (kWh)"
    assert_select ".readings-table th", text: "Prijaté (kWh)"
    assert_select ".readings-table th", text: "Rozdiel (kWh)"
  end

  test "monthly view shows Slovak month names" do
    VirtualBatteryReading.create!(date: Date.new(Date.current.year, 1, 15), exported_to_grid: 10.0, imported_from_grid: 4.0)
    VirtualBatteryReading.create!(date: Date.new(Date.current.year, 3, 10), exported_to_grid: 8.0, imported_from_grid: 3.0)

    get dashboard_readings_url(tab: "monthly")
    assert_select ".readings-table tbody td", text: "Január"
    assert_select ".readings-table tbody td", text: "Marec"
  end

  test "monthly view aggregates readings by month" do
    VirtualBatteryReading.create!(date: Date.new(Date.current.year, 2, 5), exported_to_grid: 10.0, imported_from_grid: 3.0)
    VirtualBatteryReading.create!(date: Date.new(Date.current.year, 2, 20), exported_to_grid: 12.0, imported_from_grid: 4.0)

    get dashboard_readings_url(tab: "monthly")
    # Sums: exported 22.00, imported 7.00, diff 15.00
    assert_select ".readings-table tbody td", text: /22\.00/
    assert_select ".readings-table tbody td", text: /7\.00/
    assert_select ".readings-table tbody td", text: /15\.00/
  end

  test "monthly view has no pagination" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "monthly")
    assert_select ".readings-pagination", count: 0
  end

  # --- Tabs rendering ---

  test "all three tab links are rendered" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 5.0, imported_from_grid: 2.0)

    get dashboard_readings_url(tab: "daily")
    assert_select ".readings-tab", count: 3
    assert_select ".readings-tab", text: "Denné"
    assert_select ".readings-tab", text: "Týždenné"
    assert_select ".readings-tab", text: "Mesačné"
  end

  test "only one tab is active at a time" do
    VirtualBatteryReading.create!(date: Date.current, exported_to_grid: 5.0, imported_from_grid: 2.0)

    %w[daily weekly monthly].each do |tab|
      get dashboard_readings_url(tab: tab)
      assert_select ".readings-tab--active", count: 1
    end
  end
end
