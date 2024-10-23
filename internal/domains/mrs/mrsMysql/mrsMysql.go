package mrsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs"
)

var _ mrs.MrsRepository = (*MysqlRepository)(nil)

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

func (mr *MysqlRepository) Save(ctx context.Context, t mrs.Mrs) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT mrs SET round_number_id=?,competition_id=?,start_time=?,total_goals=?, \n"+
		"goal_count=?,raw_scores=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.RoundNumberID, t.CompetitionID, t.StartTime, t.TotalGoals,
		t.GoalCount, t.RawScores)

	if err != nil {
		return d, fmt.Errorf("unable to save mrs : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last mr ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GoalPatterns(ctx context.Context, competitionID string) ([]mrs.Mrs, error) {
	var gc []mrs.Mrs
	statement := fmt.Sprintf("select mr_id,round_number_id,competition_id,start_time,total_goals,goal_count \n"+
		" from mrs where competition_id = '%s' ",
		competitionID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g mrs.Mrs
		err := raws.Scan(&g.MrID, &g.RoundNumberID, &g.CompetitionID, &g.StartTime, &g.TotalGoals, &g.GoalCount)
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

// GoalDistribution : pull the distribution to use
func (r *MysqlRepository) GoalDistribution(ctx context.Context, roundNumberID, competitionID string) ([]mrs.Mrs, error) {
	var gc []mrs.Mrs
	statement := fmt.Sprintf("select mr_id,round_number_id,competition_id,start_time,total_goals,goal_count,raw_scores \n"+
		" from mrs where competition_id = '%s' and round_number_id = '%s' ",
		competitionID, roundNumberID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g mrs.Mrs
		err := raws.Scan(&g.MrID, &g.RoundNumberID, &g.CompetitionID, &g.StartTime, &g.TotalGoals, &g.GoalCount, &g.RawScores)
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
