# SSD Virtual Battery Dashboard

[![Security Scan](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/scan.yml/badge.svg)](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/scan.yml)
[![Lint](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/lint.yml/badge.svg)](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/lint.yml)
[![Tests](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/test.yml/badge.svg)](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/test.yml)
[![Docker](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/docker.yml/badge.svg)](https://github.com/xkovacikm2/ssd-virtual-battery/actions/workflows/docker.yml)

Datamine and visualise information about status of charge of your virtual battery from SSD provider.

## Features

- **Dashboard**: View cumulative statistics for the current calendar year and the latest daily reading
- **Metrics Tracked**:
  - Total electricity exported to grid (year-to-date)
  - Total electricity imported from grid (year-to-date)
  - Net virtual battery balance (year-to-date)
- **Grid Flow Logic**: Daily outgoing energy is stored as `exported_to_grid`; daily incoming energy is stored as `imported_from_grid`. Virtual battery balance is calculated as `min(total_exported, 6000) - total_imported` (virtual battery capacity: 6,000 kWh).
- **Burndown Chart**: Visualises cumulative daily exports and imports for the current year alongside the previous year for comparison
- **Background Service**: Automated data collection job that fetches 15-minute interval data from the SSD provider and aggregates it into daily readings
- **PostgreSQL Database**: Stores daily readings with proper indexing

## Prerequisites

- Docker and Docker Compose
- VS Code with Remote - Containers extension (for devcontainer)

## Getting Started

### Using DevContainer (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/xkovacikm2/ssd-virtual-battery.git
   cd ssd-virtual-battery
   ```

2. Open in VS Code:
   ```bash
   code .
   ```

3. When prompted, click "Reopen in Container" or run the command "Remote-Containers: Reopen in Container"

4. The container will build and install all dependencies automatically

5. Once the container is ready, the database will be created and migrated
   
6. Copy .env.template as .env and provide values for variables. Name and password for your SSD provider at [ims.ssd.sk](https://ims.ssd.sk/login)
   and id of your extraction point.

### Manual Setup

1. Install dependencies:
   ```bash
   bundle install
   ```

2. Set up the database:
   ```bash
   rails db:create
   rails db:migrate
   ```

3. (Optional) Seed sample data for testing:
   ```bash
   rails virtual_battery:seed_sample_data
   ```

4. Start the server:
   ```bash
   rails server
   ```

5. Visit http://localhost:3000 to see the dashboard

## Background Data Collection

The application includes a background job for collecting virtual battery data. You can run it manually:

```bash
rails virtual_battery:collect_data
```

In a production environment, you would schedule this job to run periodically using:
- **Cron**: Add a cron job to run the rake task daily
- **Solid Queue**: Configure recurring jobs in Rails 8
- **Sidekiq**: Use sidekiq-cron or similar scheduling gem
- **Whenever gem**: For cron-like scheduling in Ruby

Example cron entry:
```
0 0 * * * cd /path/to/app && rails virtual_battery:collect_data
```

## Database Schema

### VirtualBatteryReading Model

**Stored columns** (one record per calendar date):

| Column                    | Type         | Description                                           |
|---------------------------|--------------|-------------------------------------------------------|
| date                      | date         | Date of the reading (unique, required)                |
| exported_to_grid          | decimal(10,2)| Daily electricity sent to the grid (kWh), default 0  |
| imported_from_grid        | decimal(10,2)| Daily electricity taken from the grid (kWh), default 0|

**Calculated values** (computed in memory, not stored):

| Value                     | Description                                                                 |
|---------------------------|-----------------------------------------------------------------------------|
| Year-to-date export total | Sum of `exported_to_grid` for all readings in the current year              |
| Year-to-date import total | Sum of `imported_from_grid` for all readings in the current year            |
| Net balance               | `min(total_exported, 6000) - total_imported` (virtual battery capacity: 6,000 kWh) |
| Daily difference          | `exported_to_grid − imported_from_grid` for the latest reading              |
| Cumulative chart data     | Running totals of exports and imports by date for the current year          |
| Previous year comparison  | Daily cumulative values from the previous year used as a projection baseline|

## Development

### Running Tests

```bash
rails test
```

### Running Linter

```bash
bin/rubocop
```

### Running Security Checks

```bash
bin/brakeman
bin/bundler-audit
```

## Technology Stack

- **Ruby**: 3.3
- **Rails**: 8.1.2
- **Database**: PostgreSQL 16
- **Background Jobs**: Active Job with Solid Queue
- **Development**: DevContainer with Docker Compose

## Future Enhancements

- Historical data charts and graphs
- Export data to CSV/PDF
- Alert system for charge levels
- Multi-user support with authentication
- Mobile-responsive improvements
