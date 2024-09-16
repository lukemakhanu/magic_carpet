package matches

import "context"

// MatchesRepository contains methods that implements matches struct
type MatchesRepository interface {
	Save(ctx context.Context, t Matches) (int, error)
	GetMatches(ctx context.Context) ([]Matches, error)
	GetMatchByID(ctx context.Context, matchID string) ([]Matches, error)
	UpdateMatch(ctx context.Context, seasonWeekID, HomeTeamID, awayTeamID, status, matchID string) (int64, error)
	GetMatchBySeasonWeek(ctx context.Context, seasonWeekID string) ([]Matches, error)
	UpdateMatchStatus(ctx context.Context, status, matchID string) (int64, error)

	GetSeasonWeekGames(ctx context.Context, seasonWeekID, seasonID string) ([]MatchGames, error)

	UpdateGameStatus(ctx context.Context, status, matchID string) (int64, error)
}
