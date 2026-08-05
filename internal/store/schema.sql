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

CREATE TABLE IF NOT EXISTS events (
    season          INTEGER NOT NULL,
    round           INTEGER NOT NULL,
    name            TEXT NOT NULL,
    source_url      TEXT NOT NULL,
    race_at         TEXT NOT NULL,
    circuit_id      TEXT NOT NULL,
    circuit_name    TEXT NOT NULL,
    locality        TEXT NOT NULL,
    country         TEXT NOT NULL,
    latitude        REAL NOT NULL,
    longitude       REAL NOT NULL,
    source          TEXT NOT NULL,
    synced_at       TEXT NOT NULL,
    PRIMARY KEY (season, round)
);
CREATE INDEX IF NOT EXISTS idx_events_race_at ON events(race_at);

CREATE TABLE IF NOT EXISTS event_sessions (
    season    INTEGER NOT NULL,
    round     INTEGER NOT NULL,
    type      TEXT NOT NULL,
    start_at  TEXT NOT NULL,
    PRIMARY KEY (season, round, type),
    FOREIGN KEY (season, round) REFERENCES events(season, round) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_event_sessions_start ON event_sessions(start_at);

CREATE TABLE IF NOT EXISTS driver_standings (
    season           INTEGER NOT NULL,
    driver_id        TEXT NOT NULL,
    round            INTEGER NOT NULL,
    position         INTEGER NOT NULL,
    points           REAL NOT NULL,
    wins             INTEGER NOT NULL,
    permanent_number TEXT NOT NULL,
    code             TEXT NOT NULL,
    given_name       TEXT NOT NULL,
    family_name      TEXT NOT NULL,
    date_of_birth    TEXT NOT NULL,
    nationality      TEXT NOT NULL,
    source           TEXT NOT NULL,
    synced_at        TEXT NOT NULL,
    PRIMARY KEY (season, driver_id)
);
CREATE INDEX IF NOT EXISTS idx_driver_standings_position ON driver_standings(season, position);

CREATE TABLE IF NOT EXISTS driver_standing_constructors (
    season                  INTEGER NOT NULL,
    driver_id               TEXT NOT NULL,
    constructor_id          TEXT NOT NULL,
    constructor_name        TEXT NOT NULL,
    constructor_nationality TEXT NOT NULL,
    PRIMARY KEY (season, driver_id, constructor_id),
    FOREIGN KEY (season, driver_id) REFERENCES driver_standings(season, driver_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS constructor_standings (
    season                  INTEGER NOT NULL,
    constructor_id          TEXT NOT NULL,
    round                   INTEGER NOT NULL,
    position                INTEGER NOT NULL,
    points                  REAL NOT NULL,
    wins                    INTEGER NOT NULL,
    constructor_name        TEXT NOT NULL,
    constructor_nationality TEXT NOT NULL,
    source                  TEXT NOT NULL,
    synced_at               TEXT NOT NULL,
    PRIMARY KEY (season, constructor_id)
);
CREATE INDEX IF NOT EXISTS idx_constructor_standings_position ON constructor_standings(season, position);
