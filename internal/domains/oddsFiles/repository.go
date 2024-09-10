package oddsFiles

import "context"

// OddsFilesRepository contains methods that implements odds Files struct
type OddsFilesRepository interface {
	Save(ctx context.Context, t OddsFiles) (int, error)
	SaveOdd(ctx context.Context, t OddsFiles) (int, error)
	GetOddsFiles(ctx context.Context) ([]OddsFiles, error)
	GetOddsFileByID(ctx context.Context, oddsFileID string) ([]OddsFiles, error)
	UpdateMatch(ctx context.Context, oddsFileName, fileDirectory, country, parentID, competitionID, matchID, oddsFileID string) (int64, error)
	GetOddsFileWithLimit(ctx context.Context, min, max string) ([]OddsFiles, error)
	GetOddsParentID(ctx context.Context, parentID string) ([]OddsFiles, error)
	SaveMatchOdds(ctx context.Context, roundID, parentID, country string) (int, error)
	GetReadyMatchCount(ctx context.Context) ([]MatchDet, error)

	GetAllOddsParentID(ctx context.Context, parentID, countryCode string) ([]OddsFiles, error)
}
