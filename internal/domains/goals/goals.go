package goals

import (
	"fmt"
	"time"
)

// NewGoals instantiate goals Struct
func NewGoals(matchID, country, projectID, category string) (*Goals, error) {

	if matchID == "" {
		return &Goals{}, fmt.Errorf("matchID not set")
	}

	if country == "" {
		return &Goals{}, fmt.Errorf("country not set")
	}

	if projectID == "" {
		return &Goals{}, fmt.Errorf("projectID not set")
	}

	if category == "" {
		return &Goals{}, fmt.Errorf("category not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Goals{
		MatchID:   matchID,
		Country:   country,
		ProjectID: projectID,
		Category:  category,
		Created:   created,
		Modified:  modified,
	}, nil
}
