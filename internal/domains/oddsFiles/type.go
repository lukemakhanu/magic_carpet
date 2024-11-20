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

package oddsFiles

// | odds_file | CREATE TABLE `odds_file` (
// 	`odds_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
// 	`odds_file_name` varchar(300) NOT NULL,
// 	`file_directory` varchar(300) NOT NULL,
// 	`country` varchar(50) NOT NULL,
// 	`parent_id` varchar(30) NOT NULL,
// 	`competition_id` smallint(6) NOT NULL,
// 	`match_id` int(11) NOT NULL DEFAULT '0',
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

type OddsFiles struct {
	OddsFileID    string
	OddsFileName  string
	FileDirectory string
	Country       string
	ParentID      string
	CompetitionID string
	MatchID       string
	Created       string
	Modified      string
}

type MatchOdds struct {
	MatchOddsID string
	RoundID     string
	ParentID    string
	Country     string
	Created     string
	Modified    string
}

type MatchDet struct {
	MatchCount int
	RoundID    string
}

type FinalSeasonWeek struct {
	SeasonID     string         `json:"season_id"`
	SeasonWeeKID string         `json:"season_week_id"`
	MatchDay     string         `json:"match_day"`
	StartTime    string         `json:"start_time"`
	EndTime      string         `json:"end_time"`
	FinalMatches []FinalMatches `json:"matches"`
}

type FinalMatches struct {
	MatchID      string         `json:"match_id"`
	HomeID       string         `json:"home_id"`
	HomeAlias    string         `json:"home_alias"`
	HomeTeam     string         `json:"home_team"`
	AwayID       string         `json:"away_id"`
	AwayAlias    string         `json:"away_alias"`
	AwayTeam     string         `json:"away_team"`
	FinalMarkets []FinalMarkets `json:"markets"`
}

type FinalMarkets struct {
	Name          string          `json:"name"`
	Code          string          `json:"code"`
	FinalOutcomes []FinalOutcomes `json:"outcomes"`
}

type FinalOutcomes struct {
	OutcomeID    string  `json:"outcome_id"`
	OutcomeName  string  `json:"outcome_name"`
	OddValue     float64 `json:"odd_value"`
	OutcomeAlias string  `json:"outcome_alias"`
}

// RawOdds : used to parse raw winning outcomes
type RawOdds struct {
	ParentMatchID string       `json:"parent_match_id"`
	ProjectID     string       `json:"project_id"`
	MatchID       string       `json:"match_id"`
	RawMarkets    []RawMarkets `json:"markets"`
}

type RawMarkets struct {
	Name        string        `json:"name"`
	SubTypeID   string        `json:"sub_type_id"`
	RawOutcomes []RawOutcomes `json:"outcomes"`
}

type RawOutcomes struct {
	OutcomeID    string `json:"outcome_id"`
	OutcomeName  string `json:"outcome_name"`
	OddValue     string `json:"odd_value"`
	OutcomeAlias string `json:"outcome_alias"`
}

// FinalSeasonWeekWO : used to formulate winning outcomes
type FinalSeasonWeekWO struct {
	SeasonID       string           `json:"season_id"`
	SeasonWeeKID   string           `json:"season_week_id"`
	MatchDay       string           `json:"match_day"`
	StartTime      string           `json:"start_time"`
	EndTime        string           `json:"end_time"`
	FinalMatchesWO []FinalMatchesWO `json:"matches"`
}

type FinalMatchesWO struct {
	MatchID    string      `json:"match_id"`
	HomeID     string      `json:"home_id"`
	HomeAlias  string      `json:"home_alias"`
	HomeTeam   string      `json:"home_team"`
	AwayID     string      `json:"away_id"`
	AwayAlias  string      `json:"away_alias"`
	AwayTeam   string      `json:"away_team"`
	FinalScore FinalScores `json:"final_score"`
}

type FinalScores struct {
	HomeScore            string                 `json:"home_score"`
	AwayScore            string                 `json:"away_score"`
	FinalWinningOutcomes []FinalWinningOutcomes `json:"outcomes"`
}

type FinalWinningOutcomes struct {
	SubTypeID   string `json:"sub_type_id"`
	OutcomeID   string `json:"outcome_id"`
	OutcomeName string `json:"outcome_name"`
	Result      string `json:"result"`
}

type FinalSeasonWeekLS struct {
	SeasonID       string           `json:"season_id"`
	SeasonWeeKID   string           `json:"season_week_id"`
	MatchDay       string           `json:"match_day"`
	StartTime      string           `json:"start_time"`
	EndTime        string           `json:"end_time"`
	FinalMatchesLS []FinalMatchesLS `json:"matches"`
}

// FinalMatchesLS : used to store live score
type FinalMatchesLS struct {
	MatchID         string            `json:"match_id"`
	HomeID          string            `json:"home_id"`
	HomeAlias       string            `json:"home_alias"`
	HomeTeam        string            `json:"home_team"`
	AwayID          string            `json:"away_id"`
	AwayAlias       string            `json:"away_alias"`
	AwayTeam        string            `json:"away_team"`
	FinalLiveScores []FinalLiveScores `json:"live_scores"`
}

type FinalLiveScores struct {
	HomeScore    int    `json:"home_score"`
	AwayScore    int    `json:"away_score"`
	MinuteScored string `json:"minute_scored"`
}

// RawWinningOutcomes : raw winning outcomes
type RawWinningOutcomes struct {
	RoundNumberID string   `json:"round_number_id"`
	HomeScore     string   `json:"home_score"`
	AwayScore     string   `json:"away_score"`
	RawWOs        []RawWOs `json:"winning_outcomes"`
}

// RawWOs :
type RawWOs struct {
	ParentMatchID string `json:"parent_match_id"`
	SubTypeID     string `json:"sub_type_id"`
	OutcomeID     string `json:"outcome_id"`
	OutcomeName   string `json:"outcome_name"`
	Result        string `json:"result"`
}

type RawLS struct {
	RoundNumberID string  `json:"round_number_id"`
	ParentMatchID string  `json:"parent_match_id"`
	Goals         []Goals `json:"markets"`
}

type Goals struct {
	HomeScore    int    `json:"home_score"`
	AwayScore    int    `json:"away_score"`
	MinuteScored string `json:"minute_scored"`
}

type MatchAPI struct {
	StatusCode        string       `json:"status_code"`
	StatusDescription string       `json:"status_description"`
	MatchDetails      MatchDetails `json:"match_details"`
}

type MatchDetails struct {
	MatchDate   string      `json:"match_date"`
	League      string      `json:"league"`
	MatchSeason string      `json:"match_season"`
	MatchDay    interface{} `json:"match_day"`
}

// WoAPI : winning outcome api
type WoAPI struct {
	StatusCode        string      `json:"status_code"`
	StatusDescription string      `json:"status_description"`
	Results           interface{} `json:"outcome_details"`
}

// LsAPI : winning outcome api
type LsAPI struct {
	StatusCode        string      `json:"status_code"`
	StatusDescription string      `json:"status_description"`
	LiveScore         interface{} `json:"live_scores"`
}

type CheckKeys struct {
	OddsKey      string
	ValidateKeys ValidateKeys
}

type ValidateKeys struct {
	Odds string
	Wo   string
	Ls   string
}

type FinalPeriod struct {
	PeriodID     string         `json:"period_id"`
	PlayerID     string         `json:"player_id"`
	MatchDay     string         `json:"match_day"`
	StartTime    string         `json:"start_time"`
	EndTime      string         `json:"end_time"`
	FinalMatches []FinalMatches `json:"matches"`
}
