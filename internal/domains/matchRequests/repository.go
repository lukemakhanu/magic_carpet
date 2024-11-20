package matchRequests

import "context"

// MatchRequestsRepository contains methods that implements matchRequests struct
type MatchRequestsRepository interface {
	Save(ctx context.Context, t MatchRequests) (int, error)
	MatchRequestDesc(ctx context.Context, playerID string) ([]MatchRequests, error)
}
