package ssns

import "context"

// SsnsRepository contains methods that implements seasons struct
type SsnsRepository interface {
	Save(ctx context.Context, t Ssns) (int, error)
	GetSeasons(ctx context.Context, status string) ([]Ssns, error)
	GetSeasonByID(ctx context.Context, seasonID string) ([]Ssns, error)
	UpdateSeason(ctx context.Context, leagueID, status, seasonID string) (int64, error)

	// ssn week
	SaveSsnWeek(ctx context.Context, leagueID, seasonID, weekNumber, status, startTime, endTime string) (int, error)
	SaveGames(ctx context.Context, leagueID, seasonID, seasonWeekID, weekNumber, homeTeamID, awayTeamID, status string) (int, error)

	GetLastGame(ctx context.Context, leagueID string) ([]LastGameTime, error)
	CountRemainingPeriods(ctx context.Context, leagueID string) (int, error)
}
