package ssnsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/ssns"
)

var _ ssns.SsnsRepository = (*MysqlRepository)(nil)

type MysqlRepository struct {
	db *sql.DB
}

// Create a new mysql repository
func New(connectionString string) (*MysqlRepository, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(5)
	// Maximum Open Connections
	db.SetMaxOpenConns(10)
	// Idle Connection Timeout
	db.SetConnMaxIdleTime(5 * time.Second)
	// Connection Lifetime
	db.SetConnMaxLifetime(15 * time.Second)

	return &MysqlRepository{
		db: db,
	}, nil
}

// Save :
func (mr *MysqlRepository) Save(ctx context.Context, t ssns.Ssns) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT sns SET league_id=?,status=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.LeagueID, t.Status)

	if err != nil {
		return d, fmt.Errorf("Unable to save sns : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last ssn ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetSeasons(ctx context.Context, status string) ([]ssns.Ssns, error) {
	var gc []ssns.Ssns
	statement := fmt.Sprintf("select season_id,league_id,status, created,modified from sns where status = '%s' ", status)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g ssns.Ssns
		err := raws.Scan(&g.SeasonID, &g.LeagueID, &g.Created, &g.Modified)
		if err != nil {
			return nil, err
		}
		gc = append(gc, g)
	}

	if err = raws.Err(); err != nil {
		return nil, err
	}
	raws.Close()

	return gc, nil
}

func (r *MysqlRepository) GetSeasonByID(ctx context.Context, seasonID string) ([]ssns.Ssns, error) {
	var gc []ssns.Ssns
	statement := fmt.Sprintf("select season_id,league_id,status,created,modified from sns where season_id = '%s' ",
		seasonID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g ssns.Ssns
		err := raws.Scan(&g.SeasonID, &g.LeagueID, &g.Status, &g.Created, &g.Modified)
		if err != nil {
			return nil, err
		}
		gc = append(gc, g)
	}

	if err = raws.Err(); err != nil {
		return nil, err
	}
	raws.Close()

	return gc, nil
}

// UpdateLeague : updates league
func (mr *MysqlRepository) UpdateSeason(ctx context.Context, leagueID, status, seasonID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update sns set league_id=?, status=? where season_id = ? ",
		leagueID, status, seasonID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// CREATE TABLE `ssn_weeks` (
// 	`season_week_id` int(11) NOT NULL AUTO_INCREMENT,
// 	`league_id` smallint(4) NOT NULL,
// 	`season_id` int(11) NOT NULL,
// 	`week_number` smallint(3) NOT NULL,
// 	`status` enum('inactive','active','cancelled','finished') NOT NULL,
// 	`start_time` datetime NOT NULL,
// 	`end_time` datetime NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

func (mr *MysqlRepository) SaveSsnWeek(ctx context.Context, leagueID, seasonID, weekNumber, status, startTime, endTime string) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT sn_wks SET league_id=?,season_id=?,week_number=?,status=?, \n"+
		"start_time=?,end_time=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		leagueID, seasonID, weekNumber, status, startTime, endTime)

	if err != nil {
		return d, fmt.Errorf("Unable to save sns week : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last ssn week ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (mr *MysqlRepository) SaveGames(ctx context.Context, leagueID, seasonID, seasonWeekID, weekNumber, homeTeamID, awayTeamID, status string) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT gms SET league_id=?,season_id=?,season_week_id=?,week_number=?, \n"+
		"home_team_id=?,away_team_id=?,status=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		leagueID, seasonID, seasonWeekID, weekNumber, homeTeamID, awayTeamID, status)

	if err != nil {
		return d, fmt.Errorf("Unable to save games : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last game ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetLastGame(ctx context.Context, leagueID string) ([]ssns.LastGameTime, error) {
	var gc []ssns.LastGameTime
	statement := fmt.Sprintf("select start_time from sn_wks where league_id = '%s' order by start_time desc limit 1 ",
		leagueID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g ssns.LastGameTime
		err := raws.Scan(&g.StartTime)
		if err != nil {
			return nil, err
		}
		gc = append(gc, g)
	}

	if err = raws.Err(); err != nil {
		return nil, err
	}
	raws.Close()

	return gc, nil
}

func (mr *MysqlRepository) CountRemainingPeriods(ctx context.Context, leagueID string) (int, error) {

	var count int

	statement := fmt.Sprintf("select count(season_week_id) from sn_wks where league_id='%s' and start_time > now() ", leagueID)

	row := mr.db.QueryRow(statement)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows:
		return count, nil
	case nil:
		return count, nil
	default:
		return count, fmt.Errorf("Unable to return sn_wks count :: %v", err)
	}
}
