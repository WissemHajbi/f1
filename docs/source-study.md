# Data-source study

Validated on 2026-08-05. This is an engineering summary, not legal advice.

| Source | Verified access | Limits / terms | Recommended use |
|---|---|---|---|
| OpenF1 | Historical REST works without auth; data starts in 2023. Live data is paid. | Community: 3 requests/second and 30/minute. Site states CC BY-NC-SA 4.0 and personal/non-commercial use. | Cache session metadata. Fetch high-frequency telemetry only for a driver and bounded time/lap range, then retain it. Never proxy arbitrary mobile requests upstream. |
| Jolpica-F1 | Calendar, standings, and results REST responses verified. Custom `User-Agent` required. | 4 requests/second burst; 500/hour sustained. Non-commercial data under CC BY-NC-SA 4.0. Maximum page size 100. Limits can change. | Scheduled server-side sync and local cache. Use broad/paginated queries instead of per-driver/per-lap calls. |
| Motorsport RSS | Feed verified. Advertised TTL was 100 minutes in the sampled response. | Publisher copyright remains. RSS availability is not permission to republish full articles or images. | Store headline, canonical URL, source, timestamp, and feed-provided thumbnail only when its terms permit; link to publisher. Poll no faster than TTL. |
| Sky Sports RSS | Feed verified; HTTP cache headers advertised roughly two minutes. | Copyright notice in feed. | Poll every 15 minutes; metadata and outbound links only pending a terms review. |
| BBC Sport RSS | The supplied legacy URL returns 404. | Correct URL: `https://feeds.bbci.co.uk/sport/formula1/rss.xml`. Publisher terms apply. | Poll every 15 minutes; metadata and outbound links only pending a terms review. |
| FIA documents | Public document page verified. | `robots.txt` specifies `Crawl-delay: 10`. PDF reuse rights were not established. Page shape may change. | Discover slowly and record URL/hash/title. Download each new PDF once. Keep extracted facts with document provenance; do not redistribute PDFs by default. |
| Team metadata | No authoritative timing endpoint for management/engineer rosters. | Facts change and branding assets/colors may have trademark/copyright constraints. | Version-controlled JSON with `last_verified_at` and source provenance added before production. Empty 2026 skeleton is intentional until facts are verified. |
| Official F1 media CDN | URLs can appear in API responses. | A reachable URL does not grant redistribution rights. Hotlink stability is not guaranteed. | Do not bundle/cache commercially without permission. Prefer licensed assets or user-agent fallback artwork. |
| Wikimedia Commons | Per-file licenses and attribution differ. | Attribution/share-alike requirements are asset-specific. | Persist creator, license, source page, and attribution with every selected image. |
| Sketchfab / commercial 3D | API/account and per-model license rules apply. | Downloadability is model-specific. Commercial marketplace files cannot be redistributed freely. Sim-game conversions are especially risky. | Defer. Use only owned or explicitly licensed mobile-optimized GLB assets with recorded provenance. |

## FastF1 and LiveF1

These are Python libraries, not a way around upstream data terms. They are useful later for CPU-heavy telemetry normalization, lap deltas, traces, and degradation models. Do not add a Python service to milestone 1. Add one only when a measured analysis requirement cannot be handled cleanly by batch jobs.

## Fetch policy

- Calendar/session index: daily; additionally 30 minutes after a session.
- Jolpica standings/results: every 15 minutes for two hours after a race, otherwise daily.
- OpenF1 telemetry/location: on explicit ingestion jobs, bounded by driver/lap/time; immutable historical payloads need no recurring refresh.
- RSS: every 15 minutes or the feed's longer TTL; deduplicate by canonical URL/GUID.
- FIA: daily normally; every 30 minutes on Thursday/Friday of a race weekend, always respecting at least 10 seconds between FIA requests.
- Apply timeout, response-size cap, conditional requests where supported, exponential backoff with jitter on 429/5xx, and a global per-provider limiter.

The current `probe` command is deliberately manual. It validates and snapshots bounded payloads in SQLite. Production scheduling comes after schemas and desired mobile screens are fixed.

## Attribution and commercialization gate

Both primary data APIs currently impose non-commercial/share-alike terms. Before ads, subscriptions, sponsorship, paid distribution, or another commercial use, obtain written terms from providers and review F1 trademarks/media rights. Preserve source attribution and provenance in the data model.
