package playersMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/players"
)

var _ players.PlayersRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t players.Players) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT players SET profile_tag=?,status=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.ProfileTag, t.Status)

	if err != nil {
		return d, fmt.Errorf("Unable to save player : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last player ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// UpdatePlayer :
func (mr *MysqlRepository) UpdatePlayer(ctx context.Context, profileTag, status string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update players set status=?,modified=now() where profile_tag = ? ",
		status, profileTag)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// CREATE TABLE `players` (
//   `prayer_id` bigint(30) NOT NULL,
//   `profile_tag` varchar(400) NOT NULL,
//   `status` enum('active','inactive','suspended','blocked') NOT NULL,
//   `created` datetime NOT NULL,
//   `modified` datetime NOT NULL DEFAULT current_timestamp()
// );

func (r *MysqlRepository) PlayerExists(ctx context.Context, profileTag string) ([]players.Players, error) {
	var gc []players.Players
	statement := fmt.Sprintf("select player_id,profile_tag,status,created,modified from players where profile_tag = '%s' ",
		profileTag)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g players.Players
		err := raws.Scan(&g.PlayerID, &g.ProfileTag, &g.Status, &g.Created, &g.Modified)
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
