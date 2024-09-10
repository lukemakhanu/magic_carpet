package liveScoreFilesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/liveScoreFiles"
)

var _ liveScoreFiles.LiveScoreFilesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t liveScoreFiles.LiveScoreFiles) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT live_scores_files SET ls_file_name=?,ls_dir=?,country=?,ext_id=?, \n"+
		"project_id=?,competition_id=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.LsFileName, t.LsDir, t.Country, t.ExtID, t.ProjectID, t.CompetitionID)

	if err != nil {
		return d, fmt.Errorf("unable to save live score files : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last ls file ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetLsFiles(ctx context.Context) ([]liveScoreFiles.LiveScoreFiles, error) {
	var gc []liveScoreFiles.LiveScoreFiles
	statement := fmt.Sprintf("select live_score_file_id,ls_file_name,ls_dir,country,ext_id,project_id,competition_id,created,modified from live_scores_files ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g liveScoreFiles.LiveScoreFiles
		err := raws.Scan(&g.LiveScoreFileID, &g.LsFileName, &g.LsDir, &g.Country, &g.ExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetLsFileByID(ctx context.Context, liveScoreFileID string) ([]liveScoreFiles.LiveScoreFiles, error) {
	var gc []liveScoreFiles.LiveScoreFiles
	statement := fmt.Sprintf("select live_score_file_id,ls_file_name,ls_dir,country,ext_id,project_id,competition_id,created,modified from live_scores_files where live_score_file_id='%s'  ",
		liveScoreFileID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g liveScoreFiles.LiveScoreFiles
		err := raws.Scan(&g.LiveScoreFileID, &g.LsFileName, &g.LsDir, &g.Country, &g.ExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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
func (mr *MysqlRepository) UpdateLsFile(ctx context.Context, lsFileName, lsDir, country, extID, liveScoreFileID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update live_scores_files set ls_file_name=?,ls_dir=?,country=?,ext_id=?,project_id=?,competition_id=? where live_score_file_id = ? ",
		lsFileName, lsDir, country, extID, liveScoreFileID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
