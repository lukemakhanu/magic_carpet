package goalPatterns

import "context"

// GoalPatternsRepository
type GoalPatternsRepository interface {
	Save(ctx context.Context, t GoalPatterns) (int, error)
	GoalDistributions(ctx context.Context, competitionID string) ([]GoalPatterns, error)
}
