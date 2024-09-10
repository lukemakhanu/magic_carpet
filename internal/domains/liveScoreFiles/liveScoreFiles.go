package liveScoreFiles

import (
	"fmt"
	"time"
)

// NewLiveScoreFile instantiate
func NewLiveScoreFile(lsFileName, lsDir, country, extID, projectID, competitionID string) (*LiveScoreFiles, error) {

	if lsFileName == "" {
		return &LiveScoreFiles{}, fmt.Errorf("lsFileName not set")
	}

	if lsDir == "" {
		return &LiveScoreFiles{}, fmt.Errorf("lsDir not set")
	}

	if country == "" {
		return &LiveScoreFiles{}, fmt.Errorf("country not set")
	}

	if extID == "" {
		return &LiveScoreFiles{}, fmt.Errorf("extID not set")
	}

	if projectID == "" {
		return &LiveScoreFiles{}, fmt.Errorf("projectID not set")
	}

	if competitionID == "" {
		return &LiveScoreFiles{}, fmt.Errorf("competitionID not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &LiveScoreFiles{
		LsFileName:    lsFileName,
		LsDir:         lsDir,
		ExtID:         extID,
		ProjectID:     projectID,
		CompetitionID: competitionID,
		Country:       country,
		Created:       created,
		Modified:      modified,
	}, nil
}
