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

  # Create or update a reading from profile data for a specific date
  # Returns the reading and updates current_charge in place
  def self.create_from_profile_data(date:, profile_data:, current_charge:)
    # Sum up all 15-minute intervals for the day
    # incoming = consumption from grid/battery (actualConsumption)
    # outgoing = production/export to battery (actualSupply)
    total_incoming = profile_data.sum { |row| row[:incoming].to_f / 4 } # convert from 15-min to hourly
    total_outgoing = profile_data.sum { |row| row[:outgoing].to_f / 4 } # convert from 15-min to hourly

    # Calculate battery transactions:
    # - outgoing goes INTO the virtual battery (exported_to_battery)
    # - incoming comes FROM the virtual battery first, then grid if depleted
    exported_to_battery = total_outgoing
    current_charge += exported_to_battery

    # Try to satisfy incoming demand from battery first
    imported_from_battery = 0.0
    imported_from_grid = 0.0

    if total_incoming <= current_charge
      # Battery can cover all consumption
      imported_from_battery = total_incoming
      current_charge -= total_incoming
    else
      # Battery partially covers consumption, rest from grid
      imported_from_battery = current_charge
      imported_from_grid = total_incoming - current_charge
      current_charge = 0.0
    end

    # Create or update the reading for this day
    reading = find_or_initialize_by(date: date)
    reading.assign_attributes(
      current_charge: current_charge.round(2),
      exported_to_battery: exported_to_battery.round(2),
      imported_from_battery: imported_from_battery.round(2),
      imported_from_grid: imported_from_grid.round(2)
    )
    reading.save!

    { reading: reading, current_charge: current_charge }
  end
end
