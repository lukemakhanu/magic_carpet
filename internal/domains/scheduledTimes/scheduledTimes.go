package scheduledTimes

import (
	"fmt"
	"time"
)

// NewScheduledTime instantiate scheduledTime Struct
func NewScheduledTime(scheduledTime, competitionID, status string) (*ScheduledTime, error) {

	if scheduledTime == "" {
		return &ScheduledTime{}, fmt.Errorf("scheduledTime not set")
	}

	if competitionID == "" {
		return &ScheduledTime{}, fmt.Errorf("competitionID not set")
	}

	if status == "" {
		return &ScheduledTime{}, fmt.Errorf("status not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &ScheduledTime{
		ScheduledTime: scheduledTime,
		CompetitionID: competitionID,
		Status:        status,
		Created:       created,
		Modified:      modified,
	}, nil
}
