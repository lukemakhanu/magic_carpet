package selectedMatches

import "context"

// SelectedMatchesRepository contains methods that implements players struct
type SelectedMatchesRepository interface {
	Save(ctx context.Context, t SelectedMatches) (int, error)
	GetMatchesbyPlayerID(ctx context.Context, playerID string) ([]SelectedMatches, error)
	GetMatchesbyPeriodID(ctx context.Context, periodID string) ([]SelectedMatches, error)
}
