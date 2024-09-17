// Package main runs the tavern and performs an Order
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/lukemakhanu/magic_carpet/cmd/apis/game_server/interfaces"
	"github.com/lukemakhanu/magic_carpet/internal/services/dataServerApi"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/apis/game_server/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/apis/game_server/"

func main() {
	InitConfig()

	w, err := dataServerApi.NewDataServerApiService(
		dataServerApi.WithMysqlLeaguesRepository(viper.GetString("mysql.live")),
		dataServerApi.WithMysqlSeasonWeeksRepository(viper.GetString("mysql.live")),
		dataServerApi.WithRedisProdRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		fmt.Printf("Unable to start data server api service ** %v", err)
	}

	wls := strings.Split(viper.GetString("game_server.whitelisted"), ",")

	IPWhitelist := make(map[string]bool)
	for _, v := range wls {
		IPWhitelist[v] = true
	}

	log.Println("Whitelisted ips are >>>", IPWhitelist)

	interfaces.Run(viper.GetInt("game_server.port"), w, IPWhitelist)

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("game_server.logs"), viper.GetInt("log_setting.MaxSize"),
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
