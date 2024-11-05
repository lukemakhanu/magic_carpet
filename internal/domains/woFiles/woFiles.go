package woFiles

import (
	"fmt"
	"time"
)

// NewWoFile instantiate
func NewWoFile(woFileName, woDir, country, woExtID, projectID, competitionID, status string) (*WoFiles, error) {

	if woFileName == "" {
		return &WoFiles{}, fmt.Errorf("woFileName not set")
	}

	if woDir == "" {
		return &WoFiles{}, fmt.Errorf("woDir not set")
	}

	if country == "" {
		return &WoFiles{}, fmt.Errorf("country alias not set")
	}

	if woExtID == "" {
		return &WoFiles{}, fmt.Errorf("woExtID alias not set")
	}

	if projectID == "" {
		return &WoFiles{}, fmt.Errorf("projectID alias not set")
	}

	if competitionID == "" {
		return &WoFiles{}, fmt.Errorf("competitionID alias not set")
	}

	if status == "" {
		return &WoFiles{}, fmt.Errorf("status not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &WoFiles{
		WoFileName:    woFileName,
		WoDir:         woDir,
		WoExtID:       woExtID,
		ProjectID:     projectID,
		CompetitionID: competitionID,
		Country:       country,
		Status:        status,
		Created:       created,
		Modified:      modified,
	}, nil
}
