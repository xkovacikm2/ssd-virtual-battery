class DashboardController < ApplicationController
  ALLOWED_TABS = %w[daily weekly monthly].freeze

  def index
    @summary = VirtualBatteryReading.year_to_date_summary
    @chart_data = VirtualBatteryReading.daily_chart_data
    @previous_year_data = VirtualBatteryReading.previous_year_daily_data
    @has_readings = VirtualBatteryReading.exists?
    load_readings_data if @has_readings
  end

  def readings
    load_readings_data
    render layout: false
  end

  private

  def load_readings_data
    @tab = ALLOWED_TABS.include?(params[:tab]) ? params[:tab] : "daily"
    @page = [ params.fetch(:page, 0).to_i, 0 ].max

    case @tab
    when "daily" then load_daily_data
    when "weekly" then load_weekly_data
    when "monthly" then load_monthly_data
    end
  end

  def load_daily_data
    end_date = Date.current - (@page * 7)
    start_date = end_date - 6
    readings_by_date = VirtualBatteryReading
      .where(date: start_date..end_date)
      .index_by(&:date)

    @daily_readings = (start_date..end_date).reverse_each.map do |date|
      reading = readings_by_date[date]
      {
        date: date,
        exported: reading&.exported_to_grid || BigDecimal("0"),
        imported: reading&.imported_from_grid || BigDecimal("0")
      }
    end

    @has_newer = @page > 0
    @has_older = VirtualBatteryReading.where("date < ?", start_date).exists?
  end

  def load_weekly_data
    readings = VirtualBatteryReading.current_year.order(:date)
    @weekly_readings = readings.group_by { |r| r.date.cweek }.map do |_week, week_readings|
      exported = week_readings.sum(&:exported_to_grid)
      imported = week_readings.sum(&:imported_from_grid)
      {
        start_date: week_readings.first.date,
        end_date: week_readings.last.date,
        exported: exported,
        imported: imported
      }
    end.reverse
  end

  def load_monthly_data
    readings = VirtualBatteryReading.current_year.order(:date)
    @monthly_readings = readings.group_by { |r| r.date.month }.map do |month, month_readings|
      exported = month_readings.sum(&:exported_to_grid)
      imported = month_readings.sum(&:imported_from_grid)
      {
        month: month,
        exported: exported,
        imported: imported
      }
    end.reverse
  end
end
