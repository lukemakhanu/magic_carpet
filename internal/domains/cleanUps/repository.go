package cleanUps

import "context"

// CleanUpsRepository
type CleanUpsRepository interface {
	Save(ctx context.Context, t CleanUps) (int, error)
	CleanUpsByStatus(ctx context.Context, status string) ([]CleanUps, error)
	LastCleanUps(ctx context.Context) ([]CleanUps, error)
	UpdateCleanUps(ctx context.Context, cleanUpID, status string) (int64, error)
	SaveForTomorrow(ctx context.Context, t CleanUps) (int, error)
}
