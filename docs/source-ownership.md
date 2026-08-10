# Provider ownership

The backend does not merge duplicate classifications. Each resource has one authoritative provider within this application:

| Resource | Owner |
|---|---|
| Calendar, race results, qualifying, sprint, championship standings | Jolpica |
| Practice classifications (when implemented) | OpenF1 |
| Meetings/session keys, laps, telemetry, stints, pits, weather, race control | OpenF1 |
| Attributed geographic circuit centerlines | OpenStreetMap (ODbL) |

Rules:

1. Do not ingest OpenF1 race, qualifying, or sprint classifications.
2. Do not use OpenF1 keys as Jolpica IDs or infer that they are interchangeable.
3. Keep provider references and `source` metadata in storage.
4. Historical availability is explicit: OpenF1 detail generally starts in 2023.
5. The API presents stable application models, while provider adapters remain isolated.
6. OpenStreetMap centerlines are cached locally and may be buffered only as estimated visual width; they are not official asphalt boundaries.

This policy prevents conflict resolution, silent overwrites, and source-dependent frontend behavior.
