package leaguesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/leagues"
)

var _ leagues.LeaguesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t leagues.Leagues) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT leagues SET client_id=?,league=?,league_abbrv=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.ClientID, t.League, t.LeagueAbbrv)

	if err != nil {
		return d, fmt.Errorf("unable to save leagues : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last league ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetLeagues(ctx context.Context) ([]leagues.Leagues, error) {
	var gc []leagues.Leagues
	statement := fmt.Sprintf("select league_id,client_id,league,league_abbrv,created,modified from leagues ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.Leagues
		err := raws.Scan(&g.LeagueID, &g.ClientID, &g.League, &g.LeagueAbbrv, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetLeagueByID(ctx context.Context, leagueID string) ([]leagues.Leagues, error) {
	var gc []leagues.Leagues
	statement := fmt.Sprintf("select league_id,client_id,league,league_abbrv,created,modified from leagues where league_id = '%s' ",
		leagueID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.Leagues
		err := raws.Scan(&g.LeagueID, &g.ClientID, &g.League, &g.LeagueAbbrv, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetLeagueByClientID(ctx context.Context, clientID string) ([]leagues.Leagues, error) {
	var gc []leagues.Leagues
	statement := fmt.Sprintf("select league_id,client_id,league,league_abbrv,created,modified from leagues where client_id = '%s' ",
		clientID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.Leagues
		err := raws.Scan(&g.LeagueID, &g.ClientID, &g.League, &g.LeagueAbbrv, &g.Created, &g.Modified)
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
func (mr *MysqlRepository) UpdateLeague(ctx context.Context, clientID, league, leagueAbbrv, leagueID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update leagues set client_id=?,league=?,league_abbrv=? where league_id = ? ",
		clientID, league, leagueAbbrv, leagueID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// GetMatchDetails : returns match details used by API.
func (r *MysqlRepository) GetMatchDetails(ctx context.Context, leagueAbbr string) ([]leagues.MatchDetails, error) {
	var gc []leagues.MatchDetails
	statement := fmt.Sprintf("select l.league_id,l.client_id,l.league,l.league_abbrv,date(now()),s.season_id from leagues as l \n"+
		"inner join seasons as s on s.league_id = l.league_id \n"+
		"where l.league_abbrv = '%s' order by s.season_id desc limit 1  ",
		//"where md.date = date(now()) and l.league_abbrv = '%s' order by s.season_id asc  ",
		leagueAbbr)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.MatchDetails
		err := raws.Scan(&g.LeagueID, &g.ClientID, &g.League, &g.LeagueAbbrv, &g.MatchDate, &g.SeasonID)
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

// GetWinningOutcomes : returns winning outcomes used by API.
func (r *MysqlRepository) GetWinningOutcomes(ctx context.Context, seasonWeekID string) ([]leagues.MatchWinningOutcomeAPI, error) {
	var gc []leagues.MatchWinningOutcomeAPI
	statement := fmt.Sprintf("select season_week_id,week_number,status,start_time,end_time from season_weeks \n"+
		"where season_week_id = '%s' and start_time < now() + interval 10 second   ", seasonWeekID)
	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.MatchWinningOutcomeAPI
		err := raws.Scan(&g.SeasonWeekID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime)
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

// GetProductionMatchDetails : returns production match details used by API.
func (r *MysqlRepository) GetProductionMatchDetails(ctx context.Context, leagueAbbr string) ([]leagues.MatchDetails, error) {
	var gc []leagues.MatchDetails
	statement := fmt.Sprintf("select l.league_id,l.client_id,l.league,l.league_abbrv,date(now()),s.season_id from leagues as l \n"+
		"inner join ssns as s on s.league_id = l.league_id \n"+
		"where l.league_abbrv = '%s' order by s.season_id desc limit 1  ",
		//"where md.date = date(now()) and l.league_abbrv = '%s' order by s.season_id asc  ",
		leagueAbbr)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.MatchDetails
		err := raws.Scan(&g.LeagueID, &g.ClientID, &g.League, &g.LeagueAbbrv, &g.MatchDate, &g.SeasonID)
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

// GetProductionWinningOutcomes : returns winning outcomes used by API.
func (r *MysqlRepository) GetProductionWinningOutcomes(ctx context.Context, seasonWeekID string) ([]leagues.MatchWinningOutcomeAPI, error) {
	var gc []leagues.MatchWinningOutcomeAPI
	statement := fmt.Sprintf("select season_week_id,week_number,status,start_time,end_time from ssn_weeks \n"+
		"where season_week_id = '%s' and start_time < now() + interval 10 second   ", seasonWeekID)
	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.MatchWinningOutcomeAPI
		err := raws.Scan(&g.SeasonWeekID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime)
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

// GetProductionMatchDetailsNew : returns production match details used by API.
func (r *MysqlRepository) GetProductionMatchDetailsNew(ctx context.Context, leagueAbbr string) ([]leagues.MatchDetails, error) {
	var gc []leagues.MatchDetails
	statement := fmt.Sprintf("select l.league_id,l.client_id,l.league,l.league_abbrv,date(now()),s.season_id from leagues as l \n"+
		"inner join sns as s on s.league_id = l.league_id \n"+
		"where l.league_abbrv = '%s' order by s.season_id desc limit 1  ",
		//"where md.date = date(now()) and l.league_abbrv = '%s' order by s.season_id asc  ",
		leagueAbbr)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.MatchDetails
		err := raws.Scan(&g.LeagueID, &g.ClientID, &g.League, &g.LeagueAbbrv, &g.MatchDate, &g.SeasonID)
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

// GetProductionWinningOutcomesNew : returns winning outcomes used by API.
func (r *MysqlRepository) GetProductionWinningOutcomesNew(ctx context.Context, seasonWeekID string) ([]leagues.MatchWinningOutcomeAPI, error) {
	var gc []leagues.MatchWinningOutcomeAPI
	statement := fmt.Sprintf("select season_week_id,week_number,status,start_time,end_time from sn_wks \n"+
		"where season_week_id = '%s' and start_time < now() + interval 10 second   ", seasonWeekID)
	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g leagues.MatchWinningOutcomeAPI
		err := raws.Scan(&g.SeasonWeekID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime)
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
