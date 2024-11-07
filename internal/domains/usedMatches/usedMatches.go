package usedMatches

import (
	"fmt"
	"time"
)

// NewUsedMatches instantiate used matches Struct
func NewUsedMatches(country, projectID, matchID, category string) (*UsedMatches, error) {

	if country == "" {
		return &UsedMatches{}, fmt.Errorf("country not set")
	}

	if projectID == "" {
		return &UsedMatches{}, fmt.Errorf("projectID not set")
	}

	if matchID == "" {
		return &UsedMatches{}, fmt.Errorf("matchID not set")
	}

	if category == "" {
		return &UsedMatches{}, fmt.Errorf("category not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &UsedMatches{
		Country:   country,
		ProjectID: projectID,
		MatchID:   matchID,
		Category:  category,
		Created:   created,
		Modified:  modified,
	}, nil
}
