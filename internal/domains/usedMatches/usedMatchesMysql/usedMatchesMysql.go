package usedMatchesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goals"
	"github.com/lukemakhanu/magic_carpet/internal/domains/usedMatches"
)

var _ usedMatches.UsedMatchesRepository = (*MysqlRepository)(nil)

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

//	`used_match_id` bigint(50) NOT NULL,
//	`country` varchar(10) NOT NULL,
//	`project_id` smallint(4) NOT NULL,
//	`match_id` varchar(100) NOT NULL,
//	`category` varchar(150) NOT NULL,
//	`created` datetime NOT NULL,
//	`modified` datetime NOT NULL DEFAULT current_timestamp()
//

// Save :
func (mr *MysqlRepository) Save(ctx context.Context, t usedMatches.UsedMatches) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT used_matches SET country=?,project_id=?,match_id=?,category=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.Country, t.ProjectID, t.MatchID, t.Category)

	if err != nil {
		return d, fmt.Errorf("unable to save sns : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last ssn ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetAvailable(ctx context.Context, category string) ([]goals.Goals, error) {
	var gc []goals.Goals
	statement := fmt.Sprintf("select goal_id,country,project_id,match_id,category,\n"+
		"created,modified from goals where category = '%s' and \n"+
		"match_id not in(select match_id from used_matches where category = '%s') ",
		category, category)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g goals.Goals
		err := raws.Scan(&g.GoalID, &g.Country, &g.ProjectID, &g.MatchID, &g.Category, &g.Created, &g.Modified)
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