class VirtualBatteryReading < ApplicationRecord
  VIRTUAL_BATTERY_MAX_CAPACITY = 6000
  SYNC_LOCK_KEY = 74_201_001

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

  # Returns daily export and import values from 1 year ago to end of previous year (for projection)
  def self.previous_year_daily_data
    previous_year = Date.current.year - 1
    where(date: (Date.current - 1.year)..Date.new(previous_year).end_of_year)
      .order(:date)
      .map { |r| { date: r.date.iso8601, exported_to_grid: r.exported_to_grid, imported_from_grid: r.imported_from_grid } }
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

  def self.up_to_date?
    where(date: Date.yesterday..).exists?
  end

  # Fetches readings from SSD for every day since the last reading up to yesterday.
  # Returns the number of readings saved.
  def self.sync_missing_readings!
    return 0 if up_to_date?

    with_sync_lock do
      last_date = maximum(:date)
      start_date = last_date ? last_date + 1 : Date.current.beginning_of_year
      saved = 0
      next saved if start_date > Date.yesterday

      ssd_client = SsdApiClient.new
      (start_date..Date.yesterday).each do |date|
        profile_data = ssd_client.fetch_profile_data_for_date(date)
        # SSD has not published this day yet; retry on a later visit.
        break if profile_data.blank?

        reading = create_from_profile_data(date: date, profile_data: profile_data)[:reading]
        Rails.logger.info "Saved reading for #{date}: exported_grid=#{reading.exported_to_grid}, " \
                          "imported_grid=#{reading.imported_from_grid}"
        saved += 1
      end
      saved
    end
  end

  # Skips the block (returns 0) if another process is already syncing.
  def self.with_sync_lock
    # Advisory locks are server-wide, so scope the key to this database.
    lock_args = "#{SYNC_LOCK_KEY}, hashtext(current_database())"
    return 0 unless connection.select_value("SELECT pg_try_advisory_lock(#{lock_args})")

    begin
      yield
    ensure
      connection.select_value("SELECT pg_advisory_unlock(#{lock_args})")
    end
  end
  private_class_method :with_sync_lock
end
