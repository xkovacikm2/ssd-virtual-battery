class DashboardController < ApplicationController
  def index
    @summary = VirtualBatteryReading.year_to_date_summary
    @latest_reading = VirtualBatteryReading.order(date: :desc).first
    @chart_data = VirtualBatteryReading.daily_chart_data
    @previous_year_data = VirtualBatteryReading.previous_year_daily_data
  end
end
