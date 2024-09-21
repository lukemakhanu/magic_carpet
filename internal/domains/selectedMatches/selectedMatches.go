package selectedMatches

import (
	"fmt"
	"time"
)

/* CREATE TABLE `selected_matches` (
`selected_matches_id` bigint(40) NOT NULL,
`player_id` bigint(30) NOT NULL,
`match_request_id` bigint(40) NOT NULL,
`parent_match_id` varchar(50) NOT NULL,
`created` datetime NOT NULL,
`modified` datetime NOT NULL */

// NewSelectedMatches instantiate selectedMatches Struct
func NewSelectedMatches(playerID, matchRequestID, parentMatchID string) (*SelectedMatches, error) {

	if playerID == "" {
		return &SelectedMatches{}, fmt.Errorf("playerID not set")
	}

	if matchRequestID == "" {
		return &SelectedMatches{}, fmt.Errorf("matchRequestID not set")
	}

	if parentMatchID == "" {
		return &SelectedMatches{}, fmt.Errorf("parentMatchID not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &SelectedMatches{
		PlayerID:       playerID,
		MatchRequestID: matchRequestID,
		ParentMatchID:  parentMatchID,
		Created:        created,
		Modified:       modified,
	}, nil
}
