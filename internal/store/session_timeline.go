package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"oidysts/internal/domain"
)

var sessionTimelineSelects = map[string]string{
	domain.TimelineRaceControl: `SELECT 'race_control:' || event_key AS event_id, occurred_at AS event_at,
		session_key, meeting_key, 'race_control' AS event_type, driver_number, message AS title,
		json_object('category', category, 'message', message, 'flag', flag, 'scope', scope,
			'lap_number', lap_number, 'sector', sector) AS event_data
		FROM race_control_events WHERE session_key=?`,
	domain.TimelineOvertake: `SELECT 'overtake:' || event_key AS event_id, occurred_at AS event_at,
		session_key, meeting_key, 'overtake' AS event_type, driver_number, 'Overtake' AS title,
		json_object('overtaken_driver_number', overtaken_driver_number, 'position', position) AS event_data
		FROM overtakes WHERE session_key=?`,
	domain.TimelinePitStop: `SELECT 'pit_stop:' || session_key || ':' || driver_number || ':' || stopped_at AS event_id,
		stopped_at AS event_at, session_key, meeting_key, 'pit_stop' AS event_type, driver_number, 'Pit stop' AS title,
		json_object('lap_number', lap_number, 'duration', duration) AS event_data
		FROM pit_stops WHERE session_key=?`,
	domain.TimelinePosition: `SELECT 'position:' || session_key || ':' || driver_number || ':' || sampled_at AS event_id,
		sampled_at AS event_at, session_key, meeting_key, 'position' AS event_type, driver_number, 'Position change' AS title,
		json_object('position', position) AS event_data
		FROM position_samples WHERE session_key=?`,
	domain.TimelineStint: `SELECT 'stint:' || s.session_key || ':' || s.driver_number || ':' || s.stint_number AS event_id,
		l.date_start AS event_at, s.session_key AS session_key, s.meeting_key AS meeting_key,
		'stint' AS event_type, s.driver_number AS driver_number, 'Stint started' AS title,
		json_object('stint_number', s.stint_number, 'lap_start', s.lap_start, 'lap_end', s.lap_end,
			'compound', s.compound, 'tyre_age_at_start', s.tyre_age_at_start) AS event_data
		FROM stints s JOIN laps l ON l.session_key=s.session_key AND l.driver_number=s.driver_number
			AND l.lap_number=s.lap_start
		WHERE s.session_key=? AND l.date_start IS NOT NULL`,
	domain.TimelineTeamRadio: `SELECT 'team_radio:' || id AS event_id, recorded_at AS event_at,
		session_key, meeting_key, 'team_radio' AS event_type, driver_number, 'Team radio' AS title,
		json_object('id', id, 'audio_url', '/v1/team-radio/' || id || '/audio') AS event_data
		FROM team_radio WHERE session_key=?`,
	domain.TimelineWeather: `SELECT 'weather:' || session_key || ':' || sampled_at AS event_id,
		sampled_at AS event_at, session_key, meeting_key, 'weather' AS event_type,
		NULL AS driver_number, 'Weather update' AS title,
		json_object('air_temperature', air_temperature, 'track_temperature', track_temperature,
			'humidity', humidity, 'pressure', pressure, 'rainfall', rainfall,
			'wind_direction', wind_direction, 'wind_speed', wind_speed) AS event_data
		FROM weather_samples WHERE session_key=?`,
}

func (s *Store) SessionTimeline(ctx context.Context, query domain.SessionTimelineQuery) ([]domain.SessionTimelineEvent, bool, error) {
	selects := make([]string, 0, len(query.Types))
	args := make([]any, 0, len(query.Types)+4)
	for _, eventType := range query.Types {
		statement, ok := sessionTimelineSelects[eventType]
		if !ok {
			return nil, false, fmt.Errorf("unsupported timeline type %q", eventType)
		}
		selects = append(selects, statement)
		args = append(args, query.SessionKey)
	}
	if len(selects) == 0 {
		return nil, false, fmt.Errorf("timeline types cannot be empty")
	}
	statement := `SELECT event_id, event_at, session_key, meeting_key, event_type, driver_number, title, event_data
	              FROM (` + strings.Join(selects, ` UNION ALL `) + `) WHERE 1=1`
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	if query.From != nil {
		statement += ` AND event_at>=?`
		args = append(args, query.From.UTC().Format(telemetryTimeLayout))
	}
	if query.To != nil {
		statement += ` AND event_at<=?`
		args = append(args, query.To.UTC().Format(telemetryTimeLayout))
	}
	statement += ` ORDER BY event_at, event_type, event_id LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query session timeline: %w", err)
	}
	defer rows.Close()
	events := []domain.SessionTimelineEvent{}
	for rows.Next() {
		var event domain.SessionTimelineEvent
		var timestamp, payload string
		var driver sql.NullInt64
		if err := rows.Scan(&event.ID, &timestamp, &event.SessionKey, &event.MeetingKey,
			&event.Type, &driver, &event.Title, &payload); err != nil {
			return nil, false, fmt.Errorf("scan session timeline: %w", err)
		}
		event.Timestamp, err = time.Parse(telemetryTimeLayout, timestamp)
		if err != nil {
			return nil, false, fmt.Errorf("parse session timeline time: %w", err)
		}
		if driver.Valid {
			number := int(driver.Int64)
			event.DriverNumber = &number
		}
		event.Data = json.RawMessage(payload)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	truncated := len(events) > query.Limit
	if truncated {
		events = events[:query.Limit]
	}
	if len(events) == 0 {
		return nil, false, ErrNotFound
	}
	return events, truncated, nil
}
