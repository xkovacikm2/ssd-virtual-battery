class DashboardController < ApplicationController
  def index
    @summary = VirtualBatteryReading.year_to_date_summary
    @latest_reading = VirtualBatteryReading.order(date: :desc).first
  end
end
