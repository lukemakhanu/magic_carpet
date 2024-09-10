package liveScoreFiles

import "context"

// LiveScoreFilesRepository :
type LiveScoreFilesRepository interface {
	Save(ctx context.Context, t LiveScoreFiles) (int, error)
	GetLsFiles(ctx context.Context) ([]LiveScoreFiles, error)
	GetLsFileByID(ctx context.Context, lsFileID string) ([]LiveScoreFiles, error)
	UpdateLsFile(ctx context.Context, lsFileName, lsDir, country, lsExtID, lsFileID string) (int64, error)
}
