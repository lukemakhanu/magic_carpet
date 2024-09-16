package checkMatchesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches"
)

var _ checkMatches.CheckMatchesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t checkMatches.CheckMatches) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT check_matches SET country=?,match_id=?,match_date=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.Country, t.MatchID, t.MatchDate)

	if err != nil {
		return d, fmt.Errorf("Unable to save check matches : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last check match ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (mr *MysqlRepository) MatchExist(ctx context.Context, country, matchID string) (bool, int, string, error) {

	var matchDate interface{}
	var count int

	statement := fmt.Sprintf("select count(*), match_date from check_matches where country='%s' and match_id='%s' \n"+
		"and created > now() - interval 5 day ", country, matchID)

	row := mr.db.QueryRow(statement)
	switch err := row.Scan(&count, &matchDate); err {
	case sql.ErrNoRows:
		return false, count, fmt.Sprintf("%v", matchDate), fmt.Errorf("Match does not exist : %v", err)
	case nil:
		return true, count, fmt.Sprintf("%v", matchDate), nil
	default:
		return true, count, fmt.Sprintf("%v", matchDate), fmt.Errorf("Unable to return check match count :: %v", err)
	}
}

// Delete : delete from check_matches table
func (d *MysqlRepository) Delete(ctx context.Context) (int64, error) {
	var rs int64
	result, err := d.db.Exec("delete from check_matches where created < now() - interval 10 day limit 50000")
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
