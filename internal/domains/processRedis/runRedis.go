package processRedis

import "context"

// RunRedis is implemented to save feed.
type RunRedis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	SetWithExpiry(ctx context.Context, key, value, expiry string) error
	HSet(ctx context.Context, key, attribute, value string) error
	HGet(ctx context.Context, key, field string) (string, error)
	HmSet(ctx context.Context, set string, m map[string]string) error
	ZAdd(ctx context.Context, set, priority, value string) error
	GetZRange(ctx context.Context, set string) ([]string, error)
	GetZRangeWithLimit(ctx context.Context, set string, fetched int) ([]string, error)
	ZRevRange(ctx context.Context, set string) ([]string, error)
	ZRem(ctx context.Context, nameOfSet string, val string) (interface{}, error)
	Delete(ctx context.Context, key string) (interface{}, error)
	SortedSetLen(ctx context.Context, key string) (int, error)

	GetZRevRangeWithLimit(ctx context.Context, set string, fetched int) ([]string, error)
}
