// Package main runs the tavern and performs an Order
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lukemakhanu/magic_carpet/internal/services/instantGameServer"
	instantRedisServer "github.com/lukemakhanu/magic_carpet/internal/services/instantRedis"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Router *gin.Engine

var addConfigPathLive = "/apps/go/magic_carpet/cmd/instant/api/game_server/"
var addConfigPathLocal = "/apps/go/magic_carpet/cmd/instant/api/game_server/"

var key = "my32digitkey12345678901234567890"
var iv = "my16digitIvKey12"
var authURL = "http://34.89.14.139:8050/integration/auth"
var infoURL = "http://34.89.14.139:8050/integration/info"
var betURL = "http://34.89.14.139:8050/integration/bet"
var resultURL = "http://34.89.14.139:8050/integration/result"
var selectedTimeZone = "Africa/Nairobi"

func main() {
	InitConfig()

	redisLive := "127.0.0.1:6379"
	mysqlLive := "root:tribute@tcp(127.0.0.1)/magic_carpet?charset=utf8"

	ir, err := instantRedisServer.NewInstantRedisServerService(
		instantRedisServer.WithSharedHttpConfRepository(),
		instantRedisServer.WithRedisResultsRepository(redisLive, viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	allMatches := "SANITIZED_INSTANT_ODDS"
	availableMatches, err := ir.AvailableGames(ctx, allMatches)
	if err != nil {
		log.Printf("err : %v on loading all matches", err)
	}

	ms, err := instantGameServer.NewInstantGameServerService(
		instantGameServer.WithSharedHttpConfRepository(),
		instantGameServer.WithMysqlPlayersRepository(mysqlLive),
		instantGameServer.WithMysqlMatchesRequestsRepository(mysqlLive),
		instantGameServer.WithMysqlSelectedMatchesRepository(mysqlLive),
		instantGameServer.WithAvailableMatches(availableMatches),
		instantGameServer.WithRedisResultsRepository(redisLive, viper.GetInt("redis.dbNum"),
			viper.GetInt("redis.maxIdle"), viper.GetInt("redis.maxActive"), viper.GetDuration("redis.duration")),
	)
	if err != nil {
		panic(err)
	}

	// Start Api here
	Run(viper.GetInt("instant_game_server.port"), ms)

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	s := <-sig

	fmt.Println("caught signal and exiting:::", s)
}

func Run(port int, ms *instantGameServer.InstantGameServerService) {
	Router = gin.Default()

	Router.Use(cors.New(cors.Config{
		AllowMethods:     []string{http.MethodPost, http.MethodGet, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"Origin", "content-type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if origin == "http://localhost:5174" || origin == "http://localhost:5173" || origin == "https://veimu.site" || origin == "https://veimu.site/" ||
				origin == "https://unicraftstudio.online" || origin == "https://unicraftstudio.online/" || origin == "https://unicraftstudio.info" || origin == "https://unicraftstudio.info/" ||
				origin == "https://unicraftstudio.com" || origin == "https://unicraftstudio.com/" || origin == "https://unicraftstudio.live" || origin == "https://unicraftstudio.live/" {
				return true
			} else {
				return false
			}
		},
		MaxAge: 12 * time.Hour,
	}))

	instantGames := Router.Group("/v1/")
	{
		instantGames.POST("/fetch_instant_games", ms.FetchInstantGame)
	}

	portStr := fmt.Sprintf(":%d", port)
	log.Printf("Running on port ::: %s", portStr)
	Router.Run(portStr)
}

func InitConfig() {
	configUtils(addConfigPathLive, addConfigPathLocal)
	logUtils(viper.GetString("instant_game_server.logs"), viper.GetInt("log_setting.MaxSize"),
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
		log.Printf("Error :: %v", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed:: %s", e.Name)
	})
}
