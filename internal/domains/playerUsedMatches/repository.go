package playerUsedMatches

import (
	"context"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goals"
)

type PlayerUsedMatchesRepository interface {
	Save(ctx context.Context, t PlayerUsedMatches) (int, error)
	GetAvailable(ctx context.Context, category string) ([]goals.Goals, error)
	GetMatchDetails(ctx context.Context, category, matchID string) ([]goals.Goals, error)
}
