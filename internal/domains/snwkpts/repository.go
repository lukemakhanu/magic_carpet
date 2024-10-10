package snwkpts

import "context"

// SnWkPtsRepository
type SnWkPtsRepository interface {
	Save(ctx context.Context, t SnWkPts) (int, error)
	SnWkPts(ctx context.Context, competitionID string) ([]SnWkPts, error)
}
