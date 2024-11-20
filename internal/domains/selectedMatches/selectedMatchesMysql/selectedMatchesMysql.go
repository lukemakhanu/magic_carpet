package selectedMatchesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches"
)

var _ selectedMatches.SelectedMatchesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t selectedMatches.SelectedMatches) (int, error) {
	var d int

	rs, err := mr.db.Exec("INSERT selected_matches SET player_id=?,period_id=?,parent_match_id=?, \n"+
		"home_team_id=?,away_team_id=?,status=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.PlayerID, t.PeriodID, t.ParentMatchID,
		t.HomeTeamID, t.AwayTeamID, t.Status)

	if err != nil {
		return d, fmt.Errorf("unable to save selected_matches : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last selected matches ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// | selected_matches | CREATE TABLE `selected_matches` (
//   `selected_matches_id` bigint NOT NULL AUTO_INCREMENT,
//   `player_id` bigint NOT NULL,
//   `period_id` bigint NOT NULL,
//   `parent_match_id` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
//   `home_team_id` smallint NOT NULL,
//   `away_team_id` smallint NOT NULL,
//   `status` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
//   `created` datetime NOT NULL,
//   `modified` datetime NOT NULL,

func (r *MysqlRepository) GetMatchesbyPlayerID(ctx context.Context, playerID string) ([]selectedMatches.SelectedMatches, error) {
	var gc []selectedMatches.SelectedMatches
	statement := fmt.Sprintf("select selected_matches_id,player_id,period_id,parent_match_id,\n"+
		"home_team_id,away_team_id,status,created,\n"+
		"modified from selected_matches where player_id = '%s' ",
		playerID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g selectedMatches.SelectedMatches
		err := raws.Scan(&g.SelectedMatchesID, &g.PlayerID, &g.PeriodID, &g.ParentMatchID,
			&g.HomeTeamID, &g.AwayTeamID, &g.Status,
			&g.Created, &g.Modified)
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

func (r *MysqlRepository) GetMatchesbyPeriodID(ctx context.Context, periodID string) ([]selectedMatches.SelectedMatches, error) {
	var gc []selectedMatches.SelectedMatches
	statement := fmt.Sprintf("select selected_matches_id,player_id,period_id,parent_match_id,\n"+
		"home_team_id,away_team_id,status,created,\n"+
		"modified from selected_matches where period_id = '%s' ",
		periodID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g selectedMatches.SelectedMatches
		err := raws.Scan(&g.SelectedMatchesID, &g.PlayerID, &g.PeriodID, &g.ParentMatchID,
			&g.HomeTeamID, &g.AwayTeamID, &g.Status,
			&g.Created, &g.Modified)
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
