package woFilesMysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/woFiles"
)

var _ woFiles.WoFilesRepository = (*MysqlRepository)(nil)

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
func (mr *MysqlRepository) Save(ctx context.Context, t woFiles.WoFiles) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT wo_files SET wo_file_name=?,wo_dir=?,country=?,wo_ext_id=?, \n"+
		"project_id=?,competition_id=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.WoFileName, t.WoDir, t.Country, t.WoExtID, t.ProjectID, t.CompetitionID)

	if err != nil {
		return d, fmt.Errorf("unable to save ls files : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last ls file ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

// | winning_outcome_files | CREATE TABLE `winning_outcome_files` (
// 	`wo_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
// 	`wo_file_name` varchar(300) NOT NULL,
// 	`wo_dir` varchar(300) NOT NULL,
// 	`country` varchar(100) NOT NULL,
// 	`ext_id` bigint(20) NOT NULL,
// 	`project_id` int(6) NOT NULL,
// 	`competition_id` int(6) NOT NULL,
// 	`created` datetime NOT NULL,
// 	`modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

// Save :
func (mr *MysqlRepository) SaveWo(ctx context.Context, t woFiles.WoFiles) (int, error) {
	var d int
	rs, err := mr.db.Exec("INSERT winning_outcome_files SET wo_file_name=?,wo_dir=?,country=?,ext_id=?, \n"+
		"project_id=?,competition_id=?,created=now(),modified=now() ON DUPLICATE KEY UPDATE modified=now()",
		t.WoFileName, t.WoDir, t.Country, t.WoExtID, t.ProjectID, t.CompetitionID)

	if err != nil {
		return d, fmt.Errorf("unable to save wo files : %v", err)
	}

	lastInsertedID, err := rs.LastInsertId()
	if err != nil {
		return d, fmt.Errorf("unable to retrieve last ls file ID [primary key] : %v", err)
	}

	return int(lastInsertedID), nil
}

func (r *MysqlRepository) GetWoFiles(ctx context.Context) ([]woFiles.WoFiles, error) {
	var gc []woFiles.WoFiles
	statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,wo_ext_id,project_id,competition_id,created,modified from wo_files limit 5000 ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g woFiles.WoFiles
		err := raws.Scan(&g.WoFileID, &g.WoFileName, &g.WoDir, &g.Country, &g.WoExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetWoFileByID(ctx context.Context, woFileID string) ([]woFiles.WoFiles, error) {
	var gc []woFiles.WoFiles
	statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,wo_ext_id,project_id,competition_id,created,modified from wo_files where wo_file_id='%s'  ",
		woFileID)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g woFiles.WoFiles
		err := raws.Scan(&g.WoFileID, &g.WoFileName, &g.WoDir, &g.Country, &g.WoExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

// mysql> CREATE TABLE `wo_files` (
//
//	->   `wo_file_id` bigint(20) NOT NULL,
//	->   `wo_file_name` varchar(300) NOT NULL,
//	->   `wo_dir` varchar(300) NOT NULL,
//	->   `country` varchar(100) NOT NULL,
//	->   `wo_ext_id` bigint(20) NOT NULL,
//	->   `project_id` int(6) NOT NULL,
//	->   `competition_id` int(6) NOT NULL,
//	->   `created` datetime NOT NULL,
//	->   `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP

// UpdateMatch : updates match table
func (mr *MysqlRepository) UpdateWoFile(ctx context.Context, woFileName, woDir, country, woExtID, woFileID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update wo_files set wo_file_name=?,wo_dir=?,country=?,wo_ext_id=?,project_id=?,competition_id=? where ls_file_id = ? ",
		woFileName, woDir, country, woExtID, woFileID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}

// GetWinningOutcomeFiles :
func (r *MysqlRepository) GetWinningOutcomeFiles(ctx context.Context) ([]woFiles.WinningOutcomeFiles, error) {
	var gc []woFiles.WinningOutcomeFiles

	//statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 30000 limit 40000 ")
	//statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 0  limit 40000 ")
	//statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 70000 limit 30000 ")

	statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files order by wo_file_id limit 120000 ")

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g woFiles.WinningOutcomeFiles
		err := raws.Scan(&g.WoFileID, &g.WoFileName, &g.WoDir, &g.Country, &g.ExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

func (r *MysqlRepository) GetWO(ctx context.Context, statement string) ([]woFiles.WinningOutcomeFiles, error) {
	var gc []woFiles.WinningOutcomeFiles

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g woFiles.WinningOutcomeFiles
		err := raws.Scan(&g.WoFileID, &g.WoFileName, &g.WoDir, &g.Country, &g.ExtID, &g.ProjectID, &g.CompetitionID, &g.Created, &g.Modified)
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

// GetPendingWo : used to return pending winning outcomes that havent been processed
func (r *MysqlRepository) GetPendingWo(ctx context.Context, status string) ([]woFiles.WoFiles, error) {
	var gc []woFiles.WoFiles
	statement := fmt.Sprintf("select wo_file_id,wo_file_name,wo_dir,country,wo_ext_id,project_id,competition_id,status,created,modified from winning_outcome_files where status='%s'  limit 50",
		status)

	raws, err := r.db.Query(statement)
	if err != nil {
		return nil, err
	}

	for raws.Next() {
		var g woFiles.WoFiles
		err := raws.Scan(&g.WoFileID, &g.WoFileName, &g.WoDir, &g.Country, &g.WoExtID, &g.ProjectID, &g.CompetitionID, &g.Status, &g.Created, &g.Modified)
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

// UpdateWoStatus :
func (mr *MysqlRepository) UpdateWoStatus(ctx context.Context, status, woFileID string) (int64, error) {
	var rs int64
	result, err := mr.db.Exec("update winning_outcome_files set status=?,modified=now() where wo_file_id = ? ",
		status, woFileID)
	if err != nil {
		return rs, err
	}
	return result.RowsAffected()
}
