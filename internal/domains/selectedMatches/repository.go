package selectedMatches

import "context"

// SelectedMatchesRepository contains methods that implements players struct
type SelectedMatchesRepository interface {
	Save(ctx context.Context, t SelectedMatches) (int, error)
	GetMatchesbyRequestID(ctx context.Context, matchRequestID string) ([]SelectedMatches, error)
	GetMatchesbyPlayerID(ctx context.Context, playerID string) ([]SelectedMatches, error)
	GetMatchesbyMatchRequestID(ctx context.Context, matchRequestID string) ([]SelectedMatches, error)
}
