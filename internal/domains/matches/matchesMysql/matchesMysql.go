package matchesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/matches"
)

var _ matches.MatchesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t matches.Matches) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT matches SET season_week_id=?,home_team_id=?,away_team_id=?,status=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.SeasonWeekID, t.HomeTeamID, t.AwayTeamID, t.Status)

	if err != nil {
		return d, fmt.Errorf("Unable to save leagues : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last matche ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetMatches(ctx context.Context) ([]matches.Matches, error) {
	var gc []matches.Matches
	statement := fmt.Sprintf("select match_id,season_week_id,home_team_id,away_team_id,status,created,modified from matches ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matches.Matches
		err := raws.Scan(&g.MatchID, &g.SeasonWeekID, &g.HomeTeamID, &g.AwayTeamID, &g.Status, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetMatchByID(ctx context.Context, matchID string) ([]matches.Matches, error) {
	var gc []matches.Matches
	statement := fmt.Sprintf("select match_id,season_week_id,home_team_id,away_team_id,status,created,modified from matches where match_id = '%s' ",
		matchID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matches.Matches
		err := raws.Scan(&g.MatchID, &g.SeasonWeekID, &g.HomeTeamID, &g.AwayTeamID, &g.Status, &g.Created, &g.Modified)
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

// UpdateMatch : updates match table
func (mr *MysqlRepository) UpdateMatch(ctx context.Context, seasonWeekID, HomeTeamID, awayTeamID, status, matchID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update matches set season_week_id=?,home_team_id=?,away_team_id=?,status=? where match_id = ? ",
		seasonWeekID, HomeTeamID, awayTeamID, status, matchID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// GetMatchBySeasonWeek : returns matches based
func (r *MysqlRepository) GetMatchBySeasonWeek(ctx context.Context, seasonWeekID string) ([]matches.Matches, error) {
	var gc []matches.Matches
	//statement := fmt.Sprintf("select match_id,season_week_id,home_team_id,away_team_id,status,created,modified from matches where season_week_id = '%s'  ", seasonWeekID)
	statement := fmt.Sprintf("select match_id,season_week_id,home_team_id,away_team_id,status,created,modified from matches where season_week_id = '%s' order by match_id asc limit 10 ", seasonWeekID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matches.Matches
		err := raws.Scan(&g.MatchID, &g.SeasonWeekID, &g.HomeTeamID, &g.AwayTeamID, &g.Status, &g.Created, &g.Modified)
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

// UpdateMatchStatus : updates match status
func (mr *MysqlRepository) UpdateMatchStatus(ctx context.Context, status, matchID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update matches set status=? where match_id = ? ", status, matchID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// GetSeasonWeekGames : returns matches based
func (r *MysqlRepository) GetSeasonWeekGames(ctx context.Context, seasonWeekID, seasonID string) ([]matches.MatchGames, error) {
	var gc []matches.MatchGames

	statement := fmt.Sprintf("select match_id,league_id,season_id,season_week_id,week_number,home_team_id,away_team_id,\n"+
		"status,created,modified from gms where season_week_id = '%s' and season_id='%s' order by match_id asc limit 10 ",
		seasonWeekID, seasonID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matches.MatchGames
		err := raws.Scan(&g.MatchID, &g.LeagueID, &g.SeasonID, &g.SeasonWeekID, &g.WeekNumber, &g.HomeTeamID, &g.AwayTeamID,
			&g.Status, &g.Created, &g.Modified)
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

// UpdateGameStatus : updates match status
func (mr *MysqlRepository) UpdateGameStatus(ctx context.Context, status, matchID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update gms set status=? where match_id = ? ", status, matchID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
