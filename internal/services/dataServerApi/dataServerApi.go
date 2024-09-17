package dataServerApi

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/leagues"
	"github.com/lukemakhanu/magic_carpet/internal/domains/leagues/leaguesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/seasonWeeks"
	seasonWeekMysql "github.com/lukemakhanu/magic_carpet/internal/domains/seasonWeeks/seasonWeeksMysql"
)

// DataServerApiConfiguration is an alias for a function that will take in a pointer to an DataServerApiService and modify it
type DataServerApiConfiguration func(os *DataServerApiService) error

// DataServerApiService is a implementation of the DataServerApiService
type DataServerApiService struct {
	leaguesMysql    leagues.LeaguesRepository
	seasonWeekMysql seasonWeeks.SeasonWeeksRepository
	redisProdConn   processRedis.RunRedis
}

// NewDataServerApiService : instantiate dataServerApi
func NewDataServerApiService(cfgs ...DataServerApiConfiguration) (*DataServerApiService, error) {
	// Create the DataServerApiService
	os := &DataServerApiService{}
	// Apply all Configurations passed in
	for _, cfg := range cfgs {
		// Pass the service into the configuration function
		err := cfg(os)
		if err != nil {
			return nil, err
		}
	}
	return os, nil
}

// WithMysqlLeaguesRepository :
func WithMysqlLeaguesRepository(connectionString string) DataServerApiConfiguration {
	return func(os *DataServerApiService) error {
		d, err := leaguesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.leaguesMysql = d
		return nil
	}
}

// WithMysqlSeasonWeeksRepository :
func WithMysqlSeasonWeeksRepository(connectionString string) DataServerApiConfiguration {
	return func(os *DataServerApiService) error {
		d, err := seasonWeekMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.seasonWeekMysql = d
		return nil
	}
}

func WithRedisProdRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) DataServerApiConfiguration {
	return func(os *DataServerApiService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisProdConn = d
		return nil
	}
}

// GetProdMatches : used to return matches
func (s *DataServerApiService) GetProdMatches(c *gin.Context) {
	season_week_id := c.DefaultQuery("season_week_id", "0")

	log.Println("season_week_id", season_week_id)

	selSeasonWeekID := []string{}

	if season_week_id == "0" || season_week_id == "" {
		selSeasonWeekID = append(selSeasonWeekID, "EnglishLeague")
	} else {
		selSeasonWeekID = append(selSeasonWeekID, season_week_id)
	}

	var vl oddsFiles.MatchAPI

	log.Printf("selected season week id %s", selSeasonWeekID[0])

	data, err := s.leaguesMysql.GetProductionMatchDetailsNew(c, selSeasonWeekID[0])
	if err != nil {

		log.Printf("Err : %v", err)

		vl.StatusCode = "200"
		vl.StatusDescription = "Matches not found"
		_, err = json.Marshal(vl)
		if err != nil {
			log.Printf("Err : %v", err)
		}

		c.JSON(500, vl)
		return
	}

	for _, m := range data {

		log.Printf("m.ClientID:%s, m.League:%s, m.LeagueAbbrv:%s, m.LeagueID:%s, m.MatchDate:%s, m.SeasonID:%s",
			m.ClientID, m.League, m.LeagueAbbrv, m.LeagueID, m.MatchDate, m.SeasonID)

		md := oddsFiles.MatchDetails{
			MatchDate:   m.MatchDate,
			League:      m.LeagueAbbrv,
			MatchSeason: m.SeasonID,
		}

		// Add match day here

		ssn, err := s.seasonWeekMysql.ApiSsnWeeksNew(c, m.SeasonID)
		if err != nil {
			log.Printf("Err : %v failed to get match day", err)
		}

		dd := []oddsFiles.FinalSeasonWeek{}
		for _, y := range ssn {
			log.Printf("season week id %s y.SeasonID :%s, y.ApiDate :%s, y.LeagueID :%s", y.SeasonWeekID, y.SeasonID, y.ApiDate, y.LeagueID)

			keyName := fmt.Sprintf("%s_%s_%s", "pr_odds", y.ApiDate, y.SeasonWeekID)
			log.Printf(">>> match key saved >>>> %s", keyName)

			matchDay, err := s.redisProdConn.Get(c, keyName)
			if err != nil {
				log.Printf("Err : %v failed to get match day from redis", err)
			}

			//log.Printf("match data : %s", matchDay)

			var msg oddsFiles.FinalSeasonWeek //FinalSeasonWeek{}

			err = json.Unmarshal([]byte(matchDay), &msg)
			if err != nil {
				log.Printf("Err unable to unmarshal match : %v", err)
			} else {
				// md.MatchDay = msg
				dd = append(dd, msg)
			}

		}

		md.MatchDay = dd

		vl.MatchDetails = md

		// End of match day here

		vl.StatusCode = "200"
		vl.StatusDescription = "success"
		_, err = json.Marshal(vl)
		if err != nil {
			log.Printf("Err : %v", err)
		}
		c.JSON(200, vl)

		return
	}

	vl.StatusCode = "200"
	vl.StatusDescription = "Matches not found"
	_, err = json.Marshal(vl)
	if err != nil {
		log.Printf("Err : %v", err)
	}
	c.JSON(500, vl)
	return

}

// GetProdWinningOutcomes : used to return matches
func (s *DataServerApiService) GetProdWinningOutcomes(c *gin.Context) {
	seasonWeekID := c.DefaultQuery("season_week_id", "0")

	log.Println("season_week_id", seasonWeekID)

	selSeasonWeekID := []string{}

	if seasonWeekID == "0" || seasonWeekID == "" {
		selSeasonWeekID = append(selSeasonWeekID, "0")
	} else {
		selSeasonWeekID = append(selSeasonWeekID, seasonWeekID)
	}

	var vl oddsFiles.WoAPI

	data, err := s.leaguesMysql.GetProductionWinningOutcomesNew(c, selSeasonWeekID[0])
	if err != nil {

		vl.StatusCode = "200"
		vl.StatusDescription = "Matches not found"
		_, err = json.Marshal(vl)
		if err != nil {
			log.Printf("Err : %v", err)
		}

		c.JSON(200, vl)
		return
	}

	for _, m := range data {

		sTime, err := time.Parse("2006-01-02 15:04:05", m.StartTime)
		if err != nil {
			log.Printf("Err : %v failed to convert string to time", err)
		}

		keyName := fmt.Sprintf("%s_%s_%s", "pr_wo", sTime.Format("2006-01-02"), m.SeasonWeekID)
		log.Printf(">>> winning outcome key saved >>>> %s", keyName)

		matchDayWo, err := s.redisProdConn.Get(c, keyName)
		if err != nil {
			log.Printf("Err : %v failed to get match day winning outcome from redis", err)
		}

		//log.Printf("result data found : %s", matchDayWo)

		var msg oddsFiles.FinalSeasonWeekWO
		err = json.Unmarshal([]byte(matchDayWo), &msg)
		if err != nil {
			log.Printf("Err unable to unmarshal winning outcome : %v", err)
		} else {
			vl.Results = msg
		}

		// End of match day here

		vl.StatusCode = "200"
		vl.StatusDescription = "success"
		_, err = json.Marshal(vl)
		if err != nil {
			log.Printf("Err : %v", err)
		}
		c.JSON(200, vl)

		return
	}

	vl.StatusCode = "200"
	vl.StatusDescription = "Result not found."
	_, err = json.Marshal(vl)
	if err != nil {
		log.Printf("Err : %v", err)
	}
	c.JSON(200, vl)
	return

}

// GetProdLiveScores : used to return live scores
func (s *DataServerApiService) GetProdLiveScores(c *gin.Context) {
	seasonWeekID := c.DefaultQuery("season_week_id", "0")

	log.Println("season_week_id", seasonWeekID)

	selSeasonWeekID := []string{}

	if seasonWeekID == "0" || seasonWeekID == "" {
		selSeasonWeekID = append(selSeasonWeekID, "0")
	} else {
		selSeasonWeekID = append(selSeasonWeekID, seasonWeekID)
	}

	var vl oddsFiles.LsAPI

	data, err := s.leaguesMysql.GetProductionWinningOutcomesNew(c, selSeasonWeekID[0])
	if err != nil {

		vl.StatusCode = "200"
		vl.StatusDescription = "Live scores not found"
		_, err = json.Marshal(vl)
		if err != nil {
			log.Printf("Err : %v", err)
		}

		c.JSON(200, vl)
		return
	}

	for _, m := range data {

		sTime, err := time.Parse("2006-01-02 15:04:05", m.StartTime)
		if err != nil {
			log.Printf("Err : %v failed to convert string to time", err)
		}

		keyName := fmt.Sprintf("%s_%s_%s", "pr_ls", sTime.Format("2006-01-02"), m.SeasonWeekID)
		log.Printf(">>> live score key >>>> %s", keyName)

		matchDayLS, err := s.redisProdConn.Get(c, keyName)
		if err != nil {
			log.Printf("Err : %v failed to get match day live score from redis", err)
		}

		//log.Printf("live score data found : %s", matchDayWo)

		var msg oddsFiles.FinalSeasonWeekLS
		err = json.Unmarshal([]byte(matchDayLS), &msg)
		if err != nil {
			log.Printf("Err unable to unmarshal live score : %v", err)
		} else {
			vl.LiveScore = msg
		}

		// End of match day here

		vl.StatusCode = "200"
		vl.StatusDescription = "success"
		_, err = json.Marshal(vl)
		if err != nil {
			log.Printf("Err : %v", err)
		}
		c.JSON(200, vl)

		return
	}

	vl.StatusCode = "200"
	vl.StatusDescription = "Live score not found"
	_, err = json.Marshal(vl)
	if err != nil {
		log.Printf("Err : %v", err)
	}
	c.JSON(200, vl)
	return

}
