# Architecture

## Milestone 1

```mermaid
flowchart LR
  Probe[Go probe command] --> Fetch[Bounded HTTP client]
  Fetch --> OF1[OpenF1 historical REST]
  Fetch --> J[Jolpica REST]
  Fetch --> RSS[RSS feeds]
  Fetch --> FIA[FIA document index]
  Static[Versioned team JSON] --> Probe
  Probe --> DB[(SQLite snapshots)]
  API[Go HTTP API] --> DB
  RN[React Native app later] --> API
```

The mobile app never calls providers directly. The Go backend owns limits, attribution, normalization, caching, and provider changes. SQLite is embedded in the Go process and persisted through one Docker volume; a second “SQLite container” would be incorrect.

## Intended evolution

```mermaid
flowchart TB
  RN[React Native / Expo client] -->|HTTPS normalized JSON| API[Go API]
  API --> Cache[(SQLite; PostgreSQL only if scaling requires it)]
  Scheduler[Go scheduled jobs] --> Ingest[Provider adapters]
  Ingest --> OF1[OpenF1]
  Ingest --> J[Jolpica]
  Ingest --> Feeds[Publisher RSS]
  Ingest --> FIADocs[FIA pages and PDFs]
  Metadata[Reviewed static metadata] --> Ingest
  Ingest --> Cache
  Cache -. telemetry work queue .-> Py[Optional FastF1/LiveF1 batch worker]
  Py -. derived traces/deltas .-> Cache
  Assets[Licensed image/3D object storage] --> RN
```

## Package boundaries

- `internal/fetch`: safe upstream HTTP transport.
- `internal/openf1`: typed OpenF1 ingestion client.
- `internal/probe`: source-specific validation and summaries.
- `internal/store`: SQLite schema, raw snapshots, and normalized persistence.
- `internal/server`: small HTTP surface.
- `cmd/probe`: explicit external integration smoke test.
- `cmd/sync`: explicit provider-to-normalized-database ingestion.
- `cmd/api`: database-backed client API.

Driver rosters, OpenF1 sessions, Jolpica events, event schedules, and both championship standings are normalized. All client API endpoints read only SQLite. Raw snapshots remain diagnostic evidence. Next, add normalized race results, articles, teams, upgrades, and asset provenance. High-frequency telemetry should use chunked compressed blobs or derived samples rather than one SQLite row per frame.
