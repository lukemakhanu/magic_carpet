package snwkptsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/snwkpts"
)

var _ snwkpts.SnWkPtsRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t snwkpts.SnWkPts) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT sn_wk_pts SET season_week_id=?,round_number_id=?,competition_id=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.SeasonWeekID, t.RoundNumberID, t.CompetitionID)

	if err != nil {
		return d, fmt.Errorf("unable to save sn_wk_pt : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last sn_wk_pts ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// GoalDistributions :
func (r *MysqlRepository) SnWkPts(ctx context.Context, competitionID string) ([]snwkpts.SnWkPts, error) {
	var gc []snwkpts.SnWkPts

	statement := fmt.Sprintf("select sn_wk_pt_id,season_week_id,round_number_id,competition_id from sn_wk_pts \n"+
		"where competition_id = '%s'   ", competitionID)
	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g snwkpts.SnWkPts
		err := raws.Scan(&g.SnWkPtID, &g.SeasonWeekID, &g.RoundNumberID, &g.CompetitionID)
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
