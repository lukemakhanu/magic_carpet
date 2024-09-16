package rExec

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/slowRedis"
)

var _ slowRedis.SlowRedis = (*RedisConfigs)(nil)

type RedisConfigs struct {
	redisServer string
	dbNum       int
	maxIdle     int
	maxActive   int
	idleTimeout time.Duration
	r           *redis.Pool
}

// Used to instantiate redis connection pool
func New(redisServer string, dbNum, maxIdle, maxActive int, idleTimeout time.Duration) (*RedisConfigs, error) {

	if redisServer == "" {
		return nil, fmt.Errorf("redisServer not set")
	}

	if dbNum < 0 || dbNum > 12 {
		return nil, fmt.Errorf("invalid db number provided")
	}

	if maxIdle <= 0 {
		return nil, fmt.Errorf("maxIdle not set")
	}

	if maxActive <= 0 {
		return nil, fmt.Errorf("maxActive not set")
	}

	if idleTimeout <= 0 {
		return nil, fmt.Errorf("Timeout not set")
	}

	redisPool := &redis.Pool{

		MaxIdle:     maxIdle,
		MaxActive:   maxIdle,
		IdleTimeout: 120 * time.Second,

		Dial: func() (redis.Conn, error) {

			c, err := redis.Dial("tcp", redisServer,
				redis.DialDatabase(dbNum),
				redis.DialConnectTimeout(120*time.Second),
				redis.DialReadTimeout(120*time.Second),
				redis.DialWriteTimeout(120*time.Second))
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &RedisConfigs{
		r: redisPool,
	}, nil
}

func (ns *RedisConfigs) GetZRange(ctx context.Context, set string) ([]string, error) {
	conn := ns.r.Get()
	defer conn.Close()

	s, err := redis.Strings(conn.Do("ZRANGE", set, 0, -1))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (ns *RedisConfigs) GetZRangeWithLimit(ctx context.Context, set string, fetched int) ([]string, error) {
	conn := ns.r.Get()
	defer conn.Close()

	s, err := redis.Strings(conn.Do("ZRANGE", set, 0, fetched))
	if err != nil {
		return nil, err
	}
	return s, nil
}
