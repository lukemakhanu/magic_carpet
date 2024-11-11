package cleanUpsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/cleanUps"
)

var _ cleanUps.CleanUpsRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t cleanUps.CleanUps) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT clean_ups SET project_id=?,clean_up_date=date(now()),status=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.ProjectID, t.Status)

	if err != nil {
		return d, fmt.Errorf("unable to save cleanups : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last cleanup ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) CleanUpsByStatus(ctx context.Context, status string) ([]cleanUps.CleanUps, error) {
	var gc []cleanUps.CleanUps

	statement := fmt.Sprintf("select clean_up_id,project_id,clean_up_date,status,created,modified from clean_ups \n"+
		"where status ='%s' order by 1 desc limit 1 ", status)
	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g cleanUps.CleanUps
		err := raws.Scan(&g.CleanUpID, &g.ProjectID, &g.CleanUpDate, &g.Status, &g.Created, &g.Modified)
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

func (r *MysqlRepository) LastCleanUps(ctx context.Context) ([]cleanUps.CleanUps, error) {
	var gc []cleanUps.CleanUps

	statement := fmt.Sprintf("select clean_up_id,project_id,clean_up_date,status,created,modified from clean_ups \n" +
		"order by 1 desc limit 1 ")
	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g cleanUps.CleanUps
		err := raws.Scan(&g.CleanUpID, &g.ProjectID, &g.CleanUpDate, &g.Status, &g.Created, &g.Modified)
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

// | clean_ups | CREATE TABLE `clean_ups` (
//   `clean_up_id` int NOT NULL AUTO_INCREMENT,
//   `project_id` smallint NOT NULL,
//   `clean_up_date` varchar(50) NOT NULL,
//   `status` enum('pending','cleaned','processed') NOT NULL,
//   `created` datetime NOT NULL,
//   `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

// UpdateCleanUps : updates cleanups
func (mr *MysqlRepository) UpdateCleanUps(ctx context.Context, cleanUpID, status string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update clean_ups set status=?,modified=now() where clean_up_id = ? ", status, cleanUpID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// Save :
func (mr *MysqlRepository) SaveForTomorrow(ctx context.Context, t cleanUps.CleanUps) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT clean_ups SET project_id=?,clean_up_date=date(now() + interval 1 day),status=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.ProjectID, t.Status)

	if err != nil {
		return d, fmt.Errorf("unable to save cleanups : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last cleanup ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}
