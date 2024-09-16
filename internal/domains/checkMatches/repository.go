package checkMatches

import "context"

// CheckMatchesRepository contains methods that implements check matches struct
type CheckMatchesRepository interface {
	Save(ctx context.Context, t CheckMatches) (int, error)
	MatchExist(ctx context.Context, country, matchID string) (bool, int, string, error)
	Delete(ctx context.Context) (int64, error)
}
