package v1Team

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/lsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/readFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/readFiles/readDir"
)

type Job struct {
	JobType  string
	WorkStr  string
	WorkStr1 string
	WorkStr2 string
	WorkStr3 string
}

// V1TeamConfiguration is an alias for a function that will take in a pointer to an V1TeamService and modify it
type V1TeamConfiguration func(os *V1TeamService) error

// V1TeamService is a implementation of the V1TeamService
type V1TeamService struct {
	redisConn processRedis.RunRedis
	dirReader readFiles.DirectoryReader
}

// NewV1ProcessFileService : instantiate every connection we need to run current game service
func NewV1ProcessFileService(cfgs ...V1TeamConfiguration) (*V1TeamService, error) {
	// Create the seasonService
	os := &V1TeamService{}
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

// WithRedisRepository : instantiates redis connections
func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) V1TeamConfiguration {
	return func(os *V1TeamService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

func WithDirectoryReaderRepository(directory, combination string) V1TeamConfiguration {
	return func(os *V1TeamService) error {
		// Create Matches repo
		d, err := readDir.New(directory, combination)
		if err != nil {
			return err
		}
		os.dirReader = d
		return nil
	}
}

// SaveTeams : save teams into redis
func (s *V1TeamService) SaveL1Teams(ctx context.Context, fileDir, teamsList string) error {

	allFiles, err := s.dirReader.ReadTeamDirectory(ctx)
	if err != nil {
		return fmt.Errorf("Err : %v failed to read files", err)
	}

	l1 := 1
	l2 := 1
	l3 := 1
	l4 := 1

	for _, f := range allFiles {

		log.Printf("competitionID : %s | count %s | seasonID %s |  name %s",
			f.CompetitionID, f.Count, f.SeasonID, f.Name)

		if f.CompetitionID == "1" {

			// Number of seasons (18)
			if l1 < 19 {

				// 380 games per season with 38 season weeks

				data, err := s.ReadFile(ctx, fileDir, f.Name)
				if err != nil {
					log.Printf("Err : %v failed to read file", err)
				} else {

					if len(data) > 0 {

						var t lsFiles.TeamsH2H
						json.Unmarshal([]byte(data), &t)

						log.Printf("competitionID %s | count %s", t.CompetitionID, t.Count)

						for _, r := range t.H2H {
							log.Printf("matchDay %s | teams : %s", r.MatchDay, r.Teams)

							keyName := fmt.Sprintf("%s_%d_%s", f.CompetitionID, l1, r.MatchDay)
							err = s.redisConn.ZAdd(ctx, keyName, r.MatchDay, r.Teams)
							if err != nil {
								log.Printf("Err : %v unable to save teams into redis sorted set", err)
							}

							err = s.redisConn.ZAdd(ctx, teamsList, "1", keyName)
							if err != nil {
								log.Printf("Err : %v unable to save team list into redis sorted set", err)
							}
						}

						l1++

					}

				}

			}

		} else if f.CompetitionID == "2" {

			// Number of seasons (18)
			if l2 < 19 {

				// 380 games per season with 38 season weeks

				data, err := s.ReadFile(ctx, fileDir, f.Name)
				if err != nil {
					log.Printf("Err : %v failed to read file", err)
				} else {

					if len(data) > 0 {

						var t lsFiles.TeamsH2H
						json.Unmarshal([]byte(data), &t)

						log.Printf("competitionID %s | count %s", t.CompetitionID, t.Count)

						for _, r := range t.H2H {
							log.Printf("matchDay %s | teams : %s", r.MatchDay, r.Teams)

							keyName := fmt.Sprintf("%s_%d_%s", f.CompetitionID, l2, r.MatchDay)
							err = s.redisConn.ZAdd(ctx, keyName, r.MatchDay, r.Teams)
							if err != nil {
								log.Printf("Err : %v unable to save teams into redis sorted set", err)
							}

							err = s.redisConn.ZAdd(ctx, teamsList, "2", keyName)
							if err != nil {
								log.Printf("Err : %v unable to save team list into redis sorted set", err)
							}

						}

						l2++

					}
				}

			}

		} else if f.CompetitionID == "3" {

			// Number of seasons (18)
			if l3 < 19 {

				// 342 games per season with 38 season weeks

				data, err := s.ReadFile(ctx, fileDir, f.Name)
				if err != nil {
					log.Printf("Err : %v failed to read file", err)
				} else {

					if len(data) > 0 {

						var t lsFiles.TeamsH2H
						json.Unmarshal([]byte(data), &t)

						log.Printf("competitionID %s | count %s", t.CompetitionID, t.Count)

						for _, r := range t.H2H {
							log.Printf("matchDay %s | teams : %s", r.MatchDay, r.Teams)

							keyName := fmt.Sprintf("%s_%d_%s", f.CompetitionID, l3, r.MatchDay)
							err = s.redisConn.ZAdd(ctx, keyName, r.MatchDay, r.Teams)
							if err != nil {
								log.Printf("Err : %v unable to save teams into redis sorted set", err)
							}

							err = s.redisConn.ZAdd(ctx, teamsList, "3", keyName)
							if err != nil {
								log.Printf("Err : %v unable to save team list into redis sorted set", err)
							}

						}

						l3++
					}
				}

			}

		} else if f.CompetitionID == "4" {

			// Number of seasons (18)
			if l4 < 19 {

				// 380 games per season with 38 season weeks

				data, err := s.ReadFile(ctx, fileDir, f.Name)
				if err != nil {
					log.Printf("Err : %v failed to read file", err)
				} else {

					if len(data) > 0 {

						var t lsFiles.TeamsH2H
						json.Unmarshal([]byte(data), &t)

						log.Printf("competitionID %s | count %s", t.CompetitionID, t.Count)

						for _, r := range t.H2H {
							log.Printf("matchDay %s | teams : %s", r.MatchDay, r.Teams)

							keyName := fmt.Sprintf("%s_%d_%s", f.CompetitionID, l4, r.MatchDay)
							err = s.redisConn.ZAdd(ctx, keyName, r.MatchDay, r.Teams)
							if err != nil {
								log.Printf("Err : %v unable to save teams into redis sorted set", err)
							}

							err = s.redisConn.ZAdd(ctx, teamsList, "4", keyName)
							if err != nil {
								log.Printf("Err : %v unable to save team list into redis sorted set", err)
							}
						}

						l4++
					}
				}

			}

		} else {
			log.Printf("competitionID %s | count %s", f.CompetitionID, f.Count)
		}

	}

	return nil
}

func (s *V1TeamService) ReadFile(ctx context.Context, filepath, fileName string) (string, error) {
	var str = ""
	fullPath := fmt.Sprintf("%s/%s", filepath, fileName)
	jsonFile, err := os.Open(fullPath)
	defer jsonFile.Close()
	if err != nil {
		return str, fmt.Errorf("Failed to open file %s | err : %v", fullPath, err)
	}

	log.Printf("Successfully opened : %s", fullPath)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return str, fmt.Errorf("Failed to readAll %s | err : %v", fullPath, err)
	}

	return string(byteValue), nil
}
