package periods

import "context"

type PeriodsRepository interface {
	Save(ctx context.Context, t Periods) (int, error)
	PeriodByMatchRequestID(ctx context.Context, matchRequestID string) ([]Periods, error)
	UpdateEarlyFinish(ctx context.Context, earlyFinish, periodID string) (int64, error)
	UpdateStartTime(ctx context.Context, startTime, endTime, played, periodID string) (int64, error)
	UpdateKeyCreated(ctx context.Context, keyCreated, periodID string) (int64, error)
}
