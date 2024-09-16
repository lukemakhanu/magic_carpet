package seasonWeekMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/seasonWeeks"
)

var _ seasonWeeks.SeasonWeeksRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t seasonWeeks.SeasonWeeks) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT sn_wks SET season_id=?,week_number=?,status=?,start_time=?, \n"+
		"end_time=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.SeasonID, t.WeekNumber, t.Status, t.StartTime, t.EndTime)

	if err != nil {
		return d, fmt.Errorf("Unable to save season weeks : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last season week ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetSeasonWeeks(ctx context.Context) ([]seasonWeeks.SeasonWeeks, error) {
	var gc []seasonWeeks.SeasonWeeks
	statement := fmt.Sprintf("select season_week_id,season_id,week_number,status,start_time,end_time,created,modified from sn_wks ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeeks
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetSeasonWeekByID(ctx context.Context, seasonWeekID string) ([]seasonWeeks.SeasonWeeks, error) {
	var gc []seasonWeeks.SeasonWeeks
	statement := fmt.Sprintf("select season_week_id,season_id,week_number,status,start_time,end_time,created,modified from sn_wks where season_week_id = '%s' ",
		seasonWeekID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeeks
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.Created, &g.Modified)
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

// GroupSeasonWeeks :
func (r *MysqlRepository) GroupSeasonWeeks(ctx context.Context, seasonWeekID string) ([]seasonWeeks.SeasonWeeks, error) {
	var gc []seasonWeeks.SeasonWeeks
	statement := fmt.Sprintf("select season_week_id,season_id,week_number,status,start_time,end_time,created,modified from sn_wks where season_week_id = '%s' ",
		seasonWeekID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeeks
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.Created, &g.Modified)
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

// UpdateSeasonWeek :
func (mr *MysqlRepository) UpdateSeasonWeek(ctx context.Context, seasonWeekID, seasonID, weekNumber, status, startTime, endTime string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update sn_wks set season_id=?,week_number=?,status=?,start_time=?,end_time=? where season_week_id = ? ",
		seasonID, weekNumber, status, startTime, endTime, seasonWeekID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

func (r *MysqlRepository) UpcomingSeasonWeeks(ctx context.Context) ([]seasonWeeks.SeasonWeeks, error) {
	var gc []seasonWeeks.SeasonWeeks
	statement := fmt.Sprintf("select season_week_id,season_id,week_number,status,start_time,end_time,\n" +
		"created,modified from sn_wks where start_time > now() + interval 20 minute limit 4")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeeks
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.Created, &g.Modified)
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

// UpcomingSeasonWeeksTest : used to return games for just one league for test environment
func (r *MysqlRepository) UpcomingSeasonWeeksTest(ctx context.Context) ([]seasonWeeks.SeasonWeekDetails, error) {
	var gc []seasonWeeks.SeasonWeekDetails
	statement := fmt.Sprintf("select sw.season_week_id,sw.season_id,s.league_id,sw.week_number,sw.status,sw.start_time,\n" +
		"sw.end_time, sw.created,sw.modified from sn_wks as sw \n" +
		"inner join seasons as s on sw.season_id=s.season_id \n" +
		"where s.league_id = 1 and sw.status= 'inactive' and sw.start_time > now() + interval 2 minute order by sw.start_time asc;")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeekDetails
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.LeagueID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.Created, &g.Modified)
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

// UpdateSeasonWeek :
func (mr *MysqlRepository) UpdateSeasonWeekStatus(ctx context.Context, seasonWeekID, status string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update sn_wks set status=? where season_week_id = ? ",
		status, seasonWeekID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

func (r *MysqlRepository) ApiSeasonWeeks(ctx context.Context, seasonID string) ([]seasonWeeks.SeasonWeeksAPI, error) {
	var gc []seasonWeeks.SeasonWeeksAPI
	statement := fmt.Sprintf("select season_week_id,season_id,week_number,status,start_time,end_time, date(start_time) as api_date, \n"+
		"created,modified from sn_wks where season_id = '%s' and status='active' and start_time > now() + interval 4 minute order by start_time asc limit 100 ", seasonID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeeksAPI
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.ApiDate, &g.Created, &g.Modified)
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

// UpcomingSsnWeeks : used to return games for just one league for test environment
func (r *MysqlRepository) UpcomingSsnWeeks(ctx context.Context) ([]seasonWeeks.SeasonWeekDetails, error) {
	var gc []seasonWeeks.SeasonWeekDetails

	statement := fmt.Sprintf("select sw.season_week_id,sw.season_id,s.league_id,sw.week_number,sw.status,sw.start_time,\n" +
		"sw.end_time, sw.created,sw.modified from sn_wks as sw \n" +
		"inner join sns as s on sw.season_id=s.season_id \n" +
		"where sw.status= 'inactive' and sw.start_time > now() + interval 2 minute order by sw.start_time asc limit 50")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g seasonWeeks.SeasonWeekDetails
		err := raws.Scan(&g.SeasonWeekID, &g.SeasonID, &g.LeagueID, &g.WeekNumber, &g.Status, &g.StartTime, &g.EndTime, &g.Created, &g.Modified)
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

// UpdateSsnWeekStatus :
func (mr *MysqlRepository) UpdateSsnWeekStatus(ctx context.Context, seasonWeekID, seasonID, status string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update sn_wks set status=? where season_week_id = ? and season_id = ? ",
		status, seasonWeekID, seasonID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
