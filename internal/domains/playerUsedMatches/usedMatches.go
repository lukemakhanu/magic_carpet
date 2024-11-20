package playerUsedMatches

import (
	"fmt"
	"time"
)

// NewPlayerUsedMatches instantiate used matches Struct
func NewPlayerUsedMatches(playerID, country, projectID, matchID, category string) (*PlayerUsedMatches, error) {

	if playerID == "" {
		return &PlayerUsedMatches{}, fmt.Errorf("country not set")
	}

	if country == "" {
		return &PlayerUsedMatches{}, fmt.Errorf("country not set")
	}

	if projectID == "" {
		return &PlayerUsedMatches{}, fmt.Errorf("projectID not set")
	}

	if matchID == "" {
		return &PlayerUsedMatches{}, fmt.Errorf("matchID not set")
	}

	if category == "" {
		return &PlayerUsedMatches{}, fmt.Errorf("category not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &PlayerUsedMatches{
		PlayerID:  playerID,
		Country:   country,
		ProjectID: projectID,
		MatchID:   matchID,
		Category:  category,
		Created:   created,
		Modified:  modified,
	}, nil
}
