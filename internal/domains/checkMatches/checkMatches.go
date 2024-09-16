package checkMatches

import (
	"fmt"
	"time"
)

// NewCheckMatches instantiate check match Struct
func NewCheckMatches(country, matchID, matchDate string) (*CheckMatches, error) {

	if country == "" {
		return &CheckMatches{}, fmt.Errorf("country not set")
	}

	if matchID == "" {
		return &CheckMatches{}, fmt.Errorf("matchID not set")
	}

	if matchDate == "" {
		return &CheckMatches{}, fmt.Errorf("matchDate alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &CheckMatches{
		Country:   country,
		MatchID:   matchID,
		MatchDate: matchDate,
		Created:   created,
		Modified:  modified,
	}, nil
}
