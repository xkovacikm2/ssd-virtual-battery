class VirtualBatteryReading < ApplicationRecord
  validates :date, presence: true, uniqueness: true
  validates :current_charge, :exported_to_battery, :imported_from_battery, :imported_from_grid,
            numericality: { greater_than_or_equal_to: 0 }

  # Scope to get readings for current calendar year
  scope :current_year, -> { where(date: Date.current.beginning_of_year..Date.current.end_of_year) }

  # Calculate cumulative sums for the current year
  def self.year_to_date_summary
    readings = current_year.order(:date)
    {
      current_charge: readings.last&.current_charge || 0,
      total_exported_to_battery: readings.sum(:exported_to_battery),
      total_imported_from_battery: readings.sum(:imported_from_battery),
      total_imported_from_grid: readings.sum(:imported_from_grid)
    }
  end
end
