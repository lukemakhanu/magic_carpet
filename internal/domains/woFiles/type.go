// Copyright 2023 lukemakhanu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package woFiles

type WoFiles struct {
	WoFileID      string
	WoFileName    string
	WoDir         string
	Country       string
	WoExtID       string
	ProjectID     string
	CompetitionID string
	Created       string
	Modified      string
}

type RawWinningOutcome struct {
	RoundNumberID   string            `json:"round_number_id"`
	ProjectID       string            `json:"project_id"`
	CompetitionID   string            `json:"competition_id"`
	StartTime       string            `json:"start_time"`
	EndTime         string            `json:"end_time"`
	SportType       string            `json:"sport_type"`
	WinningOutcomes []WinningOutcomes `json:"winning_outcomes"`
	Results         []Results         `json:"results"`
}

type WinningOutcomes struct {
	ParentMatchID string `json:"parent_match_id"`
	SubTypeID     string `json:"sub_type_id"`
	OutcomeID     string `json:"outcome_id"`
	OutcomeName   string `json:"outcome_name"`
	Result        string `json:"result"`
}

type Results struct {
	ParentMatchID string `json:"parent_match_id"`
	HomeTeamID    string `json:"home_team_id"`
	AwayTeamID    string `json:"away_team_id"`
	HomeScore     string `json:"home_score"`
	AwayScore     string `json:"away_score"`
}

type MatchWO struct {
	RoundNumberID    string             `json:"round_number_id"`
	HomeScore        string             `json:"home_score"`
	AwayScore        string             `json:"away_score"`
	MWinningOutcomes []MWinningOutcomes `json:"winning_outcomes"`
}

type MWinningOutcomes struct {
	ParentMatchID string `json:"parent_match_id"`
	SubTypeID     string `json:"sub_type_id"`
	OutcomeID     string `json:"outcome_id"`
	OutcomeName   string `json:"outcome_name"`
	Result        string `json:"result"`
}

type WinningOutcomeFiles struct {
	WoFileID      string
	WoFileName    string
	WoDir         string
	Country       string
	ExtID         string
	ProjectID     string
	CompetitionID string
	Created       string
	Modified      string
}
