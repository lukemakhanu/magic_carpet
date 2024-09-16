package scheduledTimeMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/scheduledTimes"
)

var _ scheduledTimes.ScheduledTimesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t scheduledTimes.ScheduledTime) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT scheduled_time SET scheduled_time=?,competition_id=?,status=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.ScheduledTime, t.CompetitionID, t.Status)

	if err != nil {
		return d, fmt.Errorf("Unable to save check scheduled time id : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last scheduled time ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// Delete : delete from check_matches table
func (d *MysqlRepository) Delete(ctx context.Context) (int64, error) {
	var rs int64
	result, err := d.db.Exec("delete from scheduled_time where start_time < now() - interval 10 day")
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// UpdateScheduleTime : updates scheduled time
func (mr *MysqlRepository) UpdateScheduleTime(ctx context.Context, status, scheduledTimeID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update scheduled_time set status=? where scheduled_time_id = ? ",
		status, scheduledTimeID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// CREATE TABLE `scheduled_time` (
// 	`scheduled_time_id` int(11) NOT NULL AUTO_INCREMENT,
// 	`scheduled_time` varchar(100) NOT NULL,
// 	`competition_id` smallint(4) NOT NULL,
// 	`status` enum('active','inactive') NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

func (r *MysqlRepository) GetScheduledTime(ctx context.Context, status, competitionID string) ([]scheduledTimes.ScheduledTime, error) {
	var gc []scheduledTimes.ScheduledTime
	statement := fmt.Sprintf("select scheduled_time_id,scheduled_time,competition_id,status, \n"+
		"created,modified from scheduled_time where status='%s' and competition_id='%s' ",
		status, competitionID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g scheduledTimes.ScheduledTime
		err := raws.Scan(&g.ScheduledTimeID, &g.ScheduledTime, &g.CompetitionID, &g.Status,
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

func (mr *MysqlRepository) CountScheduledTime(ctx context.Context, leagueID string) (int, error) {

	var count int

	statement := fmt.Sprintf("select count(scheduled_time_id) from scheduled_time where competition_id='%s' and scheduled_time > now() ", leagueID)

	row := mr.db.QueryRow(statement)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows:
		return count, nil
	case nil:
		return count, nil
	default:
		return count, fmt.Errorf("Unable to return scheduled_time count :: %v", err)
	}
}

func (ssn *MysqlRepository) FirstActiveScheduledTime(ctx context.Context, leagueID, status string) (string, string, error) {

	statement := fmt.Sprintf("select scheduled_time_id, scheduled_time from scheduled_time where competition_id='%s' and status='%s' order by scheduled_time_id asc limit 1 ",
		leagueID, status)

	var scheduledTimeID string
	var scheduledTime string

	row := ssn.db.QueryRow(statement)
	switch err := row.Scan(&scheduledTimeID, &scheduledTime); err {
	case sql.ErrNoRows:
		return scheduledTimeID, scheduledTime, fmt.Errorf("Err : %v No rows were returned!", err)
	case nil:
		return scheduledTimeID, scheduledTime, nil
	default:
		return scheduledTimeID, scheduledTime, err
	}
}
