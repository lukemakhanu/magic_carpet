package interfaces

import (
	"fmt"
	"log"
	"time"

	//"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	"github.com/lukemakhanu/magic_carpet/cmd/apis/game_server/interfaces/middleware"
	"github.com/lukemakhanu/magic_carpet/internal/services/dataServerApi"
)

var Router *gin.Engine

func GenerateCode() string {
	return fmt.Sprint(time.Now().Nanosecond())[:6]
}

func Run(port int, w *dataServerApi.DataServerApiService, IPWhitelist map[string]bool) {
	Router = gin.Default()

	v1 := Router.Group("/v1")
	v1.Use(middleware.CORSMiddleware())
	//v1.Use(middleware.IPWhiteListMiddleware(IPWhitelist))
	{
		// PRODUCTION ENDPOINTS
		v1.GET("/production_matches", w.GetProdMatches)
		v1.GET("/production_results", w.GetProdWinningOutcomes)
		v1.GET("/production_live_scores", w.GetProdLiveScores)
	}

	v2 := Router.Group("/v2")
	v2.Use(middleware.IPWhiteListMiddleware(IPWhitelist))
	v2.Use(middleware.CORSMiddleware())
	{

		// WHITELISTED END POINTS
		v2.GET("/games", w.GetProdMatches)
		v2.GET("/results", w.GetProdWinningOutcomes)
		v2.GET("/scores", w.GetProdLiveScores)
	}

	portStr := fmt.Sprintf(":%d", port)
	log.Printf("Running on port ::: %s", portStr)
	Router.Run(portStr)
}
