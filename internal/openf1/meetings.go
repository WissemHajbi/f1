package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"oidysts/internal/domain"
)

type MeetingsResult struct {
	Meetings        []domain.Meeting
	MeetingsURL     string
	MeetingsPayload []byte
	SessionsURL     string
	SessionsPayload []byte
}

func (c *Client) Meetings(ctx context.Context, year int) (MeetingsResult, error) {
	meetingsURL := c.baseURL + "/meetings?year=" + strconv.Itoa(year)
	meetingResponse, err := c.fetch.Get(ctx, meetingsURL, "application/json")
	if err != nil {
		return MeetingsResult{}, fmt.Errorf("fetch meetings: %w", err)
	}
	if err := waitForOpenF1(ctx, 350*time.Millisecond); err != nil {
		return MeetingsResult{}, err
	}
	sessionsURL := c.baseURL + "/sessions?year=" + strconv.Itoa(year)
	sessionResponse, err := c.fetch.Get(ctx, sessionsURL, "application/json")
	if err != nil {
		return MeetingsResult{}, fmt.Errorf("fetch sessions: %w", err)
	}

	var meetings []domain.Meeting
	var sessions []domain.MeetingSession
	if err := json.Unmarshal(meetingResponse.Body, &meetings); err != nil {
		return MeetingsResult{}, fmt.Errorf("decode meetings: %w", err)
	}
	if err := json.Unmarshal(sessionResponse.Body, &sessions); err != nil {
		return MeetingsResult{}, fmt.Errorf("decode sessions: %w", err)
	}
	if len(meetings) == 0 || len(sessions) == 0 {
		return MeetingsResult{}, fmt.Errorf("meetings or sessions are empty for %d", year)
	}
	byKey := make(map[int]*domain.Meeting, len(meetings))
	syncedAt := c.now().UTC()
	for index := range meetings {
		if meetings[index].Key <= 0 || meetings[index].Year != year {
			return MeetingsResult{}, fmt.Errorf("invalid meeting at index %d", index)
		}
		meetings[index].Sessions = []domain.MeetingSession{}
		meetings[index].SyncedAt = syncedAt
		byKey[meetings[index].Key] = &meetings[index]
	}
	for _, session := range sessions {
		meeting := byKey[session.MeetingKey]
		if meeting == nil {
			return MeetingsResult{}, fmt.Errorf("session %d references unknown meeting %d", session.Key, session.MeetingKey)
		}
		if session.Key <= 0 || session.Year != year {
			return MeetingsResult{}, fmt.Errorf("invalid session %d", session.Key)
		}
		meeting.Sessions = append(meeting.Sessions, session)
	}
	for index := range meetings {
		sort.Slice(meetings[index].Sessions, func(i, j int) bool {
			return meetings[index].Sessions[i].DateStart.Before(meetings[index].Sessions[j].DateStart)
		})
	}
	sort.Slice(meetings, func(i, j int) bool { return meetings[i].DateStart.Before(meetings[j].DateStart) })
	return MeetingsResult{Meetings: meetings, MeetingsURL: meetingsURL, MeetingsPayload: meetingResponse.Body,
		SessionsURL: sessionsURL, SessionsPayload: sessionResponse.Body}, nil
}

func waitForOpenF1(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
