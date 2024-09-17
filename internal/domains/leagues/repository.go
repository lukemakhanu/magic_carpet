package leagues

import "context"

// LeaguesRepository contains methods that implements leagues struct
type LeaguesRepository interface {
	Save(ctx context.Context, t Leagues) (int, error)
	GetLeagues(ctx context.Context) ([]Leagues, error)
	GetLeagueByID(ctx context.Context, leagueID string) ([]Leagues, error)
	GetLeagueByClientID(ctx context.Context, clientID string) ([]Leagues, error)
	UpdateLeague(ctx context.Context, clientID, league, leagueAbbrv, leagueID string) (int64, error)

	GetMatchDetails(ctx context.Context, leagueAbbr string) ([]MatchDetails, error)
	GetWinningOutcomes(ctx context.Context, seasonWeekID string) ([]MatchWinningOutcomeAPI, error)

	GetProductionMatchDetails(ctx context.Context, leagueAbbr string) ([]MatchDetails, error)
	GetProductionWinningOutcomes(ctx context.Context, seasonWeekID string) ([]MatchWinningOutcomeAPI, error)

	GetProductionMatchDetailsNew(ctx context.Context, leagueAbbr string) ([]MatchDetails, error)
	GetProductionWinningOutcomesNew(ctx context.Context, seasonWeekID string) ([]MatchWinningOutcomeAPI, error)
}
