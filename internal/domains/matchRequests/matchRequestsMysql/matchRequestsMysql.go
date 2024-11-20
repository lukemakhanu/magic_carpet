package matchRequestsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests"
)

var _ matchRequests.MatchRequestsRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t matchRequests.MatchRequests) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT match_requests SET player_id=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.PlayerID)

	if err != nil {
		return d, fmt.Errorf("unable to save match_requests : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve match requests ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// | match_requests | CREATE TABLE `match_requests` (
//   `match_request_id` bigint NOT NULL AUTO_INCREMENT,
//   `player_id` bigint NOT NULL,
//   `created` datetime DEFAULT NULL,
//   `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

func (r *MysqlRepository) MatchRequestDesc(ctx context.Context, playerID string) ([]matchRequests.MatchRequests, error) {
	var gc []matchRequests.MatchRequests
	statement := fmt.Sprintf("select match_request_id,player_id,start_time,end_time,\n"+
		"created,modified from match_requests where player_id = '%s' \n"+
		" order by match_request_id asc",
		playerID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matchRequests.MatchRequests
		err := raws.Scan(&g.MatchRequestID, &g.PlayerID, &g.Created, &g.Modified)
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
