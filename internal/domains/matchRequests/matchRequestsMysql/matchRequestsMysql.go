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

// ALTER TABLE `match_requests`
//   ADD KEY `match_request_id` (`match_request_id`),
//   ADD KEY `player_id` (`player_id`),
//   ADD KEY `start_time` (`start_time`),
//   ADD KEY `end_time` (`end_time`),
//   ADD KEY `early_finish` (`early_finish`),
//   ADD KEY `created` (`created`),
//   ADD KEY `played` (`played`),
//   ADD KEY `modified` (`modified`);

// Save :
func (mr *MysqlRepository) Save(ctx context.Context, t matchRequests.MatchRequests) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT match_requests SET instant_competition_id=?,player_id=?,start_time=?,end_time=?, \n"+
		"early_finish=?,played=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.InstantCompetitionID, t.PlayerID, t.StartTime, t.EndTime,
		t.EarlyFinish, t.Played)

	if err != nil {
		return d, fmt.Errorf("unable to save match_requests : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve match requests ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// UpdateEarlyFinish :
func (mr *MysqlRepository) UpdateEarlyFinish(ctx context.Context, earlyFinish, matchRequestID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update match_requests set early_finish=?,modified=now() where match_request_id = ? ",
		earlyFinish, matchRequestID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// UpdateEarlyFinish :
func (mr *MysqlRepository) UpdatePlayed(ctx context.Context, earlyFinish, matchRequestID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update match_requests set played=?,modified=now() where match_request_id = ? ",
		earlyFinish, matchRequestID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
