package mrs

import "context"

// MrsRepository
type MrsRepository interface {
	Save(context.Context, Mrs) (int, error)
	GoalPatterns(ctx context.Context, competitionID string) ([]Mrs, error)
}