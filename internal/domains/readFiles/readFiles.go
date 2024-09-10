package readFiles

import (
	"context"

	"github.com/lukemakhanu/magic_carpet/internal/domains/lsFiles"
)

// DirectoryReader is implemented to save feed.
type DirectoryReader interface {
	ReadDirectory(ctx context.Context) ([]string, error)
	AllFiles(ctx context.Context) ([]lsFiles.FileInfo, error)
	SameRoundLiveScores(ctx context.Context, fileNames []string) ([]string, error)
	ReadLiveScoreDirectory(ctx context.Context) ([]lsFiles.FileInfo, error)
	ReadWinningOutcomeDirectory(ctx context.Context) ([]lsFiles.FileInfo, error)
	ReadTeamDirectory(ctx context.Context) ([]lsFiles.FileInfoTeams, error)
}
