package snwkpts

import (
	"fmt"
	"time"
)

// NewSnWkPts
func NewSnWkPts(seasonWeekID, roundNumberID, competitionID string) (*SnWkPts, error) {

	if seasonWeekID == "" {
		return &SnWkPts{}, fmt.Errorf("seasonWeekID not set")
	}

	if roundNumberID == "" {
		return &SnWkPts{}, fmt.Errorf("roundNumberID not set")
	}

	if competitionID == "" {
		return &SnWkPts{}, fmt.Errorf("competitionID alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &SnWkPts{
		SeasonWeekID:  seasonWeekID,
		RoundNumberID: roundNumberID,
		CompetitionID: competitionID,
		Created:       created,
		Modified:      modified,
	}, nil
}
