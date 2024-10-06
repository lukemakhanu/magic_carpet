package instantRedisServer

import (
	"context"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp/sharedHttpConf"
)

// InstantRedisServerConfiguration is an alias for a function that will take in a pointer to an InstantRedisServerService and modify it
type InstantRedisServerConfiguration func(os *InstantRedisServerService) error

// InstantRedisServerService is a implementation of the InstantRedisServerService
type InstantRedisServerService struct {
	redisConn processRedis.RunRedis
	httpConf  sharedHttp.SharedHttpConfRepository
}

func NewInstantRedisServerService(cfgs ...InstantRedisServerConfiguration) (*InstantRedisServerService, error) {
	// Create the NewClientAPIService
	os := &InstantRedisServerService{}
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

func WithRedisResultsRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) InstantRedisServerConfiguration {
	return func(os *InstantRedisServerService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// WithSharedHttpConfRepository : shared functions
func WithSharedHttpConfRepository() InstantRedisServerConfiguration {
	return func(os *InstantRedisServerService) error {
		cr, err := sharedHttpConf.New()
		if err != nil {
			return err
		}
		os.httpConf = cr
		return nil
	}
}

// AvailableGames : returns all available games
func (s *InstantRedisServerService) AvailableGames(c context.Context, allMatches string) (map[string]int, error) {
	m := make(map[string]int)
	data, err := s.redisConn.GetZRange(c, allMatches)
	if err != nil {
		log.Printf("Err : %v", err)
		return m, err
	}
	//log.Println("id : ", data)

	for k, v := range data {
		m[v] = k
	}

	return m, nil
}
