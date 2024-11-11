package usedMatches

import (
	"context"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goals"
)

type UsedMatchesRepository interface {
	Save(ctx context.Context, t UsedMatches) (int, error)
	GetAvailable(ctx context.Context, category string) ([]goals.Goals, error)
	GetMatchDetails(ctx context.Context, category, matchID string) ([]goals.Goals, error)
}
