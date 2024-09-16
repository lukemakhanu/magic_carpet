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
	"github.com/lukemakhanu/magic_carpet/internal/services/prepareKey"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/prepare_keys/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/prepare_keys/"

var inProgress bool
var inProgress2 bool
var inProgress3 bool

func main() {
	InitConfig()

	sanitizedKeysSet := viper.GetString("redis-sorted-set.sanitizedSet")
	oddsSortedSet := viper.GetString("redis-sorted-set.odds")

	pg, err := prepareKey.NewPrepareKeyService(
		prepareKey.WithMysqlCheckMatchesRepository(viper.GetString("mySQL.live")),
		prepareKey.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
		prepareKey.WithSlowRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf(" **** Unable to start fp prepare-keys **** : %s", err)
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

					SelectKeys(ctx, pg, oddsSortedSet, sanitizedKeysSet, matches)

				} else {
					log.Printf("**** SelectKeys running **** %v.\n", t)
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
func SelectKeys(ctx context.Context, sm *prepareKey.PrepareKeyService, oddsSortedSet, sanitizedKeysSet string, matches []string) {

	defer func() {
		inProgress = false
		log.Printf("*** done calling SelectKeys ***")
	}()

	err := sm.SelectKeys(ctx, oddsSortedSet, sanitizedKeysSet, matches)
	if err != nil {
		log.Printf("Err : %v failed to generate sanitized", err)
	}
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("prepare_keys.logs"), viper.GetInt("log_setting.MaxSize"),
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
