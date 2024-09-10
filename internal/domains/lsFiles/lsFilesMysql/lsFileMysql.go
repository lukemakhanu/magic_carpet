package lsFilesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/lsFiles"
)

var _ lsFiles.LsFilesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t lsFiles.LsFiles) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT ls_files SET ls_file_name=?,ls_dir=?,country=?,ls_ext_id=?, \n"+
		"project_id=?,competition_id=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.LsFileName, t.LsDir, t.Country, t.LsExtID, t.ProjectID, t.CompetitionID)

	if err != nil {
		return d, fmt.Errorf("Unable to save ls files : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("Unable to retrieve last ls file ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetLsFiles(ctx context.Context) ([]lsFiles.LsFiles, error) {
	var gc []lsFiles.LsFiles
	statement := fmt.Sprintf("select ls_file_id,ls_file_name,ls_dir,country,ls_ext_id,project_id,competition_id,created,modified from ls_files ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g lsFiles.LsFiles
		err := raws.Scan(&g.LsFileID, &g.LsFileName, &g.LsDir, &g.Country, &g.LsExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetLsFileByID(ctx context.Context, lsFileID string) ([]lsFiles.LsFiles, error) {
	var gc []lsFiles.LsFiles
	statement := fmt.Sprintf("select ls_file_id,ls_file_name,ls_dir,country,ls_ext_id,project_id,competition_id,created,modified from ls_files where ls_file_id='%s'  ",
		lsFileID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g lsFiles.LsFiles
		err := raws.Scan(&g.LsFileID, &g.LsFileName, &g.LsDir, &g.Country, &g.LsExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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
func (mr *MysqlRepository) UpdateLsFile(ctx context.Context, lsFileName, lsDir, country, lsExtID, lsFileID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update ls_files set ls_file_name=?,ls_dir=?,country=?,ls_ext_id=?,project_id=?,competition_id=? where ls_file_id = ? ",
		lsFileName, lsDir, country, lsExtID, lsFileID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// GetLsFileByExtID : used to map live scores to winning outcomes
func (r *MysqlRepository) GetLsFileByExtID(ctx context.Context, lsExtID string) ([]lsFiles.LsFiles, error) {
	var gc []lsFiles.LsFiles
	statement := fmt.Sprintf("select ls_file_id,ls_file_name,ls_dir,country,ls_ext_id,project_id,\n"+
		"competition_id,created,modified from ls_files where ls_ext_id='%s'  ",
		lsExtID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g lsFiles.LsFiles
		err := raws.Scan(&g.LsFileID, &g.LsFileName, &g.LsDir, &g.Country, &g.LsExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

// GetLiveScoreFileByExtID : used to map live scores to winning outcomes
func (r *MysqlRepository) GetLiveScoreFileByExtID(ctx context.Context, lsExtID, countryCode string) ([]lsFiles.LsFiles, error) {

	var gc []lsFiles.LsFiles
	statement := fmt.Sprintf("select live_score_file_id,ls_file_name,ls_dir,country,ext_id,project_id,\n"+
		"competition_id,created,modified from live_scores_files where ext_id='%s' and country='%s'  ",
		lsExtID, countryCode)

	//log.Printf("Query ---> %s  ", statement)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g lsFiles.LsFiles
		err := raws.Scan(&g.LsFileID, &g.LsFileName, &g.LsDir, &g.Country, &g.LsExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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
