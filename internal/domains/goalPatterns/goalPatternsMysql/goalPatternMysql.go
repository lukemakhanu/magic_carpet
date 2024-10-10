package goalPatternsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns"
)

var _ goalPatterns.GoalPatternsRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t goalPatterns.GoalPatterns) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT goal_patterns SET season_id=?,round_number_id=?,competition_id=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.SeasonID, t.RoundNumberID, t.CompetitionID)

	if err != nil {
		return d, fmt.Errorf("unable to save goal_patterns : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last goal_pattern ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}
