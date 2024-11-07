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
	"github.com/lukemakhanu/magic_carpet/internal/services/productionKey"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/production_keys/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/production_keys/"

var inProgress bool

func main() {
	InitConfig()

	oddsSortedSet := viper.GetString("redis-sorted-set.odds")
	sanitizedKeysSet := viper.GetString("redis-sorted-set.sanitizedKeysSet")
	minimumRequired := viper.GetInt("redis-sorted-set.minimumRequired")

	pg, err := productionKey.NewProcessKeyService(
		productionKey.WithMysqlMatchesRepository(viper.GetString("mySQL.live")),
		productionKey.WithMysqlSeasonWeeksRepository(viper.GetString("mySQL.live")),
		productionKey.WithMysqlCheckMatchesRepository(viper.GetString("mySQL.live")),
		productionKey.WithMysqlMrsRepository(viper.GetString("mySQL.live")),
		productionKey.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf(" **** Unable to start production keys service ***** : %s", err)
	}

	ctx := context.Background()

	//matchChan := make(chan processFile.Job, 300)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if !inProgress {
					inProgress = true

					SaveFinalOddsKey(ctx, pg, oddsSortedSet, sanitizedKeysSet, minimumRequired)

				} else {
					log.Printf("**** SaveFinalOddsKey2 in process **** %v.\n", t)
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

// SaveFinalOddsKey2 : used to save final odds
func SaveFinalOddsKey(ctx context.Context, sm *productionKey.ProcessKeyService, oddsSortedSet, sanitizedKeysSet string, minimumRequired int) {

	defer func() {
		inProgress = false
		log.Printf("******* Done calling SaveFinalOddsKey2 ****** ")
	}()

	// get a list of match categories

	err := sm.GetUpcomingSeasonWeeks(ctx, oddsSortedSet, sanitizedKeysSet, minimumRequired)
	if err != nil {
		log.Printf("Err : %v failed to save production odds >> ", err)
	}
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("production_keys.logs"), viper.GetInt("log_setting.MaxSize"),
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

// 2089) "keO:47455339" ZRANGE NEW_STAGING_ODDS_USED 0 -1

// >>> 370645) "tzO:31475633" (last)  >>> 1) "keO:47455341"(first)  ZRANGE NEW_STAGING_ODDS 0 -1
