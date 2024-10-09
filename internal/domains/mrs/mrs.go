package mrs

import (
	"fmt"
	"time"
)

func NewMrs(roundNumberID int, totalGoals, goalCount, competitionID, startTime string) (*Mrs, error) {

	if roundNumberID <= 0 {
		return &Mrs{}, fmt.Errorf("roundNumberID not set")
	}

	if totalGoals == "" {
		return &Mrs{}, fmt.Errorf("totalGoals not set")
	}

	if goalCount == "" {
		return &Mrs{}, fmt.Errorf("goalCount not set")
	}

	if competitionID == "" {
		return &Mrs{}, fmt.Errorf("competitionID not set")
	}

	if startTime == "" {
		return &Mrs{}, fmt.Errorf("startTime not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Mrs{
		RoundNumberID: roundNumberID,
		TotalGoals:    totalGoals,
		GoalCount:     goalCount,
		CompetitionID: competitionID,
		StartTime:     startTime,
		Created:       created,
		Modified:      modified,
	}, nil

}
