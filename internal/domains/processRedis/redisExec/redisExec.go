// Copyright 2023 lukemakhanu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redisExec

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
)

var _ processRedis.RunRedis = (*RedisConfigs)(nil)

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
		IdleTimeout: 90 * time.Second,

		Dial: func() (redis.Conn, error) {

			/*c, err := redis.DialTimeout("tcp", redisHost, 100*time.Millisecond, 100*time.Millisecond, 100*time.Millisecond)
			if err != nil {
				return nil, err
			}
			return c, err*/

			c, err := redis.Dial("tcp", redisServer,
				redis.DialDatabase(dbNum),
				redis.DialConnectTimeout(1200*time.Millisecond),
				redis.DialReadTimeout(1200*time.Millisecond),
				redis.DialWriteTimeout(1200*time.Millisecond))
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

func (mr *RedisConfigs) Get(ctx context.Context, key string) (string, error) {
	conn := mr.r.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		return s, fmt.Errorf("Alert! this Key does not exist : %s | err : %w", key, err)
	} else if err != nil {
		return s, err
	}

	return s, nil
}

func (ns *RedisConfigs) Set(ctx context.Context, key, value string) error {
	conn := ns.r.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		return fmt.Errorf("Failed to create redis set : %s | err : %w", key, err)
	}

	return nil
}

func (ns *RedisConfigs) SetWithExpiry(ctx context.Context, key, value, expiry string) error {
	conn := ns.r.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value, "EX", expiry)
	if err != nil {
		return fmt.Errorf("error while doing EXPIRE command : %v", err)
	}

	return nil
}

func (ns *RedisConfigs) HSet(ctx context.Context, key, attribute, value string) error {
	conn := ns.r.Get()
	defer conn.Close()

	_, err := conn.Do("HSET", key, attribute, value)
	if err != nil {
		return fmt.Errorf("Failed to create redis hset : %s | err : %w", key, err)
	}

	return nil
}

func (ns *RedisConfigs) ZAdd(ctx context.Context, set, priority, value string) error {
	conn := ns.r.Get()
	defer conn.Close()

	_, err := conn.Do("zadd", set, priority, value)
	if err != nil {
		return fmt.Errorf("Error adding zset : %v", err)
	}

	return nil
}

// Delete : deletes key from redis
func (ns *RedisConfigs) Delete(ctx context.Context, key string) (interface{}, error) {
	conn := ns.r.Get()
	defer conn.Close()

	response, err := conn.Do("DEL", key)
	if err != nil {
		return nil, err
	}
	return response, nil
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

func (mr *RedisConfigs) ZRem(ctx context.Context, nameOfSet string, val string) (interface{}, error) {
	conn := mr.r.Get()
	defer conn.Close()
	response, err := conn.Do("ZREM", nameOfSet, val)
	if err != nil {
		return nil, fmt.Errorf("Failed to delete zmember key : %s from set %s | err : %w", nameOfSet, val, err)
	}
	return response, nil
}

func (mr *RedisConfigs) HGet(ctx context.Context, key, field string) (string, error) {
	conn := mr.r.Get()
	defer conn.Close()

	reply, err := redis.String(conn.Do("HGET", key, field))
	if err != nil {
		return reply, fmt.Errorf("Failed to get match item %s | err : %v", field, err)
	}
	return reply, nil
}

// ZRevRange :
func (mr *RedisConfigs) ZRevRange(ctx context.Context, set string) ([]string, error) {
	conn := mr.r.Get()
	defer conn.Close()

	s, err := redis.Strings(conn.Do("ZREVRANGE", set, 0, -1))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (ns *RedisConfigs) GetZRevRangeWithLimit(ctx context.Context, set string, fetched int) ([]string, error) {
	conn := ns.r.Get()
	defer conn.Close()

	s, err := redis.Strings(conn.Do("ZREVRANGE", set, 0, fetched))
	if err != nil {
		return nil, err
	}
	return s, nil
}

// HmSet :

// m := map[string]string{
// 	"title":  "Example2",
// 	"author": "Steve",
// 	"body":   "Map",
// }

func (mr *RedisConfigs) HmSet(ctx context.Context, set string, m map[string]string) error {
	conn := mr.r.Get()
	defer conn.Close()

	if _, err := conn.Do("HMSET", redis.Args{}.Add(set).AddFlat(m)...); err != nil {
		return fmt.Errorf("Failed to save in hmSet %s | err : %v", set, err)
	}

	return nil
}

func (mr *RedisConfigs) SortedSetLen(ctx context.Context, key string) (int, error) {
	conn := mr.r.Get()
	defer conn.Close()

	len, err := redis.Int(conn.Do("ZCARD", key))
	if err != nil {
		return 0, fmt.Errorf("Err %v failed to return len of %s or the key does not exist", err, key)
	}

	return len, nil
}
