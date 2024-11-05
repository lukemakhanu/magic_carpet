// Package main runs the tavern and performs an Order
package main

import (
	"context"

	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/services/saveFile"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/file_processors/save_file_names/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/file_processors/save_file_names/"

func main() {
	InitConfig()

	// Live Score
	liveScoreService, err := saveFile.NewSaveFileService(
		saveFile.WithMysqlLiveScoreFilesRepository(viper.GetString("mySQL.live")),
		saveFile.WithDirectoryReaderRepository(viper.GetString("ke.ls_directory"), viper.GetString("save_file_names.combination")),
	)
	if err != nil {
		log.Printf("Unable to start save data service (live scores) : %s", err)
	}

	// Winning Outcome

	winningOutcome, err := saveFile.NewSaveFileService(
		saveFile.WithMysqlWoFilesRepository(viper.GetString("mySQL.live")),
		saveFile.WithDirectoryReaderRepository(viper.GetString("ke.wo_directory"), viper.GetString("save_file_names.combination")),
	)
	if err != nil {
		log.Printf("Unable to start save file data service (winning outcomes) :: %s", err)
	}

	// Odds

	odds, err := saveFile.NewSaveFileService(
		saveFile.WithMysqlOddsFilesRepository(viper.GetString("mySQL.live")),
		saveFile.WithDirectoryReaderRepository(viper.GetString("ke.odds_directory"), viper.GetString("save_file_names.combination")),
	)
	if err != nil {
		log.Printf("Unable to start file data service (Odds) : %s", err)
	}

	ctx := context.Background()

	log.Printf("** About to start APP ***")

	// 1. Kenya

	go func() {
		err := liveScoreService.SaveLiveScoreFiles(ctx, viper.GetString("ke.ls_directory"), viper.GetString("ke.country"))
		if err != nil {
			log.Printf("Err : %v unable to save ls files", err)
		}
	}()

	go func() {
		err := winningOutcome.SaveWinningOutcomesFiles(ctx, viper.GetString("ke.wo_directory"), viper.GetString("ke.country"))
		if err != nil {
			log.Printf("Err : %v unable to save wo files", err)
		}
	}()

	go func() {
		err := odds.SaveFiles(ctx, viper.GetString("ke.odds_directory"), viper.GetString("ke.country"))
		if err != nil {
			log.Printf("Err : %v unable to save odds files...", err)
		}
	}()

	// 2. Tanzania

	// 3. Ghana

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("save_file_names.logs"), viper.GetInt("log_setting.MaxSize"),
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
