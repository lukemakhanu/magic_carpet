// Package main runs the tavern and performs an Order
package main

import (
	"context"
	"time"

	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/services/goalPattern"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/goal_patterns/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/goal_patterns/"

var inProgress bool

func main() {
	InitConfig()

	pg, err := goalPattern.NewGoalPatternService(
		goalPattern.WithMysqlMrsRepository(viper.GetString("mySQL.live")),
		goalPattern.WithMysqGoalPatternsRepository(viper.GetString("mySQL.live")),
		goalPattern.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
		goalPattern.WithSlowRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf(" **** Unable to start goal total service **** : %s", err)
	}

	ctx := context.Background()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if !inProgress {
					inProgress = true

					GoalPattern(ctx, pg)

				} else {
					log.Printf("**** SelectKeys in process **** %v.\n", t)
				}
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting", s)
}

// GoalPattern : used to save final odds
func GoalPattern(ctx context.Context, sm *goalPattern.GoalPatternService) {

	defer func() {
		inProgress = false
		log.Printf("** done calling goalPattern ** ")
	}()

	err := sm.ProcessGoalPattern(ctx)
	if err != nil {
		log.Printf("Err : %v failed to generate sanitized", err)
	}
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("goal_patterns.logs"), viper.GetInt("log_setting.MaxSize"),
		viper.GetInt("log_setting.MaxBackups"), viper.GetInt("log_setting.MaxAge"),
		viper.GetBool("log_setting.Compress"))
}

func logUtils(logDirectory string, maxSize int, maxBackups int, maxAge int, compress bool) {
	log.SetOutput(&lumberjack.Logger{
		Filename:   logDirectory,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	})
}

func configUtils(addConfigPathLive string, addConfigPathLocal string) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	viper.SetDefault("host", "localhost")
	viper.SetConfigName("config")
	viper.AddConfigPath(addConfigPathLive)
	viper.AddConfigPath(addConfigPathLocal)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Error : %v", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
	})
}
