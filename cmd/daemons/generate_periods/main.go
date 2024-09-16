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
	"github.com/lukemakhanu/magic_carpet/internal/services/generatePeriod"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var addConfigPathLive = "/apps/go/magic_carpet/cmd/generate_periods/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/generate_periods/"

var inProgress bool
var inProgress2 bool

func main() {
	InitConfig()

	pg, err := generatePeriod.NewGeneratePeriodService(
		generatePeriod.WithMysqlSsnsRepository(viper.GetString("mySQL.live")),
		generatePeriod.WithMysqlScheduledTimeRepository(viper.GetString("mySQL.live")),
		generatePeriod.WithRedisRepository(viper.GetString("redis.live"), viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		log.Printf("Unable to start generate period service ::: %s", err)
	}

	ctx := context.Background()

	locale, _ := time.LoadLocation(viper.GetString("generate_periods.time_zone"))

	status := "inactive"

	// Create Scheduled Times...

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if !inProgress {
					inProgress = true

					CreateScheduledTime(ctx, locale, pg, status)

				} else {
					log.Printf("**** CreateScheduledTime in process **** %v.\n", t)
				}
			}
		}
	}()

	ticker2 := time.NewTicker(5 * time.Second)
	defer ticker2.Stop()
	go func() {
		for {
			select {
			case t := <-ticker2.C:
				if !inProgress2 {
					inProgress2 = true

					PreparePeriods(ctx, locale, pg, status)

				} else {
					log.Printf("**** PreparePeriods in process %v.\n", t)
				}
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

// PreparePeriods :
func PreparePeriods(ctx context.Context, locale *time.Location, pg *generatePeriod.GeneratePeriodService, status string) {

	defer func() {
		inProgress2 = false
		log.Printf(" done PreparePeriods ... ")
	}()

	err := pg.PrepareGames(ctx, locale, "1", status)
	if err != nil {
		log.Printf("%v", err)
	}

	err = pg.PrepareGames(ctx, locale, "2", status)
	if err != nil {
		log.Printf("%v", err)
	}

	err = pg.PrepareGames(ctx, locale, "3", status)
	if err != nil {
		log.Printf("%v", err)
	}

	err = pg.PrepareGames(ctx, locale, "4", status)
	if err != nil {
		log.Printf("%v", err)
	}
}

// PopulateOddsSet : saves odds
func CreateScheduledTime(ctx context.Context, locale *time.Location, pg *generatePeriod.GeneratePeriodService, status string) {

	defer func() {
		inProgress = false
		log.Printf(" Done creating scheduled ")
	}()

	err := pg.CreateScheduledTime(ctx, locale, "1", status, viper.GetInt64("generate_periods.category_one"))
	if err != nil {
		log.Printf("%v", err)
	}

	err = pg.CreateScheduledTime(ctx, locale, "2", status, viper.GetInt64("generate_periods.category_two"))
	if err != nil {
		log.Printf("%v", err)
	}

	err = pg.CreateScheduledTime(ctx, locale, "3", status, viper.GetInt64("generate_periods.category_two"))
	if err != nil {
		log.Printf("%v", err)
	}

	err = pg.CreateScheduledTime(ctx, locale, "4", status, viper.GetInt64("generate_periods.category_one"))
	if err != nil {
		log.Printf("%v", err)
	}
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("generate_periods.logs"), viper.GetInt("log_setting.MaxSize"),
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
