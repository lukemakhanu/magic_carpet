package saveFileRedis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/lsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/lsFiles/lsFilesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles/oddsFilesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/woFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/woFiles/woFilesMysql"
)

type Job struct {
	JobType  string
	WorkStr  string
	WorkStr1 string
	WorkStr2 string
	WorkStr3 string
}

// DataTunnel --
type DataTunnel struct {
	WoDir      string
	WoFileName string
	WoFileID   string
}

// FileProcessorConfiguration is an alias for a function that will take in a pointer to an FileProcessorService and modify it
type FileProcessorConfiguration func(os *FileProcessorService) error

// FileProcessorService is a implementation of the FileProcessorService
type FileProcessorService struct {
	lsFilesMysql  lsFiles.LsFilesRepository
	woFilesMysql  woFiles.WoFilesRepository
	oddsFileMysql oddsFiles.OddsFilesRepository
	redisConn     processRedis.RunRedis
}

// NewFileProcessorService : instantiate every connection we need to run current game service
func NewFileProcessorService(cfgs ...FileProcessorConfiguration) (*FileProcessorService, error) {
	// Create the seasonService
	os := &FileProcessorService{}
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

// WithMysqlLiveScoreRepository : instantiates mysql to connect to live score files interface
func WithMysqlLiveScoreRepository(connectionString string) FileProcessorConfiguration {
	return func(os *FileProcessorService) error {
		d, err := lsFilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.lsFilesMysql = d
		return nil
	}
}

// WithMysqlWinningOutcomesRepository : instantiates mysql to connect to winning outcome files interface
func WithMysqlWinningOutcomesRepository(connectionString string) FileProcessorConfiguration {
	return func(os *FileProcessorService) error {
		d, err := woFilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.woFilesMysql = d
		return nil
	}
}

// WithMysqlRawOddsRepository : instantiates mysql to connect to winning outcome files interface
func WithMysqlRawOddsRepository(connectionString string) FileProcessorConfiguration {
	return func(os *FileProcessorService) error {
		d, err := oddsFilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.oddsFileMysql = d
		return nil
	}
}

// WithRedisRepository : instantiates redis connections
func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) FileProcessorConfiguration {
	return func(os *FileProcessorService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// GetRawLiveScore : Queries raw live score from redis
func (s *FileProcessorService) GetRawLiveScore(ctx context.Context, countryLiveScoreSet string) error {

	lsFiles, err := s.lsFilesMysql.GetLsFiles(ctx)
	if err != nil {
		return fmt.Errorf("err : %v unable to query live score files", err)
	}

	for _, c := range lsFiles {

		err := s.redisConn.ZAdd(ctx, countryLiveScoreSet, c.LsFileID, c.LsFileName)
		if err != nil {
			return fmt.Errorf("err : %v unable to save  ls into redis sorted set", err)
		}
	}

	return nil
}

// GetRawWinningOutcomes : Queries raw winning outcomes from redis
func (s *FileProcessorService) GetRawWinningOutcomes(ctx context.Context, rawWinningOutcomeSortedSet string) error {

	woFiles, err := s.woFilesMysql.GetWinningOutcomeFiles(ctx)
	if err != nil {
		return fmt.Errorf("err : %v unable to query winning outcomes files", err)
	}

	for _, c := range woFiles {

		fullPath := fmt.Sprintf("%s/%s", c.WoDir, c.WoFileName)

		err := s.redisConn.ZAdd(ctx, rawWinningOutcomeSortedSet, c.WoFileID, fullPath)
		if err != nil {
			return fmt.Errorf("err : %v unable to save wo into redis sorted set", err)
		}
	}

	return nil
}

// GetRawWos : Queries raw wo from redis
func (s *FileProcessorService) GetRawWos(ctx context.Context, rawWinningOutcomeSortedSet, stmt string, ch chan DataTunnel) error {

	woFiles, err := s.woFilesMysql.GetWO(ctx, stmt)
	if err != nil {
		return fmt.Errorf("err : %v unable to query winning outcomes files", err)
	}

	for _, c := range woFiles {
		fts := DataTunnel{c.WoDir, c.WoFileName, c.WoFileID}
		ch <- fts
	}

	return nil
}

// SaveRawWos : Queries raw wo from redis
func (s *FileProcessorService) SaveRawWos(ctx context.Context, rawWinningOutcomeSortedSet, woDir, WoFileName, woFileID string) error {

	fullPath := fmt.Sprintf("%s/%s", woDir, WoFileName)

	log.Printf("saving %s in key %s", woFileID, rawWinningOutcomeSortedSet)

	err := s.redisConn.ZAdd(ctx, rawWinningOutcomeSortedSet, woFileID, fullPath)
	if err != nil {
		return fmt.Errorf("err : %v unable to save wo into redis sorted set", err)
	}

	return nil
}

// ReturnParentMatchIDs : returns all parent match ids in batches
func (s *FileProcessorService) ReturnZRangeData(ctx context.Context, zSetKey string, fetched int) ([]string, error) {
	data, err := s.redisConn.GetZRangeWithLimit(ctx, zSetKey, fetched)
	if err != nil {
		return data, err
	}
	return data, nil
}

// ReturnRawWO : returns all winning outcomes
func (s *FileProcessorService) ReturnRawWO(ctx context.Context, countryWoStagingSet, fullDirPath string, matchChan chan Job) error {

	jsonFile, err := os.Open(fullDirPath)
	defer jsonFile.Close()
	if err != nil {
		return fmt.Errorf("failed to open file %s | err : %v", fullDirPath, err)
	}

	log.Printf("Successfully opened : %s", fullDirPath)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to readAll %s | err : %v", fullDirPath, err)
	}

	// fullDirPath ==> /tz/generator/winning_outcomes/wo_31236351_3_18.txt

	dd := strings.Split(fullDirPath, "/")

	if len(dd) == 5 {

		ss := strings.Split(dd[4], "_") //wo_31104532_3_19.txt
		if len(ss) == 4 {

			var wo woFiles.RawWinningOutcome
			json.Unmarshal(byteValue, &wo)

			m := make(map[string]string)
			for _, i := range wo.WinningOutcomes {
				m[i.ParentMatchID] = i.ParentMatchID
			}

			// Create result for a specific parent match id.

			for _, i := range m {

				hh := woFiles.MatchWO{}
				hh.RoundNumberID = wo.RoundNumberID

				for _, x := range wo.WinningOutcomes {

					if i == x.ParentMatchID {

						j := woFiles.MWinningOutcomes{
							ParentMatchID: x.ParentMatchID,
							SubTypeID:     x.SubTypeID,
							OutcomeID:     x.OutcomeID,
							OutcomeName:   x.OutcomeName,
							Result:        x.Result,
						}

						hh.MWinningOutcomes = append(hh.MWinningOutcomes, j)

					}

				}

				for _, x := range wo.Results {

					if i == x.ParentMatchID {
						hh.AwayScore = x.AwayScore
						hh.HomeScore = x.HomeScore
					}

				}

				// Save this to redis
				jsonData, err := json.Marshal(hh)
				if err != nil {
					log.Println("Error:", err)
				}

				key := fmt.Sprintf("%s%s:%s", dd[1], "Wo", i)
				log.Printf("wo Key saved --> : %s", key)
				err = s.redisConn.Set(ctx, key, string(jsonData))
				if err != nil {
					return fmt.Errorf("err: %v | failed to save %s as set in redis ", err, key)
				}

				// Used for production
				err = s.redisConn.ZAdd(ctx, countryWoStagingSet, "1", key)
				if err != nil {
					return fmt.Errorf("err : %v unable to save wo into redis sorted set", err)
				}

			}

			for _, v := range m {

				fj := Job{
					JobType:  "query_odds",
					WorkStr:  v,
					WorkStr1: ss[1],
					WorkStr2: dd[1],
					WorkStr3: "",
				}
				matchChan <- fj

			}

			for _, v := range m {

				fj2 := Job{
					JobType:  "query_live_scores",
					WorkStr:  v,
					WorkStr1: ss[1],
					WorkStr2: dd[1],
					WorkStr3: "",
				}
				matchChan <- fj2

			}

			return nil

		}

	}

	return fmt.Errorf("data saved in wrong format :%s ", fullDirPath)

}

// ReturnOdds : returns all odds
func (s *FileProcessorService) ReturnOdds(ctx context.Context, countryOddsStagingSet, parentMatchID, roundNumberID, countryCode string, matchChan chan Job) error {

	// Query mysql to get
	odds, err := s.oddsFileMysql.GetAllOddsParentID(ctx, parentMatchID, countryCode)
	if err != nil {
		return fmt.Errorf("err %s | failed to get odds file in db roundID : %v | parent id %s", err, roundNumberID, parentMatchID)
	}

	for _, x := range odds {

		fullPath := fmt.Sprintf("%s/%s", x.FileDirectory, x.OddsFileName)
		jsonFile, err := os.Open(fullPath)
		defer jsonFile.Close()
		if err != nil {
			log.Printf("Failed to open file %s | err : %v", fullPath, err)
		}

		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to readAll %s | err : %v", fullPath, err)
		}

		// Push this Raw odds as a key in redis.
		key := fmt.Sprintf("%s%s%s", countryCode, "O:", parentMatchID)
		log.Printf("odds Key saved --> : %s", key)
		err = s.redisConn.Set(ctx, key, string(byteValue))
		if err != nil {
			return fmt.Errorf("err: %v | failed to save %s as set in redis ", err, key)
		}

		// Used for production
		err = s.redisConn.ZAdd(ctx, countryOddsStagingSet, "1", key)
		if err != nil {
			return fmt.Errorf("err : %v unable to save wo into redis sorted set", err)
		}

		return nil

	}

	return nil
}

func (s *FileProcessorService) ReturnRawLs(ctx context.Context, countryLsStagingSet, roundNumberID, countryCode string, matchChan chan Job) error {
	log.Printf("roundNumberID : %s | countryCode : %s", roundNumberID, countryCode)

	liveScores, err := s.lsFilesMysql.GetLiveScoreFileByExtID(ctx, roundNumberID, countryCode)
	if err != nil {
		return fmt.Errorf("err %s | failed to get live score file in db roundID : %v", err, roundNumberID)
	}

	for _, x := range liveScores {

		fullPath := fmt.Sprintf("%s/%s", x.LsDir, x.LsFileName)
		jsonFile, err := os.Open(fullPath)
		defer jsonFile.Close()
		if err != nil {
			return fmt.Errorf("failed to open file %s | err : %v", fullPath, err)
		}

		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to readAll %s | err :: %v", fullPath, err)
		}

		var ls lsFiles.LsData
		json.Unmarshal(byteValue, &ls)

		m := make(map[string]string)
		for _, i := range ls.LsMarkets {
			m[i.ParentMatchID] = i.ParentMatchID
		}

		for _, x := range m {

			hh := lsFiles.SingleLs{}
			hh.RoundNumberID = ls.RoundNumberID

			for _, r := range ls.LsMarkets {

				if x == r.ParentMatchID {

					hh.ParentMatchID = x

					ss := lsFiles.SMarkets{
						HomeScore:    r.HomeScore,
						AwayScore:    r.AwayScore,
						MinuteScored: r.MinuteScored,
					}

					hh.SMarkets = append(hh.SMarkets, ss)

				}
			}

			// Save this to redis
			jsonData, err := json.Marshal(hh)
			if err != nil {
				log.Println("Error:", err)
			}

			key := fmt.Sprintf("%s%s:%s", countryCode, "Ls", x)
			log.Printf("ls Key saved --> : %s", key)
			err = s.redisConn.Set(ctx, key, string(jsonData))
			if err != nil {
				return fmt.Errorf("err: %v | failed to save %s as set in redis ", err, key)
			}

			// Used for production
			err = s.redisConn.ZAdd(ctx, countryLsStagingSet, "1", key)
			if err != nil {
				return fmt.Errorf("err : %v unable to save wo into redis sorted set", err)
			}

			//log.Printf("Err : %v failed to save into %s", err, key)
		}
	}

	return nil

}

func (s *FileProcessorService) RemoveFromList(ctx context.Context, list, item string) error {
	_, err := s.redisConn.ZRem(ctx, list, item)
	if err != nil {
		return fmt.Errorf("err: %v | failed to remove %s from z range list %s", err, list, item)
	}
	return nil
}

func (s *FileProcessorService) AddToList(ctx context.Context, list, item string) error {

	err := s.redisConn.ZAdd(ctx, list, "1", item)
	if err != nil {
		return fmt.Errorf("err : %v unable to save  ls into redis sorted set", err)
	}

	return nil
}

func (s *FileProcessorService) SelectedWo(ctx context.Context, status string) ([]woFiles.WoFiles, error) {
	wo, err := s.woFilesMysql.GetPendingWo(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("err %v | failed to return raw winning outcomes", err)
	}
	return wo, nil
}

func (s *FileProcessorService) UpdateWoStatus(ctx context.Context, status, woFileID string) (int64, error) {
	wID, err := s.woFilesMysql.UpdateWoStatus(ctx, status, woFileID)
	if err != nil {
		return 0, fmt.Errorf("err %v | failed to return raw winning outcomes", err)
	}
	return wID, nil
}
