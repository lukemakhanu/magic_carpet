package selectedMatches

import (
	"fmt"
	"time"
)

// NewSelectedMatches instantiate selectedMatches Struct
func NewSelectedMatches(playerID, periodID, parentMatchID, homeTeamID, awayTeamID, status string) (*SelectedMatches, error) {

	if playerID == "" {
		return &SelectedMatches{}, fmt.Errorf("playerID not set")
	}

	if periodID == "" {
		return &SelectedMatches{}, fmt.Errorf("periodID not set")
	}

	if parentMatchID == "" {
		return &SelectedMatches{}, fmt.Errorf("parentMatchID not set")
	}

	if homeTeamID == "" {
		return &SelectedMatches{}, fmt.Errorf("homeTeamID not set")
	}

	if awayTeamID == "" {
		return &SelectedMatches{}, fmt.Errorf("awayTeamID not set")
	}

	if status == "" {
		return &SelectedMatches{}, fmt.Errorf("status not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &SelectedMatches{
		PlayerID:      playerID,
		PeriodID:      periodID,
		ParentMatchID: parentMatchID,
		HomeTeamID:    homeTeamID,
		AwayTeamID:    awayTeamID,
		Status:        status,
		Created:       created,
		Modified:      modified,
	}, nil
}
