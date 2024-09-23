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
	rs, err := mr.db.Exec("INSERT selected_matches SET player_id=?,match_request_id=?,parent_match_id=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.PlayerID, t.MatchRequestID, t.ParentMatchID)

	if err != nil {
		return d, fmt.Errorf("Unable to save selected_matches : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last selected matches ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

/* CREATE TABLE `selected_matches` (
`selected_matches_id` bigint(40) NOT NULL,
`player_id` bigint(30) NOT NULL,
`match_request_id` bigint(40) NOT NULL,
`parent_match_id` varchar(50) NOT NULL,
`created` datetime NOT NULL,
`modified` datetime NOT NULL */

func (r *MysqlRepository) GetMatchesbyRequestID(ctx context.Context, matchRequestID string) ([]selectedMatches.SelectedMatches, error) {
	var gc []selectedMatches.SelectedMatches
	statement := fmt.Sprintf("select selected_matches_id,player_id,match_request_id,parent_match_id,created,\n"+
		"modified from selected_matches where selected_matches_id = '%s' ",
		matchRequestID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g selectedMatches.SelectedMatches
		err := raws.Scan(&g.SelectedMatchesID, &g.PlayerID, &g.MatchRequestID, &g.ParentMatchID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetMatchesbyPlayerID(ctx context.Context, playerID string) ([]selectedMatches.SelectedMatches, error) {
	var gc []selectedMatches.SelectedMatches
	statement := fmt.Sprintf("select selected_matches_id,player_id,match_request_id,parent_match_id,created,\n"+
		"modified from selected_matches where player_id = '%s' ",
		playerID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g selectedMatches.SelectedMatches
		err := raws.Scan(&g.SelectedMatchesID, &g.PlayerID, &g.MatchRequestID, &g.ParentMatchID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetMatchesbyMatchRequestID(ctx context.Context, matchRequestID string) ([]selectedMatches.SelectedMatches, error) {
	var gc []selectedMatches.SelectedMatches
	statement := fmt.Sprintf("select selected_matches_id,player_id,match_request_id,parent_match_id,created,\n"+
		"modified from selected_matches where match_request_id = '%s' ",
		matchRequestID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g selectedMatches.SelectedMatches
		err := raws.Scan(&g.SelectedMatchesID, &g.PlayerID, &g.MatchRequestID, &g.ParentMatchID, &g.Created, &g.Modified)
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
