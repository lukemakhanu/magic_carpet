package goals

import "context"

// GoalsRepository
type GoalsRepository interface {
	Save(ctx context.Context, t Goals) (int, error)
	GetGoalCategory(ctx context.Context, country, category, projectID string) ([]Goals, error)
}
