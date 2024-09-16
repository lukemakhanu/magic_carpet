package ssns

import (
	"fmt"
	"time"
)

// NewSsns instantiate season Struct
func NewSsns(leagueID, status string) (*Ssns, error) {

	if leagueID == "" {
		return &Ssns{}, fmt.Errorf("leagueID not set")
	}

	if status == "" {
		return &Ssns{}, fmt.Errorf("status not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Ssns{
		LeagueID: leagueID,
		Status:   status,
		Created:  created,
		Modified: modified,
	}, nil
}
