// Package main runs the tavern and performs an Order
package main

import (
	"context"

	"github.com/lukemakhanu/magic_carpet/internal/services/consumeMatchResult"

	//"odi_league_v4/internal/services/consumeRawLiveScore"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/odi_league_v4/cmd/ke/receivers/totalGoals/"
var addConfigPathLocal = "/apps/go/odi_league_v4/cmd/ke/receivers/totalGoals/"

func main() {
	InitConfig()

	mrService, err := consumeMatchResult.NewConsumeMatchResultService(
		consumeMatchResult.WithMysqlMrsRepository(viper.GetString("mySQL.live")),
		consumeMatchResult.WithRabbitConsumeMatchResult(viper.GetString("mQ.conn"), viper.GetString("mrq.queueName"), viper.GetString("mrq.connName"), viper.GetString("mrq.consumerName")),
	)
	if err != nil {
		log.Printf("Unable to start matchResult service : %s", err)
	}

	data := make(chan consumeMatchResult.Data, 100)

	ctx := context.Background()

	go func() {
		mrService.ConsumeMatchResult(ctx, data)
	}()

	go func() {
		mrService.SaveMatchResult(ctx, data)
	}()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("totalGoals.logs"), viper.GetInt("log_setting.MaxSize"),
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
