// Package main runs the tavern and performs an Order
package main

import (
	"context"
	"sync"
	"time"

	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/services/saveFileRedis"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/file_processors/save_file_in_redis/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/file_processors/save_file_in_redis/"

var inProgress bool

func main() {
	InitConfig()

	rawWinningOutcomeSortedSet := viper.GetString("redis-sorted-set.rawWinningOutcome")

	pg, err := saveFileRedis.NewFileProcessorService(
		saveFileRedis.WithMysqlLiveScoreRepository(viper.GetString("mySQL.live")),
		saveFileRedis.WithMysqlWinningOutcomesRepository(viper.GetString("mySQL.live")),
		saveFileRedis.WithMysqlRawOddsRepository(viper.GetString("mySQL.live")),
		saveFileRedis.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf("Unable to start file processor service ::: %s", err)
	}

	dataChan := make(chan saveFileRedis.DataTunnel, 1000)

	ctx := context.Background()

	var w sync.WaitGroup
	w.Add(300)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if !inProgress {
					inProgress = true

					QueryWinningOutcomes(ctx, dataChan, pg, rawWinningOutcomeSortedSet)
				} else {
					log.Printf("GetPendingBets still running.. %v.\n", t)
				}
			}
		}
	}()

	for i := 1; i <= 300; i++ {
		go func(i int, ci <-chan saveFileRedis.DataTunnel) {
			j := 1
			for v := range ci {

				log.Printf("goroutine :%d | num %d | WoDir : %s | WoFileID : %s | WoFileName : %s ", i, j, v.WoDir, v.WoFileID, v.WoFileName)
				j += 1

				err := pg.SaveRawWos(ctx, rawWinningOutcomeSortedSet, v.WoDir, v.WoFileName, v.WoFileID)
				if err != nil {
					log.Printf("err : %v", err)
				}

			}

			w.Done()

		}(i, dataChan)
	}

	w.Wait()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

func QueryWinningOutcomes(ctx context.Context, ch chan saveFileRedis.DataTunnel, sm *saveFileRedis.FileProcessorService, rawWinningOutcomeSortedSet string) {

	defer func() {
		inProgress = false
		log.Printf("Done QueryWinningOutcomes.. ")
	}()

	queries := []string{
		"select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 0 and wo_file_id < 100000 ",
		"select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 100000 and wo_file_id < 200000 ",
		"select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 200000 and wo_file_id < 300000 ",
		"select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 300000 and wo_file_id < 400000 ",
		"select wo_file_id,wo_file_name,wo_dir,country,ext_id,project_id,competition_id,created,modified from winning_outcome_files where wo_file_id > 400000 and wo_file_id < 500000 ",
	}

	for _, stmt := range queries {

		log.Printf(" query running ... %s", stmt)

		err := sm.GetRawWos(ctx, rawWinningOutcomeSortedSet, stmt, ch)
		if err != nil {
			log.Printf("Err : failed to save raw wo : %v", err)
		}

	}

}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("file_processor.logs"), viper.GetInt("log_setting.MaxSize"),
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
