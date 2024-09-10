package lsFiles

import "context"

// LsFilesRepository :
type LsFilesRepository interface {
	Save(ctx context.Context, t LsFiles) (int, error)
	GetLsFiles(ctx context.Context) ([]LsFiles, error)
	GetLsFileByID(ctx context.Context, lsFileID string) ([]LsFiles, error)
	UpdateLsFile(ctx context.Context, lsFileName, lsDir, country, lsExtID, lsFileID string) (int64, error)
	GetLsFileByExtID(ctx context.Context, lsExtID string) ([]LsFiles, error)

	GetLiveScoreFileByExtID(ctx context.Context, lsExtID, countryCode string) ([]LsFiles, error)
}
