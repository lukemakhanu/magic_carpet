package goalPattern

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns/goalPatternsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs/mrsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/slowRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/slowRedis/rExec"
)

type Job struct {
	JobType  string
	WorkStr  string
	WorkStr1 string
	WorkStr2 string
	WorkStr3 string
}

// GoalPatternConfiguration is an alias for a function that will take in a pointer to an GoalPatternService and modify it
type GoalPatternConfiguration func(os *GoalPatternService) error

// GoalPatternService is a implementation of the GoalPatternService
type GoalPatternService struct {
	redisConn         processRedis.RunRedis
	slowRedisConn     slowRedis.SlowRedis
	mrsMysql          mrs.MrsRepository
	goalPatternsMysql goalPatterns.GoalPatternsRepository
}

// NewGoalPatternService : instantiate every connection we need to run current game service
func NewGoalPatternService(cfgs ...GoalPatternConfiguration) (*GoalPatternService, error) {
	// Create the seasonService
	os := &GoalPatternService{}
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
func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) GoalPatternConfiguration {
	return func(os *GoalPatternService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// WithSlowRedisRepository : instantiates redis connections
func WithSlowRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) GoalPatternConfiguration {
	return func(os *GoalPatternService) error {
		d, err := rExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.slowRedisConn = d
		return nil
	}
}

// WithMysqlMrsRepository : instantiates mysql to connect to matches interface
func WithMysqlMrsRepository(connectionString string) GoalPatternConfiguration {
	return func(os *GoalPatternService) error {
		d, err := mrsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.mrsMysql = d
		return nil
	}
}

// WithMysqlMrsRepository : instantiates mysql to connect to matches interface
func WithMysqGoalPatternsRepository(connectionString string) GoalPatternConfiguration {
	return func(os *GoalPatternService) error {
		d, err := goalPatternsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.goalPatternsMysql = d
		return nil
	}
}

// ProcessGoalPattern : used to select keys to be used later.
func (s *GoalPatternService) ProcessGoalPattern(ctx context.Context) error {

	ss := []string{"1", "2", "3", "4"}

	for _, c := range ss {

		m := make(map[int]int)

		data, err := s.mrsMysql.GoalPatterns(ctx, c)
		if err != nil {
			return fmt.Errorf("err : %v failed to return goal patterns from db ", err)
		}

		for _, x := range data {
			m[x.RoundNumberID] = x.RoundNumberID
		}

		// push this ids into list

		comp := []int{}
		for _, v := range m {
			comp = append(comp, v)
		}

		dd := s.GoalChunks(comp, 38)

		for a, x := range dd {
			log.Println("competition ", c, " id ", a, " >> xx selected ", x)

			if len(x) == 38 {

				ff := []string{}
				for _, r := range x {
					gg := fmt.Sprintf("%d", r)
					ff = append(ff, gg)
				}

				log.Printf("proceed as we have the correct number of games")

				roundNumberIDs := strings.Join(ff, ",")
				log.Printf("competitionID %s | number %d | games %s", c, a, roundNumberIDs)

				// // Save into database
				seasonID := fmt.Sprintf("%d", a)
				dd, err := goalPatterns.NewGoalPatterns(seasonID, roundNumberIDs, c)
				if err != nil {
					return fmt.Errorf("err : %v failed to initialize goal pattern ", err)
				}

				lastID, err := s.goalPatternsMysql.Save(ctx, *dd)
				if err != nil {
					return fmt.Errorf("err : %v failed to save a goal pattern ", err)
				}

				log.Printf("lastID : %d", lastID)

			} else {
				log.Printf("Skip, we have less games...")
			}

		}

	}

	return nil
}

// GoalChunks :
func (s *GoalPatternService) GoalChunks(slice []int, chunkSize int) [][]int {
	var chunks [][]int
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
