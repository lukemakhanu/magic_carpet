package leagues

import (
	"fmt"
	"time"
)

// NewLeagues instantiate league Struct
func NewLeagues(clientID, league, leagueAbbrv string) (*Leagues, error) {

	if clientID == "" {
		return &Leagues{}, fmt.Errorf("clientID not set")
	}

	if league == "" {
		return &Leagues{}, fmt.Errorf("league not set")
	}

	if leagueAbbrv == "" {
		return &Leagues{}, fmt.Errorf("league alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Leagues{
		ClientID:    clientID,
		League:      league,
		LeagueAbbrv: leagueAbbrv,
		Created:     created,
		Modified:    modified,
	}, nil
}
