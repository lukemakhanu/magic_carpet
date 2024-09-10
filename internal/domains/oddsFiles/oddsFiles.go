package oddsFiles

import (
	"fmt"
	"time"
)

// NewOddsFile instantiate oddsFile Struct
func NewOddsFile(oddsFileName, fileDirectory, country, parentID, competitionID, matchID string) (*OddsFiles, error) {

	if oddsFileName == "" {
		return &OddsFiles{}, fmt.Errorf("oddsFileName not set")
	}

	if fileDirectory == "" {
		return &OddsFiles{}, fmt.Errorf("fileDirectory not set")
	}

	if country == "" {
		return &OddsFiles{}, fmt.Errorf("country not set")
	}

	if parentID == "" {
		return &OddsFiles{}, fmt.Errorf("parentID not set")
	}

	if competitionID == "" {
		return &OddsFiles{}, fmt.Errorf("competitionID not set")
	}

	if matchID == "" {
		return &OddsFiles{}, fmt.Errorf("matchID not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &OddsFiles{
		OddsFileName:  oddsFileName,
		FileDirectory: fileDirectory,
		Country:       country,
		ParentID:      parentID,
		CompetitionID: competitionID,
		MatchID:       matchID,
		Created:       created,
		Modified:      modified,
	}, nil
}
