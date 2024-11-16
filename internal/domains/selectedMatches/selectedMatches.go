package selectedMatches

import (
	"fmt"
	"time"
)

/* CREATE TABLE `selected_matches` (
`selected_matches_id` bigint(40) NOT NULL,
`player_id` bigint(30) NOT NULL,
`period_id` bigint(40) NOT NULL,
`parent_match_id` varchar(50) NOT NULL,
`created` datetime NOT NULL,
`modified` datetime NOT NULL */

// NewSelectedMatches instantiate selectedMatches Struct
func NewSelectedMatches(playerID, periodID, parentMatchID string) (*SelectedMatches, error) {

	if playerID == "" {
		return &SelectedMatches{}, fmt.Errorf("playerID not set")
	}

	if periodID == "" {
		return &SelectedMatches{}, fmt.Errorf("periodID not set")
	}

	if parentMatchID == "" {
		return &SelectedMatches{}, fmt.Errorf("parentMatchID not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &SelectedMatches{
		PlayerID:      playerID,
		PeriodID:      periodID,
		ParentMatchID: parentMatchID,
		Created:       created,
		Modified:      modified,
	}, nil
}
