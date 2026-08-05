# oidysts F1 backend

Small Go foundation for validating F1 data providers, storing bounded snapshots in SQLite, and exposing source health to a future React Native client.

## Requirements

- Go 1.26+
- Docker (optional)

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
```

## Run API

```bash
go run ./cmd/api
curl http://localhost:8080/v1/health
curl http://localhost:8080/v1/sources
curl "http://localhost:8080/v1/drivers?season=2025"
```

`GET /v1/drivers` reads SQLite only. It never contacts OpenF1. If the season has not been synced, it returns `404`.

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
docker compose up -d
```

The named volume `oidysts-data` persists SQLite for both commands.

## Verify

```bash
go test ./...
go vet ./...
```

See [source study](docs/source-study.md) and [architecture](docs/architecture.md). The selected provider data is non-commercial; review the documented licensing gate before monetization.
