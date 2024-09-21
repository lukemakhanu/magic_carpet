package matchData

import (
	"context"

	"github.com/lukemakhanu/veimu_apps/internal/domain/matches"
)

// MatchFetcher is implemented to match and odds.
type MatchFetcher interface {
	GetMatch(ctx context.Context) (matches.MatchJson, error)
}
