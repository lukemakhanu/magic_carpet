package prepareMatch

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/cleanUps"
	"github.com/lukemakhanu/magic_carpet/internal/domains/cleanUps/cleanUpsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goals"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/usedMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/usedMatches/usedMatchesMysql"
)

type PrepareMatchConfiguration func(os *PrepareMatchService) error

// PrepareMatchService is a implementation of the PrepareMatchService
type PrepareMatchService struct {
	cleanUpMysql   cleanUps.CleanUpsRepository
	usedMatchMysql usedMatches.UsedMatchesRepository
	redisConn      processRedis.RunRedis
}

// NewPrepareMatchService : instantiate every connection we need to run current game service
func NewPrepareMatchService(cfgs ...PrepareMatchConfiguration) (*PrepareMatchService, error) {
	// Create the seasonService
	os := &PrepareMatchService{}
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

func WithMysqlUsedMatchesRepository(connectionString string) PrepareMatchConfiguration {
	return func(os *PrepareMatchService) error {
		d, err := usedMatchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.usedMatchMysql = d
		return nil
	}
}

// WithMysqlCleanUpsRepository : returns cleanups
func WithMysqlCleanUpsRepository(connectionString string) PrepareMatchConfiguration {
	return func(os *PrepareMatchService) error {
		d, err := cleanUpsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.cleanUpMysql = d
		return nil
	}
}

// WithRedisRepository : instantiates redis connections
func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) PrepareMatchConfiguration {
	return func(os *PrepareMatchService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// TodaysGame : cleanup data to be used to generate games today.
func (s *PrepareMatchService) TodaysGame(ctx context.Context) error {

	// select to see if the data is already saved
	data, err := s.cleanUpMysql.LastCleanUps(ctx)
	if err != nil {
		return fmt.Errorf("err : %v failed to return last cleanup ", err)
	}

	if len(data) > 0 {
		// Proceed with further processing

		cleanDate := data[0].CleanUpDate
		status := data[0].Status
		cleanupID := data[0].CleanUpID

		current_time := time.Now().Local()
		log.Println("system date ", current_time.Format("2006-01-02"), "Db date", cleanDate, "status ", status)

		//if cleanDate == current_time.Format("2006-01-02") && status == "pending" {
		if status == "pending" {

			log.Printf("All is set >> clean up data")

			// Get all records that havent been used from mysql

			dd, err := s.AvailableMatches(ctx)
			if err != nil {
				return fmt.Errorf("err : %v failed to return available matches ", err)
			}

			for k, v := range dd {

				// loop through to get each match
				newKeyName := fmt.Sprintf("%s_%s", "CL", k)

				// Start by deleting previous records

				_, err := s.redisConn.Delete(ctx, newKeyName)
				if err != nil {
					return fmt.Errorf("err : %v on deleting key %s ", err, newKeyName)
				}

				for _, c := range v {

					err := s.redisConn.ZAdd(ctx, newKeyName, c.GoalID, c.MatchID)
					if err != nil {
						return fmt.Errorf("err : %v adding new cleaned data into sorted set ", err)
					}

				}

			}

			// update the record as processed
			status := "cleaned"
			updated, err := s.cleanUpMysql.UpdateCleanUps(ctx, cleanupID, status)
			if err != nil {
				return fmt.Errorf("err : %v updating cleanup data ", err)
			}
			log.Printf("Updated record %d", updated)

		} else if cleanDate == current_time.Format("2006-01-02") && status == "cleaned" {

			log.Printf("process key creates keys that will later update to processed")

		} else if cleanDate == current_time.Format("2006-01-02") && status == "processed" {

			// Insert one record to be used for the next day.
			status := "pending"
			data, err := s.cleanUpMysql.CleanUpsByStatus(ctx, status)
			if err != nil {
				return fmt.Errorf("err : %v failed to return last cleanup ", err)
			}

			if len(data) > 0 {
				log.Printf("**** Tomorrows data already saved ****")
			} else {

				projectID := "1"
				cleanUps, err := cleanUps.NewCleanUps(projectID, status)
				if err != nil {
					return fmt.Errorf("err : %v failed to instantiate cleanup ", err)
				}

				cleanupID, err := s.cleanUpMysql.SaveForTomorrow(ctx, *cleanUps)
				if err != nil {
					return fmt.Errorf("err : %v failed to save new cleanup data ", err)
				}

				log.Printf("cleanupID : %d", cleanupID)

			}

		}

	} else {

		// Create new record
		projectID := "1"
		status := "pending"
		cleanUps, err := cleanUps.NewCleanUps(projectID, status)
		if err != nil {
			return fmt.Errorf("err : %v failed to instantiate cleanup ", err)
		}

		cleanupID, err := s.cleanUpMysql.Save(ctx, *cleanUps)
		if err != nil {
			return fmt.Errorf("err : %v failed to save new cleanup data ", err)
		}

		log.Printf("cleanupID : %d", cleanupID)
	}

	return nil
}

func (s *PrepareMatchService) AvailableMatches(ctx context.Context) (map[string][]goals.Goals, error) {

	categories := []string{
		"SANITIZED_ODDS_0",
		"SANITIZED_ODDS_1_a",
		"SANITIZED_ODDS_1_h",
		"SANITIZED_ODDS_2_gg_d",
		"SANITIZED_ODDS_2_ng_a",
		"SANITIZED_ODDS_2_ng_h",
		"SANITIZED_ODDS_3_gg_a",
		"SANITIZED_ODDS_3_gg_h",
		"SANITIZED_ODDS_3_ng_a",
		"SANITIZED_ODDS_3_ng_h",
		"SANITIZED_ODDS_4_gg_a",
		"SANITIZED_ODDS_4_gg_h",
		"SANITIZED_ODDS_4_ng_a",
		"SANITIZED_ODDS_4_ng_h",
		"SANITIZED_ODDS_4_gg_d",
		"SANITIZED_ODDS_5_gg_a",
		"SANITIZED_ODDS_5_gg_h",
		"SANITIZED_ODDS_5_ng_a",
		"SANITIZED_ODDS_5_ng_h",
		"SANITIZED_ODDS_6_gg_a",
		"SANITIZED_ODDS_6_gg_d",
		"SANITIZED_ODDS_6_gg_h",
		"SANITIZED_ODDS_6_ng_a",
		"SANITIZED_ODDS_6_ng_h",
	}

	m := make(map[string][]goals.Goals)

	for _, d := range categories {

		data, err := s.usedMatchMysql.GetAvailable(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("failed to return available matches : %v", err)
		}

		m[d] = data

		log.Printf("category %s >> count %d", d, len(data))

	}

	return m, nil

}
