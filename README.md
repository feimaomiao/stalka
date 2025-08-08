# Stalka

A backend microservice for esports calendar (implementation pending) that periodically fetches data from the PandaScore API and stores it in a PostgreSQL database.

## Overview

Stalka is a Go-based data aggregation service designed to collect esports tournament, match, team, and league information from the PandaScore API. It provides a reliable way to maintain an up-to-date database of esports events and related data.

## Features

- **Periodic Data Fetching**: Automatically pulls fresh data from PandaScore API on configurable intervals
- **Safe Type Conversions**: Ensures that data is valid and safe before use
- **Comprehensive Data Coverage**: Fetches games, leagues, series, tournaments, matches, and teams
- **Database Integration**: Direct PostgreSQL integration with conflict resolution
- **Structured Logging**: Uses Zap for detailed logging and monitoring
- **Docker Support**: Containerized deployment with Docker Compose

## Architecture

### Core Components

- **PandaClient**: Main API client for interacting with PandaScore
- **Database Layer**: PostgreSQL connection and query management
- **Data Types**: Strongly typed structures for API responses and database rows
- **Safe Conversions**: Overflow-safe integer conversion utilities

### Data Flow

1. **Initial Setup**: Fetches comprehensive data across all categories
2. **Periodic Updates**:
   - Matches updated every hour
   - Full data refresh every 24 hours
3. **Dependency Resolution**: Automatically fetches missing related entities
4. **Database Storage**: Upserts data with conflict resolution

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- PandaScore API key

### Environment Variables

Create a `.env` file with the following variables:

```env
pandascore_secret=your_pandascore_api_key
writer_password=your_database_password
```

### Docker Deployment

1. Build and run with Docker Compose:

   ```bash
   docker-compose up -d
   ```

## Configuration

### Fetch Intervals

- **Matches**: Every hour (configurable via `matchTicker`)
- **Full Refresh**: Every 24 hours (configurable via `setupTicker`)

### Page Limits

- **Regular Updates**: 20 pages per entity type
- **Setup Mode**: 60 pages for comprehensive initial data

## API Data Sources

The service fetches data from the following PandaScore API endpoints:

- `/videogames` - Game information
- `/leagues` - League details and metadata
- `/series` - Tournament series data
- `/tournaments` - Individual tournament information
- `/matches/upcoming` - Future scheduled matches
- `/matches/past` - Historical match results
- `/teams` - Team profiles and statistics

## Database Schema

The service maintains the following core entities:

- **Games**: Video game information
- **Leagues**: Competition leagues
- **Series**: Tournament series
- **Tournaments**: Individual tournaments
- **Matches**: Individual matches with results
- **Teams**: Competing teams

## Error Handling

- **Dependency Resolution**: Automatically fetches missing related entities
- **API Rate Limiting**: Respects PandaScore API limits

## Logging

The service uses structured logging with different levels:

- **Info**: General operation status
- **Debug**: Detailed operation information
- **Error**: Error conditions and failures
- **Fatal**: Critical errors that terminate the service

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the terms specified in the LICENSE file.

## Support

For issues and questions, please open an issue.
