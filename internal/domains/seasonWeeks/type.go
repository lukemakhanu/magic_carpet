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

package seasonWeeks

// | season_weeks | CREATE TABLE `season_weeks` (
// 	`season_week_id` int(11) NOT NULL,
// 	`season_id` int(11) NOT NULL,
// 	`week_number` smallint(3) NOT NULL,
// 	`status` enum('inactive','active','cancelled','finished') NOT NULL,
// 	`start_time` datetime NOT NULL,
// 	`end_time` datetime NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

type SeasonWeeks struct {
	SeasonWeekID string
	SeasonID     string
	WeekNumber   string
	Status       string
	StartTime    string
	EndTime      string
	Created      string
	Modified     string
}

type SeasonWeekDetails struct {
	SeasonWeekID string
	SeasonID     string
	LeagueID     string
	WeekNumber   string
	Status       string
	StartTime    string
	EndTime      string
	Created      string
	Modified     string
}

type SeasonWkDetails struct {
	SeasonWeekID  string
	SeasonID      string
	LeagueID      string
	WeekNumber    string
	Status        string
	StartTime     string
	EndTime       string
	CompetitionID string
	RoundNumberID string
	Created       string
	Modified      string
}

// SeasonWeeksAPI :
type SeasonWeeksAPI struct {
	SeasonWeekID string
	SeasonID     string
	WeekNumber   string
	Status       string
	StartTime    string
	EndTime      string
	ApiDate      string
	Created      string
	Modified     string
}

// ProductionSeasonWeeksAPI :
type ProductionSeasonWeeksAPI struct {
	SeasonWeekID string
	LeagueID     string
	SeasonID     string
	WeekNumber   string
	Status       string
	StartTime    string
	EndTime      string
	ApiDate      string
	Created      string
	Modified     string
}
