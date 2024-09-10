package saveFile

import (
	"context"
	"fmt"
	"log"

	"github.com/lukemakhanu/magic_carpet/internal/domains/liveScoreFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/liveScoreFiles/liveScoreFilesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles/oddsFilesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/readFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/readFiles/readDir"
	"github.com/lukemakhanu/magic_carpet/internal/domains/woFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/woFiles/woFilesMysql"
)

// SaveFileConfiguration is an alias for a function that will take in a pointer to an SaveFileService and modify it
type SaveFileConfiguration func(os *SaveFileService) error

// SaveFileService is a implementation of the SaveFileService
type SaveFileService struct {
	oddsFileMysql oddsFiles.OddsFilesRepository
	woFilesMysql  woFiles.WoFilesRepository
	dirReader     readFiles.DirectoryReader

	// New Implementation
	liveScoreFilesMysql liveScoreFiles.LiveScoreFilesRepository
}

// NewSaveFileService : instantiate every connection we need to run current game service
func NewSaveFileService(cfgs ...SaveFileConfiguration) (*SaveFileService, error) {
	// Create the seasonService
	os := &SaveFileService{}
	// Apply all Configurations passed in
	for _, cfg := range cfgs {
		// Pass the service into the configuration function
		err := cfg(os)
		if err != nil {
			return nil, err
		}
	}
	return os, nil
}

// WithMysqlLiveScoreFilesRepository :
func WithMysqlLiveScoreFilesRepository(connectionString string) SaveFileConfiguration {
	return func(os *SaveFileService) error {
		d, err := liveScoreFilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.liveScoreFilesMysql = d
		return nil
	}
}

// WithMysqlOddsFilesRepository : instantiates mysql to connect to teams interface
func WithMysqlOddsFilesRepository(connectionString string) SaveFileConfiguration {
	return func(os *SaveFileService) error {
		// Create Matches repo
		d, err := oddsFilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.oddsFileMysql = d
		return nil
	}
}

// WithMysqlWoFilesRepository :
func WithMysqlWoFilesRepository(connectionString string) SaveFileConfiguration {
	return func(os *SaveFileService) error {
		// Create Matches repo
		d, err := woFilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.woFilesMysql = d
		return nil
	}
}

// WithDirectoryReaderRepository : read directory
func WithDirectoryReaderRepository(directory, combination string) SaveFileConfiguration {
	return func(os *SaveFileService) error {
		// Create Matches repo
		d, err := readDir.New(directory, combination)
		if err != nil {
			return err
		}
		os.dirReader = d
		return nil
	}
}

// SaveFiles : used to save data files into db
func (s *SaveFileService) SaveFiles(ctx context.Context, fileDir, country string) error {

	allFiles, err := s.dirReader.AllFiles(ctx)
	if err != nil {
		return fmt.Errorf("Err : %v failed to read files", err)
	}

	for _, f := range allFiles {

		// Save this files into our databases

		odds, err := oddsFiles.NewOddsFile(f.LsFileName, fileDir, country, f.ExtID, f.CompetitionID, f.ProjectID)
		if err != nil {
			return fmt.Errorf("Err : %v failed to instantiate files", err)
		}

		lastID, err := s.oddsFileMysql.SaveOdd(ctx, *odds)
		if err != nil {
			return fmt.Errorf("Err : %v failed to save files", err)
		}

		log.Printf("Last inserted ID  %d", lastID)

	}

	return nil
}

// SaveLiveScoreFiles : used to save live score files into db
func (s *SaveFileService) SaveLiveScoreFiles(ctx context.Context, fileDir, country string) error {

	allFiles, err := s.dirReader.ReadLiveScoreDirectory(ctx)
	if err != nil {
		return fmt.Errorf("Err : %v failed to read files", err)
	}

	for _, f := range allFiles {

		// Save this files into our databases

		ls, err := liveScoreFiles.NewLiveScoreFile(f.LsFileName, fileDir, country, f.ExtID, f.ProjectID, f.CompetitionID)
		if err != nil {
			return fmt.Errorf("Err : %v failed to instantiate live score files", err)
		}

		lastID, err := s.liveScoreFilesMysql.Save(ctx, *ls)
		if err != nil {
			return fmt.Errorf("Err : %v failed to save live score files", err)
		}

		log.Printf("Last inserted ID : %d", lastID)

	}

	return nil
}

// SaveWinningOutcomesFiles : used to save live score files into db
func (s *SaveFileService) SaveWinningOutcomesFiles(ctx context.Context, fileDir, country string) error {

	allFiles, err := s.dirReader.ReadWinningOutcomeDirectory(ctx)
	if err != nil {
		return fmt.Errorf("Err : %v failed to read files", err)
	}

	for _, f := range allFiles {

		// Save this files into our databases

		wo, err := woFiles.NewWoFile(f.LsFileName, fileDir, country, f.ExtID, f.ProjectID, f.CompetitionID)
		if err != nil {
			return fmt.Errorf("Err : %v failed to instantiate wo files", err)
		}

		lastID, err := s.woFilesMysql.SaveWo(ctx, *wo)
		if err != nil {
			return fmt.Errorf("Err : %v failed to save wo files", err)
		}

		log.Printf("Last inserted ID : %d", lastID)

	}

	return nil
}
