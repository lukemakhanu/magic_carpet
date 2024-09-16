package matches

import (
	"fmt"
	"time"
)

// | matches | CREATE TABLE `matches` (
// 	`match_id` bigint(20) NOT NULL AUTO_INCREMENT,
// 	`season_week_id` int(11) NOT NULL,
// 	`home_team_id` smallint(3) NOT NULL,
// 	`away_team_id` smallint(3) NOT NULL,
// 	`status` enum('inactive','active','cancelled','finished') NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

// NewMatches instantiate matches Struct
func NewMatches(seasonWeekID, homeTeamID, awayTeamID, status string) (*Matches, error) {

	if seasonWeekID == "" {
		return &Matches{}, fmt.Errorf("seasonWeekID not set")
	}

	if homeTeamID == "" {
		return &Matches{}, fmt.Errorf("homeTeamID not set")
	}

	if awayTeamID == "" {
		return &Matches{}, fmt.Errorf("awayTeamID alias not set")
	}

	if status == "" {
		return &Matches{}, fmt.Errorf("status alias not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Matches{
		SeasonWeekID: seasonWeekID,
		HomeTeamID:   homeTeamID,
		AwayTeamID:   awayTeamID,
		Status:       status,
		Created:      created,
		Modified:     modified,
	}, nil
}
