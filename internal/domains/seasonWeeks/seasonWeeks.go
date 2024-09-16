package seasonWeeks

import (
	"fmt"
	"time"
)

// | season_weeks | CREATE TABLE `season_weeks` (
// 	`season_week_id` int(11) NOT NULL,
// 	`season_id` int(11) NOT NULL,
// 	`week_number` smallint(3) NOT NULL,
// 	`status` enum('inactive','active','cancelled','finished') NOT NULL,
// 	`start_time` datetime NOT NULL,
// 	`end_time` datetime NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

// NewSeasonWeeks instantiate season week Struct
func NewSeasonWeeks(seasonID, weekNumber, status, startTime, endTime string) (*SeasonWeeks, error) {

	if seasonID == "" {
		return &SeasonWeeks{}, fmt.Errorf("seasonID not set")
	}

	if weekNumber == "" {
		return &SeasonWeeks{}, fmt.Errorf("weekNumber not set")
	}

	if status == "" {
		return &SeasonWeeks{}, fmt.Errorf("status alias not set")
	}

	if startTime == "" {
		return &SeasonWeeks{}, fmt.Errorf("startTime alias not set")
	}

	if endTime == "" {
		return &SeasonWeeks{}, fmt.Errorf("endTime alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &SeasonWeeks{
		SeasonID:   seasonID,
		WeekNumber: weekNumber,
		Status:     status,
		StartTime:  startTime,
		EndTime:    endTime,
		Created:    created,
		Modified:   modified,
	}, nil
}
