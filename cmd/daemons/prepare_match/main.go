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
	"github.com/lukemakhanu/magic_carpet/internal/services/prepareMatch"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/prepare_match/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/prepare_match/"

var inProgress bool

func main() {
	InitConfig()

	pg, err := prepareMatch.NewPrepareMatchService(
		prepareMatch.WithMysqlUsedMatchesRepository(viper.GetString("mySQL.live")),
		prepareMatch.WithMysqlCleanUpsRepository(viper.GetString("mySQL.live")),
		prepareMatch.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf(" **** Unable to start prepare match service ***** : %s", err)
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

					CleanGames(ctx, pg)

				} else {
					log.Printf("**** CleanGames in process **** %v.\n", t)
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

// CleanGames : used to save final odds
func CleanGames(ctx context.Context, sm *prepareMatch.PrepareMatchService) {

	defer func() {
		inProgress = false
		log.Printf("** Done calling TodaysGame ** ")
	}()

	err := sm.TodaysGame(ctx)
	if err != nil {
		log.Printf("Err : %v failed to cleanup todays games >> ", err)
	}
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("prepare_match.logs"), viper.GetInt("log_setting.MaxSize"),
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
