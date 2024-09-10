package woFiles

import "context"

// LsFilesRepository :
type WoFilesRepository interface {
	Save(ctx context.Context, t WoFiles) (int, error)
	SaveWo(ctx context.Context, t WoFiles) (int, error)
	GetWoFiles(ctx context.Context) ([]WoFiles, error)
	GetWoFileByID(ctx context.Context, woFileID string) ([]WoFiles, error)
	UpdateWoFile(ctx context.Context, woFileName, woDir, country, woExtID, woFileID string) (int64, error)

	GetWinningOutcomeFiles(ctx context.Context) ([]WinningOutcomeFiles, error)

	GetWO(ctx context.Context, statement string) ([]WinningOutcomeFiles, error)
}
