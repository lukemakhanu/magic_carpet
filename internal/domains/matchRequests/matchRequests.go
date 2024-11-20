package matchRequests

import (
	"fmt"
	"time"
)

// NewMatchRequests instantiate players Struct
func NewMatchRequests(playerID string) (*MatchRequests, error) {

	if playerID == "" {
		return &MatchRequests{}, fmt.Errorf("playerID not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &MatchRequests{
		PlayerID: playerID,
		Created:  created,
		Modified: modified,
	}, nil
}
