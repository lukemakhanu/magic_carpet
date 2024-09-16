package slowRedis

import "context"

// RunRedis is implemented to save feed.
type SlowRedis interface {
	GetZRange(ctx context.Context, set string) ([]string, error)
	GetZRangeWithLimit(ctx context.Context, set string, fetched int) ([]string, error)
}
