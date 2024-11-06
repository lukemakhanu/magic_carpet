package goalsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goals"
)

var _ goals.GoalsRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t goals.Goals) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT goals SET match_id=?,country=?,project_id=?,category=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.MatchID, t.Country, t.ProjectID, t.Category)

	if err != nil {
		return d, fmt.Errorf("unable to save leagues : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last goal ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// GetGoalCategory : returns data from a certain category
func (r *MysqlRepository) GetGoalCategory(ctx context.Context, country, category, projectID string) ([]goals.Goals, error) {
	var gc []goals.Goals
	statement := fmt.Sprintf("select goal_id,match_id,country,project_id,category,created,modified \n"+
		"from goals where country='%s' and category='%s' and project_id='%s' ",
		country, category, projectID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g goals.Goals
		err := raws.Scan(&g.GoalID, &g.MatchID, &g.Country, &g.ProjectID, &g.Category, &g.Created, &g.Modified)
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
