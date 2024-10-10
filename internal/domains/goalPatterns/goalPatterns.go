package goalPatterns

import (
	"fmt"
	"time"
)

// NewGoalPatterns instantiate goalPatterns
func NewGoalPatterns(seasonID, roundNumberID, competitionID string) (*GoalPatterns, error) {

	if seasonID == "" {
		return &GoalPatterns{}, fmt.Errorf("seasonID not set")
	}

	if roundNumberID == "" {
		return &GoalPatterns{}, fmt.Errorf("roundNumberID not set")
	}

	if competitionID == "" {
		return &GoalPatterns{}, fmt.Errorf("competitionID alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &GoalPatterns{
		SeasonID:      seasonID,
		RoundNumberID: roundNumberID,
		CompetitionID: competitionID,
		Created:       created,
		Modified:      modified,
	}, nil
}
