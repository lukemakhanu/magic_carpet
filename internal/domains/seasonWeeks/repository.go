package seasonWeeks

import "context"

// LeaguesRepository contains methods that implements leagues struct
type SeasonWeeksRepository interface {
	Save(ctx context.Context, t SeasonWeeks) (int, error)
	GetSeasonWeeks(ctx context.Context) ([]SeasonWeeks, error)
	GetSeasonWeekByID(ctx context.Context, seasonWeekID string) ([]SeasonWeeks, error)
	UpdateSeasonWeek(ctx context.Context, seasonWeekID, seasonID, weekNumber, status, startTime, endTime string) (int64, error)
	UpcomingSeasonWeeks(ctx context.Context) ([]SeasonWeeks, error)
	UpdateSeasonWeekStatus(ctx context.Context, seasonWeekID, status string) (int64, error)
	UpcomingSeasonWeeksTest(ctx context.Context) ([]SeasonWeekDetails, error)

	ApiSeasonWeeks(ctx context.Context, seasonID string) ([]SeasonWeeksAPI, error)

	UpcomingSsnWeeks(ctx context.Context) ([]SeasonWeekDetails, error)
	UpdateSsnWeekStatus(ctx context.Context, seasonWeekID, seasonID, status string) (int64, error)

	ApiSsnWeeksNew(ctx context.Context, seasonID string) ([]ProductionSeasonWeeksAPI, error)
}
