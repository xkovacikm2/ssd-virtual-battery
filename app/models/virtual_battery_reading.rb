class VirtualBatteryReading < ApplicationRecord
  VIRTUAL_BATTERY_MAX_CAPACITY = 6000

  validates :date, presence: true, uniqueness: true
  validates :exported_to_grid, :imported_from_grid,
            numericality: { greater_than_or_equal_to: 0 }

  # Scope to get readings for current calendar year
  scope :current_year, -> { where(date: Date.current.beginning_of_year..Date.current.end_of_year) }

  # Calculate cumulative sums for the current year
  def self.year_to_date_summary
    readings = current_year.order(:date)
    total_exported_to_grid = readings.sum(:exported_to_grid)
    total_imported_from_grid = readings.sum(:imported_from_grid)
    net_grid_flow = [ VIRTUAL_BATTERY_MAX_CAPACITY, total_exported_to_grid ].min - total_imported_from_grid

    {
      total_exported_to_grid: total_exported_to_grid,
      total_imported_from_grid: total_imported_from_grid,
      net_grid_flow: net_grid_flow
    }
  end

  # Returns daily cumulative export and import for the current year (for charting)
  def self.daily_chart_data
    readings = current_year.order(:date)
    cumulative_exported = 0
    cumulative_imported = 0
    readings.map do |r|
      cumulative_exported += r.exported_to_grid
      cumulative_imported += r.imported_from_grid
      {
        date: r.date.iso8601,
        cumulative_exported: cumulative_exported.round(2),
        cumulative_imported: cumulative_imported.round(2)
      }
    end
  end

  # Create or update a reading from profile data for a specific date
  # Returns the reading
  def self.create_from_profile_data(date:, profile_data:)
    # Sum up all 15-minute intervals for the day
    # incoming = imported from public grid (actualConsumption)
    # outgoing = exported to public grid (actualSupply)
    total_incoming = profile_data.sum { |row| row[:incoming].to_f / 4 } # convert from 15-min to hourly
    total_outgoing = profile_data.sum { |row| row[:outgoing].to_f / 4 } # convert from 15-min to hourly

    # Daily grid transactions:
    # - all outgoing is exported to grid
    # - all incoming is imported from grid
    exported_to_grid = total_outgoing
    imported_from_grid = total_incoming

    # Create or update the reading for this day
    reading = find_or_initialize_by(date: date)
    reading.assign_attributes(
      exported_to_grid: exported_to_grid.round(2),
      imported_from_grid: imported_from_grid.round(2)
    )
    reading.save!

    { reading: reading }
  end
end
