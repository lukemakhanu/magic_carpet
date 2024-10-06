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
	"github.com/lukemakhanu/magic_carpet/internal/services/v1Team"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/file_processors/save_team_in_redis/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/file_processors/save_team_in_redis/"

func main() {
	InitConfig()

	teamsList := "TEAMS_H2H"

	pg, err := v1Team.NewV1ProcessFileService(
		v1Team.WithDirectoryReaderRepository(viper.GetString("teams.directory"), viper.GetString("teams.combination")),
		v1Team.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf("Unable to start v1 team service : %s", err)
	}

	ctx := context.Background()

	go func() {
		err := pg.SaveL1Teams(ctx, viper.GetString("teams.directory"), teamsList)
		if err != nil {
			log.Printf("Err : %v unable to save teams", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:", s)
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("teams.logs"), viper.GetInt("log_setting.MaxSize"),
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
