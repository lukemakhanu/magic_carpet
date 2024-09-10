package lsFiles

import (
	"fmt"
	"time"
)

// NewLsFile instantiate
func NewLsFile(lsFileName, lsDir, country, lsExtID, projectID, competitionID string) (*LsFiles, error) {

	if lsFileName == "" {
		return &LsFiles{}, fmt.Errorf("lsFileName not set")
	}

	if lsDir == "" {
		return &LsFiles{}, fmt.Errorf("lsDir not set")
	}

	if country == "" {
		return &LsFiles{}, fmt.Errorf("country alias not set")
	}

	if lsExtID == "" {
		return &LsFiles{}, fmt.Errorf("lsExtID alias not set")
	}

	if projectID == "" {
		return &LsFiles{}, fmt.Errorf("projectID alias not set")
	}

	if competitionID == "" {
		return &LsFiles{}, fmt.Errorf("competitionID alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &LsFiles{
		LsFileName:    lsFileName,
		LsDir:         lsDir,
		LsExtID:       lsExtID,
		ProjectID:     projectID,
		CompetitionID: competitionID,
		Country:       country,
		Created:       created,
		Modified:      modified,
	}, nil
}
