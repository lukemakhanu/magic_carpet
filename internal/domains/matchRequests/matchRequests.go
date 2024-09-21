package matchRequests

import (
	"fmt"
	"time"
)

// ALTER TABLE `match_requests`
//   ADD KEY `match_request_id` (`match_request_id`),
//   ADD KEY `player_id` (`player_id`),
//   ADD KEY `start_time` (`start_time`),
//   ADD KEY `end_time` (`end_time`),
//   ADD KEY `early_finish` (`early_finish`),
//   ADD KEY `created` (`created`),
//   ADD KEY `played` (`played`),
//   ADD KEY `modified` (`modified`);

// NewMatchRequests instantiate players Struct
func NewMatchRequests(instantCompetitionID, playerID, startTime, endTime, earlyFinish, played string) (*MatchRequests, error) {

	if instantCompetitionID == "" {
		return &MatchRequests{}, fmt.Errorf("instantCompetitionID not set")
	}

	if playerID == "" {
		return &MatchRequests{}, fmt.Errorf("playerID not set")
	}

	if startTime == "" {
		return &MatchRequests{}, fmt.Errorf("startTime not set")
	}

	if endTime == "" {
		return &MatchRequests{}, fmt.Errorf("endTime not set")
	}

	if earlyFinish == "" {
		return &MatchRequests{}, fmt.Errorf("earlyFinish not set")
	}

	if played == "" {
		return &MatchRequests{}, fmt.Errorf("played not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &MatchRequests{
		InstantCompetitionID: instantCompetitionID,
		PlayerID:             playerID,
		StartTime:            startTime,
		EndTime:              endTime,
		EarlyFinish:          earlyFinish,
		Played:               played,
		Created:              created,
		Modified:             modified,
	}, nil
}
