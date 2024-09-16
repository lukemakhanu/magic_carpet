package scheduledTimes

import "context"

// ScheduledTimesRepository contains methods that implements scheduled time struct
type ScheduledTimesRepository interface {
	Save(ctx context.Context, t ScheduledTime) (int, error)
	CountScheduledTime(ctx context.Context, leagueID string) (int, error)
	Delete(ctx context.Context) (int64, error)
	UpdateScheduleTime(ctx context.Context, status, scheduledTimeID string) (int64, error)
	GetScheduledTime(ctx context.Context, status, competitionID string) ([]ScheduledTime, error)
	FirstActiveScheduledTime(ctx context.Context, leagueID, status string) (string, string, error)
}
