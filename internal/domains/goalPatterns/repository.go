package goalPatterns

import "context"

// GoalPatternsRepository
type GoalPatternsRepository interface {
	Save(ctx context.Context, t GoalPatterns) (int, error)
}
