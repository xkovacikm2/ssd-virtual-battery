# Copilot Instructions

## Project Overview

SSD Virtual Battery Dashboard is a Ruby on Rails 8 application that collects and visualises virtual battery charge data from an SSD provider. It stores daily electricity readings (exported to grid, imported from grid) in a PostgreSQL database and displays cumulative statistics on a dashboard.

## Technology Stack

- **Language**: Ruby 3.3
- **Framework**: Rails 8.1.2
- **Database**: PostgreSQL 16
- **Background Jobs**: Active Job with Solid Queue
- **Asset Pipeline**: Propshaft with importmap-rails
- **Frontend**: Hotwire (Turbo + Stimulus)
- **Development Environment**: DevContainer with Docker Compose

## Setup & Development Commands

```bash
# Install dependencies
bundle install

# Set up the database
rails db:create db:migrate

# Seed sample data for development/testing
rails virtual_battery:seed_sample_data

# Start the development server
rails server

# Collect virtual battery data manually
rails virtual_battery:collect_data
```

## Testing

```bash
# Run all tests
rails test

# Run system tests
rails test:system
```

## Linting & Security

```bash
# Run RuboCop linter (Omakase Rails style)
bin/rubocop

# Run security scanner
bin/brakeman

# Audit gems for known vulnerabilities
bin/bundler-audit
```

## Code Conventions

- Follow the [Omakase Ruby styling for Rails](https://github.com/rails/rubocop-rails-omakase) enforced by RuboCop. Run `bin/rubocop -a` to auto-correct offenses before committing.
- Use Rails conventions: fat models, thin controllers, RESTful routes.
- All business logic for data collection lives under `app/jobs/` and `lib/tasks/virtual_battery.rake`.
- Decimal columns (e.g. energy values in kWh) should use `decimal` type with appropriate precision/scale in migrations.
- Every database migration must be reversible where possible.

## Domain Knowledge

- **VirtualBatteryReading**: the core model; one record per calendar date with `exported_to_grid` and `imported_from_grid` decimal columns.
- **Grid Flow Logic**: daily incoming energy is recorded as `imported_from_grid`; daily outgoing energy is recorded as `exported_to_grid`.
- SSD provider login credentials and extraction point ID are stored in `.env` (see `.env.template`). Never commit secrets.

## Environment Variables

Copy `.env.template` to `.env` and fill in:
- `SSD_USERNAME` – username for [ims.ssd.sk](https://ims.ssd.sk/login)
- `SSD_PASSWORD` – password for the SSD provider
- `SSD_EXTRACTION_POINT_ID` – ID of the extraction point

Never commit `.env` or any file containing secrets.

## Pull Request Guidelines

- Reference the related issue in the PR description.
- Ensure `rails test` passes before requesting a review.
- Ensure `bin/rubocop` reports no offenses.
- Ensure `bin/brakeman` reports no new warnings.
- Keep migrations and schema changes backwards-compatible where possible.
