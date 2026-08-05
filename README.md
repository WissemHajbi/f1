# oidysts F1 backend

Small Go foundation for validating F1 data providers, storing bounded snapshots in SQLite, and exposing source health to a future React Native client.

## Requirements

- Go 1.26+
- Docker (optional)

## Synchronize a complete database

Build once and populate season data plus one or more race sessions in dependency-safe order:

```powershell
powershell.exe -ExecutionPolicy Bypass -File scripts/sync-all.ps1 `
  -Season 2025 `
  -SessionKeys 9839
```

This synchronizes meetings, drivers, calendar, standings, results, classifications, laps, stints, pits, weather, race control, overtakes, positions, intervals, and locally cached team radio. It is idempotent and safe to rerun after interruption.

High-volume car data and physical location are opt-in and automatically divided into 15-minute chunks:

```powershell
powershell.exe -ExecutionPolicy Bypass -File scripts/sync-all.ps1 `
  -Season 2025 `
  -SessionKeys 9839 `
  -IncludeTelemetry
```

With `-IncludeTelemetry`, the script discovers all session drivers and session start/end times, then tolerates empty chunks after retirements. It accepts one session per run because of the data volume. You can override discovery with `-DriverNumbers`, `-TelemetryFrom`, and `-TelemetryTo`.

## Run source probes

```bash
go run ./cmd/probe -sources all -year 2025
```

Successful payloads are saved to `data/oidysts.db`. Select providers with `-sources openf1,jolpica,rss,fia,static`.

These are real external integration checks, so run them intentionally rather than in unit-test CI. Historical OpenF1 examples use a small fixed 2025 session/time window to prevent accidental large telemetry downloads.

## Sync normalized drivers

The sync command contacts OpenF1, selects the latest race that ended at least 30 minutes ago, and atomically stores its complete driver roster in SQLite:

```bash
go run ./cmd/sync -resource drivers -year 2025
go run ./cmd/sync -resource meetings -year 2025
go run ./cmd/sync -resource calendar -year 2025
go run ./cmd/sync -resource calendar -year 2026
go run ./cmd/sync -resource standings -year 2025
go run ./cmd/sync -resource results -year 2025
go run ./cmd/sync -resource classifications -year 2025
go run ./cmd/sync -resource laps -session 9839
go run ./cmd/sync -resource stints -session 9839
go run ./cmd/sync -resource pit -session 9839
go run ./cmd/sync -resource weather -session 9839
go run ./cmd/sync -resource race-control -session 9839
go run ./cmd/sync -resource overtakes -session 9839
go run ./cmd/sync -resource positions -session 9839
go run ./cmd/sync -resource intervals -session 9839
go run ./cmd/sync -resource location -session 9839 -driver 1 -from 2025-12-07T13:00:00Z -to 2025-12-07T13:15:00Z
go run ./cmd/sync -resource team-radio -session 9839 -driver 1
go run ./cmd/sync -resource car-data -session 9839 -driver 1 -from 2025-12-07T13:00:00Z -to 2025-12-07T13:01:00Z
```

Calendar sync replaces one season atomically, preventing clients from seeing a partially updated schedule.

## Run API

```bash
go run ./cmd/api
curl http://localhost:8080/v1/health
curl http://localhost:8080/v1/sources
curl "http://localhost:8080/v1/drivers?season=2025"
curl "http://localhost:8080/v1/meetings?season=2025"
curl "http://localhost:8080/v1/sessions?season=2025"
curl "http://localhost:8080/v1/sessions?season=2025&meeting_key=1276"
curl "http://localhost:8080/v1/calendar?season=2025"
curl http://localhost:8080/v1/calendar/next
curl "http://localhost:8080/v1/standings/drivers?season=2025"
curl "http://localhost:8080/v1/standings/constructors?season=2025"
curl "http://localhost:8080/v1/results?season=2025"
curl "http://localhost:8080/v1/results?season=2025&round=24"
curl http://localhost:8080/v1/results/latest
curl "http://localhost:8080/v1/classifications?season=2025&type=qualifying"
curl "http://localhost:8080/v1/classifications?season=2025&type=sprint&round=23"
curl "http://localhost:8080/v1/laps?session_key=9839"
curl "http://localhost:8080/v1/laps?session_key=9839&driver_number=1"
curl "http://localhost:8080/v1/laps?session_key=9839&driver_number=1&lap_number=10"
curl "http://localhost:8080/v1/stints?session_key=9839"
curl "http://localhost:8080/v1/stints?session_key=9839&driver_number=1"
curl "http://localhost:8080/v1/pit-stops?session_key=9839"
curl "http://localhost:8080/v1/pit-stops?session_key=9839&driver_number=1"
curl "http://localhost:8080/v1/weather?session_key=9839&limit=500"
curl "http://localhost:8080/v1/weather?session_key=9839&from=2025-12-07T13:00:00Z&to=2025-12-07T13:01:00Z"
curl "http://localhost:8080/v1/race-control?session_key=9839&limit=1000"
curl "http://localhost:8080/v1/race-control?session_key=9839&category=Flag"
curl "http://localhost:8080/v1/overtakes?session_key=9839&driver_number=1"
curl "http://localhost:8080/v1/positions?session_key=9839&driver_number=1"
curl "http://localhost:8080/v1/intervals?session_key=9839&driver_number=1&limit=5000"
curl "http://localhost:8080/v1/location?session_key=9839&driver_number=1&limit=5000"
curl "http://localhost:8080/v1/team-radio?session_key=9839&driver_number=1"
curl "http://localhost:8080/v1/session-timeline?session_key=9839&limit=5000"
curl "http://localhost:8080/v1/session-timeline?session_key=9839&driver_number=1&types=overtake,position,team_radio"
curl "http://localhost:8080/v1/car-data?session_key=9839&driver_number=1&limit=1000"
```

All API endpoints read SQLite only. They never contact providers. Unsynced resources return `404`. `GET /v1/calendar/next` selects the earliest stored race whose race time is in the future.

Lap, stint, and pit ingestion accept `-session` and an optional `-driver`; omit the driver to fetch the complete session. Laps include sector/mini-sector timing and speed traps. Stints include compound, lap range, and tyre age at the start. OpenF1 pit records provide the stop timestamp, lap, and pit duration; they do not provide separate entry/exit timestamps. Weather ingestion stores the complete session timeline; the API supports optional `from`, `to`, and `limit` parameters. Race-control events support optional `category`, `driver_number`, `from`, `to`, and `limit` filters. Overtake, position, and interval timelines support driver/time/limit filters; overtakes additionally support `overtaken_driver_number`. Interval labels such as `+1 LAP` are preserved alongside numeric seconds when available. Location ingestion requires one session, one driver, explicit timestamps, and a maximum 15-minute window. Team-radio sync caches audio in SQLite; returned `/v1/team-radio/{id}/audio` URLs stream locally with byte-range support, so clients never contact OpenF1. The unified `/v1/session-timeline` endpoint combines synchronized race control, overtakes, pit stops, positions, stints, team radio, and weather into one chronological feed. It supports optional `driver_number`, comma-separated `types`, `from`, `to`, and `limit` filters.

Car telemetry ingestion requires a previously synced OpenF1 session and an explicit time range no longer than 15 minutes. Re-running overlapping ranges is idempotent. The API defaults to 1,000 samples and permits at most 5,000 per response; use `from` and `to` RFC3339 filters for paging/range selection.

Environment:

| Variable | Default |
|---|---|
| `HTTP_ADDRESS` | `:8080` |
| `DB_PATH` | `data/oidysts.db` |
| `UPSTREAM_USER_AGENT` | `oidysts/0.1` |
| `UPSTREAM_TIMEOUT` | `20s` |
| `UPSTREAM_MAX_BODY_BYTES` | `10485760` |

Use an identifiable versioned User-Agent in deployed environments.

## Docker

```bash
docker compose build
docker compose run --rm api /app/probe -sources all -year 2025
docker compose run --rm api /app/sync -resource drivers -year 2025
docker compose run --rm api /app/sync -resource meetings -year 2025
docker compose run --rm api /app/sync -resource calendar -year 2025
docker compose run --rm api /app/sync -resource calendar -year 2026
docker compose run --rm api /app/sync -resource standings -year 2025
docker compose run --rm api /app/sync -resource results -year 2025
docker compose run --rm api /app/sync -resource classifications -year 2025
docker compose run --rm api /app/sync -resource laps -session 9839
docker compose run --rm api /app/sync -resource stints -session 9839
docker compose run --rm api /app/sync -resource pit -session 9839
docker compose run --rm api /app/sync -resource weather -session 9839
docker compose run --rm api /app/sync -resource race-control -session 9839
docker compose run --rm api /app/sync -resource overtakes -session 9839
docker compose run --rm api /app/sync -resource positions -session 9839
docker compose run --rm api /app/sync -resource intervals -session 9839
docker compose run --rm api /app/sync -resource location -session 9839 -driver 1 -from 2025-12-07T13:00:00Z -to 2025-12-07T13:15:00Z
docker compose run --rm api /app/sync -resource team-radio -session 9839 -driver 1
docker compose run --rm api /app/sync -resource car-data -session 9839 -driver 1 -from 2025-12-07T13:00:00Z -to 2025-12-07T13:01:00Z
docker compose up -d
```

The named volume `oidysts-data` persists SQLite for both commands.

## Postman

Import `postman/oidysts-f1-api.postman_collection.json` into Postman. The collection contains every API route, organized folders, reusable variables, parameter descriptions, and response tests. Update collection variables such as `season`, `round`, `sessionKey`, and `driverNumber` as needed.

## Verify

```bash
go test ./...
go vet ./...
```

See [source study](docs/source-study.md), [provider ownership](docs/source-ownership.md), and [architecture](docs/architecture.md). The selected provider data is non-commercial; review the documented licensing gate before monetization.
