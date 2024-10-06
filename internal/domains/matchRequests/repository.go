package matchRequests

import "context"

// MatchRequestsRepository contains methods that implements matchRequests struct
type MatchRequestsRepository interface {
	Save(ctx context.Context, t MatchRequests) (int, error)
	UpdateEarlyFinish(ctx context.Context, earlyFinish, matchRequestID string) (int64, error)
	UpdatePlayed(ctx context.Context, earlyFinish, matchRequestID string) (int64, error)
	PendingRequestedMatchDesc(ctx context.Context, playerID, competitionID, keyCreated string) ([]MatchRequests, error)
	PlayerUsedMatchesDesc(ctx context.Context, playerID string) ([]UsedParentMatchIDs, error)
}
