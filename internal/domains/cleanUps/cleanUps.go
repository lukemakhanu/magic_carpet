package cleanUps

import (
	"fmt"
	"time"
)

// NewCleanUps instantiate cleanUps
func NewCleanUps(projectID, status string) (*CleanUps, error) {

	if projectID == "" {
		return &CleanUps{}, fmt.Errorf("projectID not set")
	}

	if status == "" {
		return &CleanUps{}, fmt.Errorf("status alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &CleanUps{
		ProjectID: projectID,
		Status:    status,
		Created:   created,
		Modified:  modified,
	}, nil
}
