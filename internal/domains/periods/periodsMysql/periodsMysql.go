package periodsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/periods"
)

var _ periods.PeriodsRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t periods.Periods) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT periods SET competition_id=?,match_request_id=?,start_time=?,end_time=?, \n"+
		"early_finish=?,played=?,game_started=?,key_created=?,round_number_id=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.CompetitionID, t.MatchRequestID, t.MatchRequestID, t.StartTime, t.EndTime,
		t.EarlyFinish, t.Played, t.GameStarted, t.KeyCreated, t.RoundNumberID)

	if err != nil {
		return d, fmt.Errorf("Unable to save period : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last period ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) PeriodByMatchRequestID(ctx context.Context, matchRequestID string) ([]periods.Periods, error) {
	var gc []periods.Periods
	statement := fmt.Sprintf("select period_id,competition_id,match_request_id,start_time,end_time,\n"+
		"early_finish,played,game_started,key_created,round_number_id,created,modified \n"+
		"from periods where match_request_id = '%s' ",
		matchRequestID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g periods.Periods
		err := raws.Scan(&g.PeriodID, &g.CompetitionID, &g.MatchRequestID, &g.StartTime, &g.EndTime,
			&g.EarlyFinish, &g.Played, &g.GameStarted, &g.KeyCreated, &g.RoundNumberID, &g.Created, &g.Modified)
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

// | periods | CREATE TABLE `periods` (
//   `period_id` bigint NOT NULL AUTO_INCREMENT,
//   `competition_id` smallint NOT NULL,
//   `match_request_id` bigint NOT NULL,
//   `start_time` datetime DEFAULT NULL,
//   `end_time` datetime DEFAULT NULL,
//   `early_finish` enum('yes','no') NOT NULL,
//   `played` enum('yes','no') NOT NULL,
//   `game_started` enum('yes','no') NOT NULL,
//   `key_created` enum('pending','created','spoilt') NOT NULL,
//   `round_number_id` varchar(60) NOT NULL,
//   `created` datetime NOT NULL,
//   `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

func (mr *MysqlRepository) UpdateEarlyFinish(ctx context.Context, earlyFinish, periodID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update periods set early_finish=?,modified=now() where period_id = ? ",
		earlyFinish, periodID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

func (mr *MysqlRepository) UpdateStartTime(ctx context.Context, startTime, endTime, played, periodID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update periods set start_time=?,end_time=?,played=?,modified=now() where period_id = ? ",
		startTime, endTime, played, periodID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

func (mr *MysqlRepository) UpdateKeyCreated(ctx context.Context, keyCreated, periodID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update periods set key_created=?,modified=now() where period_id = ? ",
		keyCreated, periodID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
