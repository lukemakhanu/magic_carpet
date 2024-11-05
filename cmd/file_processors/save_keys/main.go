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

var addConfigPathLive = "/apps/go/magic_carpet/file_processors/cmd/save_keys/"
var addConfigPathLocal = "/apps/go/magic_carpet/file_processors/cmd/save_keys/"

var inProgress bool

func main() {
	InitConfig()

	//rawWinningOutcomeSortedSet := viper.GetString("redis-sorted-set.rawWinningOutcome")
	woSortedSet := viper.GetString("redis-sorted-set.winningOutcome")
	liveScoreSortedSet := viper.GetString("redis-sorted-set.liveScore")
	oddsSortedSet := viper.GetString("redis-sorted-set.odds")

	pg, err := saveFileRedis.NewFileProcessorService(
		saveFileRedis.WithMysqlLiveScoreRepository(viper.GetString("mySQL.live")),
		saveFileRedis.WithMysqlWinningOutcomesRepository(viper.GetString("mySQL.live")),
		saveFileRedis.WithMysqlRawOddsRepository(viper.GetString("mySQL.live")),
		saveFileRedis.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf("Unable to start save file service ::: %s", err)
	}

	ctx := context.Background()

	var w sync.WaitGroup
	w.Add(300)

	matchChan := make(chan saveFileRedis.Job, 10000)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if !inProgress {
					inProgress = true
					SaveWinningOutcomes(ctx, matchChan, pg)
				} else {
					log.Printf("in process.. %v.\n", t)
				}
			}
		}
	}()

	for i := 1; i <= 300; i++ {
		go func(i int, ci <-chan saveFileRedis.Job) {
			for v := range ci {
				//time.Sleep(time.Millisecond)
				log.Printf("job type : %s | work : %s", v.JobType, v.WorkStr)

				if v.JobType == "query_match_id" {

					err := pg.ReturnRawWO(ctx, woSortedSet, v.WorkStr, matchChan)
					if err != nil {
						log.Printf("Err : failed to save parent match ids : %v", err)
					}

				} else if v.JobType == "query_odds" {

					err := pg.ReturnOdds(ctx, oddsSortedSet, v.WorkStr, v.WorkStr1, v.WorkStr2, matchChan)
					if err != nil {
						log.Printf("Err : failed to save parent match ids ::: %v", err)
					}

				} else if v.JobType == "query_live_scores" {

					//log.Printf("RoundID is : %s | countryCode %s", v.WorkStr1, v.WorkStr2)

					err := pg.ReturnRawLs(ctx, liveScoreSortedSet, v.WorkStr1, v.WorkStr2, matchChan)
					if err != nil {
						log.Printf("Err : failed to save parent match ids : %v", err)
					}

				} else {
					log.Printf("Wrong job type send : %s", v.JobType)
				}

			}

			w.Done()

		}(i, matchChan)
	}

	w.Wait()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

func SaveWinningOutcomes(ctx context.Context, matchChan chan saveFileRedis.Job, sm *saveFileRedis.FileProcessorService) {

	defer func() {
		inProgress = false
		log.Printf("Done calling  ....")
	}()

	status := "pending"
	data, err := sm.SelectedWo(ctx, status)
	if err != nil {
		log.Printf("Err : %v", err)
	}

	for _, wo := range data {

		fullPath := fmt.Sprintf("%s/%s", wo.WoDir, wo.WoFileName)

		// Update this as processed.
		processedStatus := "processed"
		sm.UpdateWoStatus(ctx, processedStatus, wo.WoFileID)

		fj := saveFileRedis.Job{
			JobType:  "query_match_id",
			WorkStr:  fullPath,
			WorkStr1: "",
			WorkStr2: "",
			WorkStr3: "",
		}
		matchChan <- fj

	}

}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("save_keys.logs"), viper.GetInt("log_setting.MaxSize"),
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
