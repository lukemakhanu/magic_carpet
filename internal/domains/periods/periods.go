package periods

import (
	"fmt"
	"time"
)

func NewPeriods(competitionID, matchRequestID, startTime, endTime, earlyFinish, played, gameStarted, keyCreated, roundNumberID string) (*Periods, error) {

	if competitionID == "" {
		return &Periods{}, fmt.Errorf("competitionID not set")
	}

	if matchRequestID == "" {
		return &Periods{}, fmt.Errorf("matchRequestID not set")
	}

	if startTime == "" {
		return &Periods{}, fmt.Errorf("startTime not set")
	}

	if endTime == "" {
		return &Periods{}, fmt.Errorf("endTime not set")
	}

	if earlyFinish == "" {
		return &Periods{}, fmt.Errorf("earlyFinish not set")
	}

	if played == "" {
		return &Periods{}, fmt.Errorf("played not set")
	}

	if gameStarted == "" {
		return &Periods{}, fmt.Errorf("gameStarted not set")
	}

	if keyCreated == "" {
		return &Periods{}, fmt.Errorf("keyCreated not set")
	}

	if roundNumberID == "" {
		return &Periods{}, fmt.Errorf("roundNumberID not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Periods{
		CompetitionID:  competitionID,
		MatchRequestID: matchRequestID,
		StartTime:      startTime,
		EndTime:        endTime,
		EarlyFinish:    earlyFinish,
		Played:         played,
		GameStarted:    gameStarted,
		KeyCreated:     keyCreated,
		RoundNumberID:  roundNumberID,
		Created:        created,
		Modified:       modified,
	}, nil
}
