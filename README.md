# leaderboard-service

A Go-based leaderboard service with a Postgres backend, orchestrated via Docker Compose. The project is structured idiomatically and follows clean architecture principles.

## Features
- Player CRUD (create, read, update)
- Matchmaking queue and worker (groups waiting players into competitions)
- Score submission (with validation and error handling)
- Leaderboard retrieval (per competition)
- Only one active leaderboard (competition) at a time
- Competition status management (ACTIVE, COMPLETED, CANCELLED)
- Concurrency-safe matchmaking worker with context cancellation
- Robust error handling and comprehensive logging at all layers
- Configurable matchmaking interval and competition duration via environment variables
- Graceful shutdown and connection pool management

## Project Structure
- `cmd/server/` - Main entrypoint
- `internal/api/` - HTTP handlers and routing
- `internal/service/` - Business logic and matchmaking worker
- `internal/repository/` - Database access and queries
- `internal/model/` - Data models and enums
- `internal/db/` - Database connection helpers
- `initdb/schema.sql` - Database schema

## Configuration
Set the following environment variables (with defaults):
- `MATCHMAKING_INTERVAL` (default: `30s`)
- `COMPETITION_DURATION` (default: `1m`)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (for Postgres)

## API Endpoints
- `POST /player` - Create player
- `GET /player/{player_id}` - Get player
- `PUT /player/{player_id}` - Update player
- `POST /leaderboard/join` - Join matchmaking queue
- `POST /leaderboard/score` - Submit score (200 OK, no body on success)
- `GET /leaderboard/player/{player_id}` - Get player's leaderboard
- `GET /leaderboard/{leaderboardID}` - Get leaderboard by competition

## Error Handling
- Returns appropriate HTTP status codes (404 for not found, 409 for conflicts, 400 for bad requests, 500 for server errors)
- Prevents duplicate players in the waiting queue
- Prevents multiple active leaderboards
- Returns 404 if submitting a score for a non-existent player

## Logging
- Logs all key actions and errors at both service and repository layers

## Testing
- Unit tests for repository CRUD operations
- (Recommended) Add integration tests for service and API layers

## Running
Use Docker Compose to start the service and Postgres database. Ensure environment variables are set as needed.

---

**Recent Improvements:**
- Removed all dummy methods from the repository
- Added player existence check to score submission
- Improved error handling and logging
- Prevented duplicate entries in the matchmaking queue
- Enforced single active competition at a time
- Made matchmaking interval and competition duration configurable
- Cleaned up and documented API responses