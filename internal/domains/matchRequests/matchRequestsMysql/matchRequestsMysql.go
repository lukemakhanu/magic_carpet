package matchRequestsMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests"
)

var _ matchRequests.MatchRequestsRepository = (*MysqlRepository)(nil)

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

// ALTER TABLE `match_requests`
//   ADD KEY `match_request_id` (`match_request_id`),
//   ADD KEY `player_id` (`player_id`),
//   ADD KEY `start_time` (`start_time`),
//   ADD KEY `end_time` (`end_time`),
//   ADD KEY `early_finish` (`early_finish`),
//   ADD KEY `created` (`created`),
//   ADD KEY `played` (`played`),
//   ADD KEY `modified` (`modified`);

// Save :
func (mr *MysqlRepository) Save(ctx context.Context, t matchRequests.MatchRequests) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT match_requests SET instant_competition_id=?,player_id=?,start_time=?,end_time=?, \n"+
		"early_finish=?,played=?,key_created=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.InstantCompetitionID, t.PlayerID, t.StartTime, t.EndTime,
		t.EarlyFinish, t.Played, t.KeyCreated)

	if err != nil {
		return d, fmt.Errorf("unable to save match_requests : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve match requests ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// UpdateEarlyFinish :
func (mr *MysqlRepository) UpdateEarlyFinish(ctx context.Context, earlyFinish, matchRequestID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update match_requests set early_finish=?,modified=now() where match_request_id = ? ",
		earlyFinish, matchRequestID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// UpdateEarlyFinish :
func (mr *MysqlRepository) UpdatePlayed(ctx context.Context, earlyFinish, matchRequestID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update match_requests set played=?,modified=now() where match_request_id = ? ",
		earlyFinish, matchRequestID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// | match_requests | CREATE TABLE `match_requests` (
//   `match_request_id` bigint(40) NOT NULL AUTO_INCREMENT,
//   `instant_competition_id` int(6) NOT NULL,
//   `player_id` bigint(30) NOT NULL,
//   `start_time` datetime NOT NULL,
//   `end_time` datetime NOT NULL,
//   `early_finish` enum('no','yes') NOT NULL DEFAULT 'no',
//   `played` enum('no','yes') NOT NULL DEFAULT 'no',
//   `key_created` enum('pending','created','spoilt') NOT NULL,
//   `created` datetime DEFAULT NULL,
//   `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

func (r *MysqlRepository) PendingRequestedMatchDesc(ctx context.Context, playerID, competitionID, keyCreated string) ([]matchRequests.MatchRequests, error) {
	var gc []matchRequests.MatchRequests
	statement := fmt.Sprintf("select match_request_id,instant_competition_id,player_id,start_time,end_time,\n"+
		"early_finish,played,key_created,created,modified from match_requests where player_id = '%s' and \n"+
		"instant_competition_id= '%s' and key_created='%s' order by match_request_id asc",
		playerID, competitionID, keyCreated)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matchRequests.MatchRequests
		err := raws.Scan(&g.MatchRequestID, &g.InstantCompetitionID, &g.PlayerID, &g.StartTime, &g.EndTime,
			&g.EarlyFinish, &g.Played, &g.KeyCreated, &g.Created, &g.Modified)
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

// PlayerUsedMatchesDesc : returns players recently used matches
func (r *MysqlRepository) PlayerUsedMatchesDesc(ctx context.Context, playerID string) ([]matchRequests.UsedParentMatchIDs, error) {
	var gc []matchRequests.UsedParentMatchIDs
	statement := fmt.Sprintf("select s.parent_match_id from match_requests as mr inner join selected_matches as s \n"+
		"where mr.player_id = '%s' order by match_request_id asc",
		playerID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g matchRequests.UsedParentMatchIDs
		err := raws.Scan(&g.ParentMatchID)
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
