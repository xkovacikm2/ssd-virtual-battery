# SSD Virtual Battery Dashboard

Datamine and visualise information about status of charge of your virtual battery from SSD provider.

## Features

- **Dashboard**: View current charge status and cumulative statistics for the current calendar year
- **Metrics Tracked**:
  - Current charge in kWh
  - Total electricity exported to virtual battery
  - Total electricity imported from virtual battery
  - Total electricity imported from public grid
- **Background Service**: Automated data collection job
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

| Column                    | Type    | Description                                      |
|---------------------------|---------|--------------------------------------------------|
| date                      | date    | Date of the reading (unique)                     |
| current_charge            | decimal | Current charge status in kWh                     |
| exported_to_battery       | decimal | Daily electricity exported to virtual battery    |
| imported_from_battery     | decimal | Daily electricity imported from virtual battery  |
| imported_from_grid        | decimal | Daily electricity imported from public grid      |

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

- **Ruby**: 3.2.3
- **Rails**: 8.1.2
- **Database**: PostgreSQL 16
- **Background Jobs**: Active Job with Solid Queue
- **Development**: DevContainer with Docker Compose

## Future Enhancements

- Integration with actual SSD provider API
- Historical data charts and graphs
- Export data to CSV/PDF
- Alert system for charge levels
- Multi-user support with authentication
- Mobile-responsive improvements
