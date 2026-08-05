CREATE TABLE IF NOT EXISTS source_snapshots (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    source        TEXT NOT NULL,
    resource      TEXT NOT NULL,
    endpoint      TEXT NOT NULL,
    fetched_at    TEXT NOT NULL,
    status_code   INTEGER NOT NULL,
    content_type  TEXT NOT NULL,
    record_count  INTEGER NOT NULL DEFAULT 0 CHECK (record_count >= 0),
    payload_bytes INTEGER NOT NULL DEFAULT 0 CHECK (payload_bytes >= 0),
    summary       BLOB,
    payload       BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_source_snapshots_latest
    ON source_snapshots(source, resource, fetched_at DESC);

CREATE TABLE IF NOT EXISTS sessions (
    session_key INTEGER PRIMARY KEY,
    year        INTEGER NOT NULL,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,
    location    TEXT NOT NULL,
    date_start  TEXT NOT NULL,
    date_end    TEXT NOT NULL,
    source      TEXT NOT NULL,
    synced_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_year_start
    ON sessions(year, date_start DESC);

CREATE TABLE IF NOT EXISTS drivers (
    session_key    INTEGER NOT NULL REFERENCES sessions(session_key) ON DELETE CASCADE,
    driver_number INTEGER NOT NULL,
    broadcast_name TEXT NOT NULL,
    full_name      TEXT NOT NULL,
    acronym        TEXT NOT NULL,
    team_name      TEXT NOT NULL,
    team_colour    TEXT NOT NULL,
    first_name     TEXT NOT NULL,
    last_name      TEXT NOT NULL,
    headshot_url   TEXT NOT NULL,
    country_code   TEXT,
    source         TEXT NOT NULL,
    synced_at      TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number)
);
CREATE INDEX IF NOT EXISTS idx_drivers_session
    ON drivers(session_key, driver_number);
