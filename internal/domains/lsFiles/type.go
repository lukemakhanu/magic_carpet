package lsFiles

import "time"

// | ls_files | CREATE TABLE `ls_files` (
// 	`ls_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
// 	`ls_file_name` varchar(300) NOT NULL,
// 	`ls_dir` varchar(300) NOT NULL,
// 	`country` varchar(100) NOT NULL,
// 	`ls_ext_id` bigint(20) NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

type LsFiles struct {
	LsFileID      string
	LsFileName    string
	LsDir         string
	Country       string
	LsExtID       string
	ProjectID     string
	CompetitionID string
	Created       string
	Modified      string
}

type FileInfo struct {
	ExtID         string
	ProjectID     string
	CompetitionID string
	LsFileName    string
}

type LsData struct {
	RoundNumberID string      `json:"round_number_id"`
	ProjectID     string      `json:"project_id"`
	CompetitionID string      `json:"competition_id"`
	StartTime     string      `json:"start_time"`
	EndTime       string      `json:"end_time"`
	LsMarkets     []LsMarkets `json:"markets"`
}
type LsMarkets struct {
	HomeScore     int       `json:"home_score"`
	AwayScore     int       `json:"away_score"`
	ScoreTime     time.Time `json:"score_time"`
	ParentMatchID string    `json:"parent_match_id"`
	MinuteScored  string    `json:"minute_scored"`
}

type SingleLs struct {
	RoundNumberID string     `json:"round_number_id"`
	ParentMatchID string     `json:"parent_match_id"`
	SMarkets      []SMarkets `json:"markets"`
}

type SMarkets struct {
	HomeScore    int    `json:"home_score"`
	AwayScore    int    `json:"away_score"`
	MinuteScored string `json:"minute_scored"`
}

type FileInfoTeams struct {
	SeasonID      string
	CompetitionID string
	Count         string
	Name          string
}

type TeamsH2H struct {
	SeasonID      string `json:"season_id"`
	CompetitionID string `json:"competition_id"`
	Count         string `json:"count"`
	H2H           []H2H  `json:"h2h"`
}

type H2H struct {
	Teams    string `json:"teams"`
	MatchDay string `json:"match_day"`
}
