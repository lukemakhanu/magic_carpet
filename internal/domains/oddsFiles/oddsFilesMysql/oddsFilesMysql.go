package oddsFilesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
)

var _ oddsFiles.OddsFilesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t oddsFiles.OddsFiles) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT odds_file SET odds_file_name=?,file_directory=?,country=?, \n"+
		"parent_id=?,competition_id=?,match_id=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.OddsFileName, t.FileDirectory, t.Country, t.ParentID, t.CompetitionID, t.MatchID)

	if err != nil {
		return d, fmt.Errorf("unable to save leagues : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last odds file ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// SaveOdd :
func (mr *MysqlRepository) SaveOdd(ctx context.Context, t oddsFiles.OddsFiles) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT o_files SET odds_file_name=?,file_directory=?,country=?, \n"+
		"parent_id=?,competition_id=?,match_id=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.OddsFileName, t.FileDirectory, t.Country, t.ParentID, t.CompetitionID, t.MatchID)

	if err != nil {
		return d, fmt.Errorf("unable to save leagues : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last odds file ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetOddsFiles(ctx context.Context) ([]oddsFiles.OddsFiles, error) {
	var gc []oddsFiles.OddsFiles
	statement := fmt.Sprintf("select odds_file_id,odds_file_name,file_directory,country,parent_id,competition_id,match_id,created,modified from odds_file ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g oddsFiles.OddsFiles
		err := raws.Scan(&g.OddsFileID, &g.OddsFileName, &g.FileDirectory, &g.Country, &g.ParentID, &g.CompetitionID, &g.MatchID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetOddsFileByID(ctx context.Context, oddsFileID string) ([]oddsFiles.OddsFiles, error) {
	var gc []oddsFiles.OddsFiles
	statement := fmt.Sprintf("select odds_file_id,odds_file_name,file_directory,country,parent_id,competition_id,match_id,created,modified from odds_file where odds_file_id = '%s' ",
		oddsFileID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g oddsFiles.OddsFiles
		err := raws.Scan(&g.OddsFileID, &g.OddsFileName, &g.FileDirectory, &g.Country, &g.ParentID, &g.CompetitionID, &g.MatchID, &g.Created, &g.Modified)
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

// GetOddsFileWithLimit : used to get odds with limit.
func (r *MysqlRepository) GetOddsFileWithLimit(ctx context.Context, min, max string) ([]oddsFiles.OddsFiles, error) {
	var gc []oddsFiles.OddsFiles
	statement := fmt.Sprintf("select odds_file_id,odds_file_name,file_directory,country,parent_id,competition_id,match_id,created,modified from odds_file where odds_file_id >= '%s'  and odds_file_id <= '%s' ",
		min, max)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g oddsFiles.OddsFiles
		err := raws.Scan(&g.OddsFileID, &g.OddsFileName, &g.FileDirectory, &g.Country, &g.ParentID, &g.CompetitionID, &g.MatchID, &g.Created, &g.Modified)
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

// UpdateMatch : updates match table
func (mr *MysqlRepository) UpdateMatch(ctx context.Context, oddsFileName, fileDirectory, country, parentID, competitionID, matchID, oddsFileID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update odds_file set odds_file_name=?,file_directory=?,country=?,parent_id=?,competition_id=?,match_id=? where odds_file_id = ? ",
		oddsFileName, fileDirectory, country, parentID, competitionID, matchID, oddsFileID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

func (r *MysqlRepository) GetOddsParentID(ctx context.Context, parentID string) ([]oddsFiles.OddsFiles, error) {
	var gc []oddsFiles.OddsFiles
	statement := fmt.Sprintf("select odds_file_id,odds_file_name,file_directory,country,parent_id,competition_id,match_id,created,modified from odds_file where parent_id = '%s' ",
		parentID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g oddsFiles.OddsFiles
		err := raws.Scan(&g.OddsFileID, &g.OddsFileName, &g.FileDirectory, &g.Country, &g.ParentID, &g.CompetitionID, &g.MatchID, &g.Created, &g.Modified)
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

func (mr *MysqlRepository) SaveMatchOdds(ctx context.Context, roundID, parentID, country string) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT match_odds SET round_id=?,parent_id=?,country=?, \n"+
		"created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		roundID, parentID, country)

	if err != nil {
		return d, fmt.Errorf("unable to save match_odds : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last match odds ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetReadyMatchCount(ctx context.Context) ([]oddsFiles.MatchDet, error) {
	var gc []oddsFiles.MatchDet
	statement := fmt.Sprintf("select count(*) as matchCount,round_id from match_odds group by round_id")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g oddsFiles.MatchDet
		err := raws.Scan(&g.MatchCount, &g.RoundID)
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

// GetAllOddsParentID :
func (r *MysqlRepository) GetAllOddsParentID(ctx context.Context, parentID, countryCode string) ([]oddsFiles.OddsFiles, error) {
	var gc []oddsFiles.OddsFiles
	statement := fmt.Sprintf("select odds_file_id,odds_file_name,file_directory,country,parent_id,competition_id,match_id,\n"+
		"created,modified from o_files where parent_id = '%s' and country = '%s' ",
		parentID, countryCode)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g oddsFiles.OddsFiles
		err := raws.Scan(&g.OddsFileID, &g.OddsFileName, &g.FileDirectory, &g.Country, &g.ParentID, &g.CompetitionID, &g.MatchID, &g.Created, &g.Modified)
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
