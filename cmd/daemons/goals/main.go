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
	"github.com/lukemakhanu/magic_carpet/internal/services/goal"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/goals/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/goals/"

var inProgress bool

func main() {
	InitConfig()

	sanitizedKeysSet := viper.GetString("redis-sorted-set.sanitizedSet")
	oddsSortedSet := viper.GetString("redis-sorted-set.odds")
	projectID := viper.GetString("redis-sorted-set.projectID")

	pg, err := goal.NewGoalService(
		goal.WithMysqlCheckMatchesRepository(viper.GetString("mySQL.live")),
		goal.WithMysqlGoalsRepository(viper.GetString("mySQL.live")),
		goal.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
		goal.WithSlowRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf(" **** Unable to start goal service **** : %s", err)
	}

	ctx := context.Background()

	matches, err := pg.ReturnOdds(ctx, oddsSortedSet)
	if err != nil {
		log.Printf("Err : %v failed to return odd for key :  %s", err, oddsSortedSet)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if !inProgress {
					inProgress = true

					SelectKeys(ctx, pg, oddsSortedSet, sanitizedKeysSet, matches, projectID)

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

// SelectKeys : used to save final odds
func SelectKeys(ctx context.Context, sm *goal.GoalService, oddsSortedSet, sanitizedKeysSet string, matches []string, projectID string) {

	defer func() {
		inProgress = false
		log.Printf("** done calling SelectKeys ** ")
	}()

	err := sm.SelectKeys(ctx, oddsSortedSet, sanitizedKeysSet, matches, projectID)
	if err != nil {
		log.Printf("Err : %v failed to generate sanitized", err)
	}
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("goals.logs"), viper.GetInt("log_setting.MaxSize"),
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
