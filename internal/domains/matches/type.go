// Copyright 2023 lukemakhanu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matches

// | matches | CREATE TABLE `matches` (
// 	`match_id` bigint(20) NOT NULL AUTO_INCREMENT,
// 	`season_week_id` int(11) NOT NULL,
// 	`home_team_id` smallint(3) NOT NULL,
// 	`away_team_id` smallint(3) NOT NULL,
// 	`status` enum('inactive','active','cancelled','finished') NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

type Matches struct {
	MatchID      string
	SeasonWeekID string
	HomeTeamID   string
	AwayTeamID   string
	Status       string
	Created      string
	Modified     string
}

type MatchGames struct {
	MatchID      string
	LeagueID     string
	SeasonID     string
	SeasonWeekID string
	WeekNumber   string
	HomeTeamID   string
	AwayTeamID   string
	Status       string
	Created      string
	Modified     string
}
