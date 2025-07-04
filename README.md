# leaderboard-service

A Go-based leaderboard service with a Postgres backend, orchestrated via Docker Compose. The project is idiomatically structured and follows clean architecture principles.

---

## Features

- **Player CRUD:** Create, read, and update player profiles.
- **Matchmaking:** Players join a waiting queue; a background worker groups them into competitions of 10, matching by player level (optionally extensible to country).
- **Competition Management:** Only one active competition per player at a time. Competitions have statuses: ACTIVE, COMPLETED, CANCELLED.
- **Score Submission:** Players submit scores during an active competition; scores are incrementally added.
- **Leaderboard Retrieval:** Retrieve leaderboard standings for a player's current/past competition or by competition ID.
- **Concurrency:** Race-free matchmaking and score updates, with context propagation and graceful shutdown.
- **Logging:** Comprehensive logging and robust error handling at all layers.
- **Configuration:** Matchmaking interval and competition duration are configurable via environment variables.
- **Testing:** Full unit test coverage for repository, service, and handler layers. CI pipeline with Dockerized Postgres.
- **Graceful Shutdown:** Clean exit for HTTP server and background workers.
- **(Bonus-ready):** Easily extensible for country-aware grouping and Prometheus metrics.

---

## Project Structure

- `cmd/server/` — Main entrypoint
- `internal/api/` — HTTP handlers and routing
- `internal/service/` — Business logic and matchmaking worker
- `internal/repository/` — Database access and queries
- `internal/model/` — Data models and enums
- `internal/db/` — Database connection helpers
- `initdb/schema.sql` — Database schema (applied at container startup)

---

## Configuration

Set the following environment variables (defaults in parentheses):

- `MATCHMAKING_INTERVAL` (`30s`)
- `COMPETITION_DURATION` (`1h`)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (for Postgres)

---

## API Endpoints

- `POST /player` — Create player
- `GET /player/{player_id}` — Get player
- `PUT /player/{player_id}` — Update player
- `POST /leaderboard/join?player_id={id}` — Join matchmaking queue (202 Accepted if waiting, 409 Conflict if already in competition)
- `POST /leaderboard/score` — Submit score (200 OK on success, 409/404 on error)
- `GET /leaderboard/player/{player_id}` — Get player's current or last competition leaderboard
- `GET /leaderboard/{leaderboardID}` — Get leaderboard by competition ID

**All endpoints return appropriate HTTP status codes and error messages.**

---

## Error Handling

- Returns 404 for not found, 409 for conflicts, 400 for bad requests, 500 for server errors.
- Prevents duplicate players in the waiting queue and multiple active competitions per player.
- Returns 404 if submitting a score for a non-existent player or competition.

---

## Logging

- Logs all key actions and errors at service and repository layers.

---

## Testing

- **Repository, service, and handler layers**: Full unit test coverage, including edge and error cases.
- **CI/CD**: GitHub Actions workflow runs all tests with Dockerized Postgres and schema migration.
- **(Optional)**: Add end-to-end/integration tests for extra coverage.

---

## Running

1. **Docker Compose** (recommended):
   ```sh
   docker-compose up --build
   ```
   The service and Postgres will start, and the schema will be applied automatically.

2. **Non-Docker**:
   - Set up a local Postgres instance and apply `initdb/schema.sql`.
   - Set environment variables as needed.
   - Run:
     ```sh
     go run ./cmd/server
     ```

---

## Design Decisions & Trade-offs

- **Clean architecture**: Separation of concerns for maintainability and testability.
- **Concurrency**: Matchmaking and score updates are race-free and context-aware.
- **Extensibility**: Country-aware grouping and Prometheus metrics can be added with minimal changes.
- **Testing**: Focused on unit tests for all layers; integration tests are recommended for further robustness.

---

## Recent Improvements

- Full test coverage for all layers and edge cases.
- CI pipeline with Dockerized Postgres and schema migration.
- Improved error handling, logging, and API response consistency.
- Graceful shutdown and context propagation.
- Configurable matchmaking and competition durations.
- Cleaned up and documented API responses.

---

**For details on API usage and error codes, see the handler tests or contact the maintainer.**