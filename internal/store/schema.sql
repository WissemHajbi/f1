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

CREATE TABLE IF NOT EXISTS result_races (
    season       INTEGER NOT NULL,
    round        INTEGER NOT NULL,
    name         TEXT NOT NULL,
    race_at      TEXT NOT NULL,
    circuit_id   TEXT NOT NULL,
    circuit_name TEXT NOT NULL,
    locality     TEXT NOT NULL,
    country      TEXT NOT NULL,
    latitude     REAL NOT NULL,
    longitude    REAL NOT NULL,
    source       TEXT NOT NULL,
    synced_at    TEXT NOT NULL,
    PRIMARY KEY (season, round)
);
CREATE INDEX IF NOT EXISTS idx_result_races_date ON result_races(race_at DESC);

CREATE TABLE IF NOT EXISTS race_results (
    season                  INTEGER NOT NULL,
    round                   INTEGER NOT NULL,
    position                INTEGER NOT NULL,
    position_text           TEXT NOT NULL,
    points                  REAL NOT NULL,
    grid_position           INTEGER NOT NULL,
    laps                    INTEGER NOT NULL,
    status                  TEXT NOT NULL,
    result_time             TEXT NOT NULL,
    time_millis             INTEGER,
    driver_id               TEXT NOT NULL,
    driver_number           TEXT NOT NULL,
    driver_code             TEXT NOT NULL,
    given_name              TEXT NOT NULL,
    family_name             TEXT NOT NULL,
    driver_nationality      TEXT NOT NULL,
    constructor_id          TEXT NOT NULL,
    constructor_name        TEXT NOT NULL,
    constructor_nationality TEXT NOT NULL,
    fastest_lap_rank        INTEGER,
    fastest_lap_number      INTEGER,
    fastest_lap_time        TEXT,
    fastest_lap_speed       REAL,
    fastest_lap_speed_units TEXT,
    PRIMARY KEY (season, round, driver_id),
    FOREIGN KEY (season, round) REFERENCES result_races(season, round) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_race_results_position ON race_results(season, round, position);

CREATE TABLE IF NOT EXISTS openf1_meetings (
    meeting_key        INTEGER PRIMARY KEY,
    year               INTEGER NOT NULL,
    name               TEXT NOT NULL,
    official_name      TEXT NOT NULL,
    location           TEXT NOT NULL,
    country_key        INTEGER NOT NULL,
    country_code       TEXT NOT NULL,
    country_name       TEXT NOT NULL,
    circuit_key        INTEGER NOT NULL,
    circuit_short_name TEXT NOT NULL,
    date_start         TEXT NOT NULL,
    gmt_offset         TEXT NOT NULL,
    source             TEXT NOT NULL,
    synced_at          TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_openf1_meetings_year_date ON openf1_meetings(year, date_start);

CREATE TABLE IF NOT EXISTS openf1_sessions (
    session_key        INTEGER PRIMARY KEY,
    meeting_key        INTEGER NOT NULL REFERENCES openf1_meetings(meeting_key) ON DELETE CASCADE,
    year               INTEGER NOT NULL,
    name               TEXT NOT NULL,
    type               TEXT NOT NULL,
    location           TEXT NOT NULL,
    country_code       TEXT NOT NULL,
    country_name       TEXT NOT NULL,
    circuit_key        INTEGER NOT NULL,
    circuit_short_name TEXT NOT NULL,
    date_start         TEXT NOT NULL,
    date_end           TEXT NOT NULL,
    gmt_offset         TEXT NOT NULL,
    is_cancelled       INTEGER NOT NULL,
    source             TEXT NOT NULL,
    synced_at          TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_openf1_sessions_meeting_date ON openf1_sessions(meeting_key, date_start);
CREATE INDEX IF NOT EXISTS idx_openf1_sessions_year_date ON openf1_sessions(year, date_start);

CREATE TABLE IF NOT EXISTS race_session_links (
    season      INTEGER NOT NULL,
    round       INTEGER NOT NULL,
    session_key INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    source      TEXT NOT NULL,
    linked_at   TEXT NOT NULL,
    PRIMARY KEY (season, round),
    UNIQUE (session_key),
    FOREIGN KEY (season, round) REFERENCES events(season, round) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS car_data_samples (
    session_key  INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number INTEGER NOT NULL,
    sampled_at   TEXT NOT NULL,
    meeting_key  INTEGER NOT NULL,
    speed        INTEGER NOT NULL,
    rpm          INTEGER NOT NULL,
    gear         INTEGER NOT NULL,
    throttle     INTEGER NOT NULL,
    brake        INTEGER NOT NULL,
    drs          INTEGER NOT NULL,
    source       TEXT NOT NULL,
    synced_at    TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, sampled_at)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS location_samples (
    session_key   INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number INTEGER NOT NULL,
    sampled_at    TEXT NOT NULL,
    meeting_key   INTEGER NOT NULL,
    x             INTEGER NOT NULL,
    y             INTEGER NOT NULL,
    z             INTEGER NOT NULL,
    source        TEXT NOT NULL,
    synced_at     TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, sampled_at)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS session_classifications (
    season       INTEGER NOT NULL,
    round        INTEGER NOT NULL,
    type         TEXT NOT NULL,
    name         TEXT NOT NULL,
    start_at     TEXT,
    circuit_id   TEXT NOT NULL,
    circuit_name TEXT NOT NULL,
    locality     TEXT NOT NULL,
    country      TEXT NOT NULL,
    latitude     REAL NOT NULL,
    longitude    REAL NOT NULL,
    source       TEXT NOT NULL,
    synced_at    TEXT NOT NULL,
    PRIMARY KEY (season, round, type)
);

CREATE TABLE IF NOT EXISTS session_classification_results (
    season                  INTEGER NOT NULL,
    round                   INTEGER NOT NULL,
    type                    TEXT NOT NULL,
    position                INTEGER NOT NULL,
    driver_id               TEXT NOT NULL,
    driver_number           TEXT NOT NULL,
    driver_code             TEXT NOT NULL,
    given_name              TEXT NOT NULL,
    family_name             TEXT NOT NULL,
    nationality             TEXT NOT NULL,
    constructor_id          TEXT NOT NULL,
    constructor_name        TEXT NOT NULL,
    constructor_nationality TEXT NOT NULL,
    q1                      TEXT NOT NULL,
    q2                      TEXT NOT NULL,
    q3                      TEXT NOT NULL,
    points                  REAL,
    grid_position           INTEGER,
    laps                    INTEGER,
    status                  TEXT NOT NULL,
    result_time             TEXT NOT NULL,
    time_millis             INTEGER,
    fastest_lap_rank        INTEGER,
    fastest_lap_number      INTEGER,
    fastest_lap_time        TEXT,
    fastest_lap_speed       REAL,
    fastest_lap_speed_units TEXT,
    PRIMARY KEY (season, round, type, driver_id),
    FOREIGN KEY (season, round, type) REFERENCES session_classifications(season, round, type) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_session_classification_position
    ON session_classification_results(season, round, type, position);

CREATE TABLE IF NOT EXISTS overtakes (
    event_key               TEXT PRIMARY KEY,
    session_key             INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    occurred_at             TEXT NOT NULL,
    meeting_key             INTEGER NOT NULL,
    driver_number           INTEGER NOT NULL,
    overtaken_driver_number INTEGER NOT NULL,
    position                INTEGER NOT NULL,
    source                  TEXT NOT NULL,
    synced_at               TEXT NOT NULL
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS idx_overtakes_session_time ON overtakes(session_key, occurred_at);
CREATE INDEX IF NOT EXISTS idx_overtakes_driver ON overtakes(session_key, driver_number);

CREATE TABLE IF NOT EXISTS position_samples (
    session_key   INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number INTEGER NOT NULL,
    sampled_at    TEXT NOT NULL,
    meeting_key   INTEGER NOT NULL,
    position      INTEGER NOT NULL,
    source        TEXT NOT NULL,
    synced_at     TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, sampled_at)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS interval_samples (
    session_key          INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number        INTEGER NOT NULL,
    sampled_at           TEXT NOT NULL,
    meeting_key          INTEGER NOT NULL,
    gap_to_leader        TEXT NOT NULL,
    gap_to_leader_seconds REAL,
    interval_value       TEXT NOT NULL,
    interval_seconds     REAL,
    source               TEXT NOT NULL,
    synced_at            TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, sampled_at)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS team_radio (
    id               TEXT PRIMARY KEY,
    session_key      INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    recorded_at      TEXT NOT NULL,
    meeting_key      INTEGER NOT NULL,
    driver_number    INTEGER NOT NULL,
    recording_source TEXT NOT NULL,
    content_type     TEXT NOT NULL,
    audio_data       BLOB NOT NULL,
    source           TEXT NOT NULL,
    synced_at        TEXT NOT NULL
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS idx_team_radio_session_time ON team_radio(session_key, recorded_at);
CREATE INDEX IF NOT EXISTS idx_team_radio_driver ON team_radio(session_key, driver_number);

CREATE TABLE IF NOT EXISTS race_control_events (
    event_key     TEXT PRIMARY KEY,
    session_key   INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    occurred_at   TEXT NOT NULL,
    meeting_key   INTEGER NOT NULL,
    category      TEXT NOT NULL,
    message       TEXT NOT NULL,
    flag          TEXT NOT NULL,
    scope         TEXT NOT NULL,
    driver_number INTEGER,
    lap_number    INTEGER,
    sector        INTEGER,
    source        TEXT NOT NULL,
    synced_at     TEXT NOT NULL
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS idx_race_control_session_time ON race_control_events(session_key, occurred_at);
CREATE INDEX IF NOT EXISTS idx_race_control_session_category ON race_control_events(session_key, category);

CREATE TABLE IF NOT EXISTS weather_samples (
    session_key       INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    sampled_at        TEXT NOT NULL,
    meeting_key       INTEGER NOT NULL,
    air_temperature   REAL,
    track_temperature REAL,
    humidity          REAL,
    pressure          REAL,
    rainfall          REAL,
    wind_direction    REAL,
    wind_speed        REAL,
    source            TEXT NOT NULL,
    synced_at         TEXT NOT NULL,
    PRIMARY KEY (session_key, sampled_at)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS pit_stops (
    session_key   INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number INTEGER NOT NULL,
    stopped_at    TEXT NOT NULL,
    meeting_key   INTEGER NOT NULL,
    lap_number    INTEGER NOT NULL,
    duration      REAL,
    source        TEXT NOT NULL,
    synced_at     TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, stopped_at)
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS idx_pit_stops_session_lap ON pit_stops(session_key, lap_number);

CREATE TABLE IF NOT EXISTS stints (
    session_key      INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number    INTEGER NOT NULL,
    stint_number     INTEGER NOT NULL,
    meeting_key      INTEGER NOT NULL,
    lap_start        INTEGER NOT NULL,
    lap_end          INTEGER,
    compound         TEXT NOT NULL,
    tyre_age_at_start INTEGER,
    source           TEXT NOT NULL,
    synced_at        TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, stint_number)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS laps (
    session_key      INTEGER NOT NULL REFERENCES openf1_sessions(session_key) ON DELETE CASCADE,
    driver_number    INTEGER NOT NULL,
    lap_number       INTEGER NOT NULL,
    meeting_key      INTEGER NOT NULL,
    date_start       TEXT,
    lap_duration     REAL,
    sector_1_duration REAL,
    sector_2_duration REAL,
    sector_3_duration REAL,
    i1_speed         INTEGER,
    i2_speed         INTEGER,
    speed_trap       INTEGER,
    sector_1_segments TEXT NOT NULL,
    sector_2_segments TEXT NOT NULL,
    sector_3_segments TEXT NOT NULL,
    is_pit_out_lap   INTEGER NOT NULL,
    source           TEXT NOT NULL,
    synced_at        TEXT NOT NULL,
    PRIMARY KEY (session_key, driver_number, lap_number)
) WITHOUT ROWID;
