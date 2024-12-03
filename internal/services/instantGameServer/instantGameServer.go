package instantGameServer

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns/goalPatternsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests/matchRequestsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs/mrsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/periods"
	"github.com/lukemakhanu/magic_carpet/internal/domains/periods/periodsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/playerUsedMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/playerUsedMatches/playerUsedMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/players"
	"github.com/lukemakhanu/magic_carpet/internal/domains/players/playersMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches/selectedMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp/sharedHttpConf"
)

type MatchDetails struct {
	Home     int
	Away     int
	Total    int
	Category string
}

// InstantGameServerConfiguration is an alias for a function that will take in a pointer to an InstantGameServerService and modify it
type InstantGameServerConfiguration func(os *InstantGameServerService) error

// InstantGameServerService is a implementation of the InstantGameServerService
type InstantGameServerService struct {
	redisConn          processRedis.RunRedis
	httpConf           sharedHttp.SharedHttpConfRepository
	playersMysql       players.PlayersRepository
	matchRequestMysql  matchRequests.MatchRequestsRepository
	selectedMatchMysql selectedMatches.SelectedMatchesRepository
	availableMatches   map[string]int

	goalPatternsMysql       goalPatterns.GoalPatternsRepository
	periodMysql             periods.PeriodsRepository
	mrsMysql                mrs.MrsRepository
	playersUsedMatchesMysql playerUsedMatches.PlayerUsedMatchesRepository
}

func NewInstantGameServerService(cfgs ...InstantGameServerConfiguration) (*InstantGameServerService, error) {
	// Create the NewClientAPIService
	os := &InstantGameServerService{}
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

func WithMysqlPlayersRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := playersMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.playersMysql = d
		return nil
	}
}

func WithMysqlMatchesRequestsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := matchRequestsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.matchRequestMysql = d
		return nil
	}
}

func WithMysqlSelectedMatchesRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := selectedMatchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.selectedMatchMysql = d
		return nil
	}
}

func WithRedisResultsRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// WithAvailableMatches : all available matches
func WithAvailableMatches(availableMatches map[string]int) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		os.availableMatches = availableMatches

		return nil
	}
}

// WithSharedHttpConfRepository : shared functions
func WithSharedHttpConfRepository() InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		cr, err := sharedHttpConf.New()
		if err != nil {
			return err
		}
		os.httpConf = cr
		return nil
	}
}

func WithMysqlGoalPatternsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := goalPatternsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.goalPatternsMysql = d
		return nil
	}
}

func WithMysqlPeriodsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := periodsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.periodMysql = d
		return nil
	}
}

// WithMysqlMrsRepository : returns match results
func WithMysqlMrsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := mrsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.mrsMysql = d
		return nil
	}
}

// WithMysqlPlayerUsedMatchesRepository : returns match results
func WithMysqlPlayerUsedMatchesRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := playerUsedMatchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.playersUsedMatchesMysql = d
		return nil
	}
}

// GetCORS : return cors
func (s *InstantGameServerService) GetCORS() gin.HandlerFunc {
	return s.httpConf.CORSMiddleware()
}

func (s *InstantGameServerService) FetchInstantGame(c *gin.Context) {
	var p players.PlayerRequests
	err := c.Bind(&p)
	if err != nil {
		log.Printf("err : %v unable to create requests for matches", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create requests for matches"})
		return
	}

	if len(p.ProfileTag) == 0 {
		log.Printf("Missing profile Tag")
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	player, err := s.playersMysql.PlayerExists(c.Request.Context(), p.ProfileTag)
	if err != nil {
		log.Printf("err : %v unable to return a player information", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to return a player information"})
		return
	}

	log.Println("player >>>> ", player)

	s.httpConf.JSON(c.Writer, http.StatusOK, gin.H{"error": "Welcome to instant games API"})
	return

}

/*
func (s *InstantGameServerService) FetchInstantGames(c *gin.Context) {
	oddsSortedSet := "SANITIZED_ODDS"
	var p players.PlayerRequests
	err := c.Bind(&p)
	if err != nil {
		log.Printf("err : %v unable to create requests for matches", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create requests for matches"})
		return
	}

	if len(p.ProfileTag) == 0 {
		log.Printf("Missing profile Tag")
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	player, err := s.playersMysql.PlayerExists(c.Request.Context(), p.ProfileTag)
	if err != nil {
		log.Printf("err : %v unable to return a player information", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to return a player information"})
		return
	}

	log.Println("player >>>> ", player)

	if len(player) == 0 {
		log.Printf("User %s does not exist", p.ProfileTag)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	allCompetitions := []string{"1", "2", "3", "4"}
	for _, competitionID := range allCompetitions {

		// 1. Create Match request
		playerID := player[0].PlayerID
		pl, err := matchRequests.NewMatchRequests(playerID)
		if err != nil {
			log.Printf("err %v unable to instantiate match request", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		matchRequestID, err := s.matchRequestMysql.Save(c.Request.Context(), *pl)
		if err != nil {
			log.Printf("err %v unable to save match request", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		// 2. Create periods for this match request

		matchReqID := fmt.Sprintf("%d", matchRequestID)

		goalDistribution, err := s.goalPatternsMysql.GoalDistributions(c.Request.Context(), competitionID)
		if err != nil {
			log.Printf("err : %v failed to return goal distribution", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		if len(goalDistribution) == 0 {
			log.Printf("err : %v no goal distribution returned %d", err, len(goalDistribution))
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		availableList := len(goalDistribution)
		selCategory := rand.IntN(availableList)

		selectedBatch := goalDistribution[selCategory]

		log.Println("selCategory", selCategory, "selectedBatch >>>> ", selectedBatch)

		parentIDs := strings.Split(selectedBatch.RoundNumberID, ",")

		if len(parentIDs) != 38 {
			log.Printf("err : %v there round number ids returned arent enough %d", err, len(parentIDs))
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		distr := make(map[int]string)
		kk := 1
		for _, f := range parentIDs {
			distr[kk] = f
			kk++
		}

		// Create a list of 38 records in periods table
		for _, rnID := range distr {

			current_time := time.Now().Local()

			startTime := current_time.Format("2006-01-02")
			endTime := current_time.Format("2006-01-02")
			earlyFinish := "no"
			played := "no"
			gameStarted := "no"
			keyCreated := "no"

			pp, err := periods.NewPeriods(competitionID, matchReqID, startTime, endTime, earlyFinish, played, gameStarted, keyCreated, rnID)
			if err != nil {
				log.Printf("err : %v failed to instantiate periods", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
				return
			}

			periodID, err := s.periodMysql.Save(c.Request.Context(), *pp)
			if err != nil {
				log.Printf("err : %v failed to save periods", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
				return
			}

			log.Printf("Save PeriodID >>> %d >>>", periodID)

			// Create selected Matched here. // We save 9 or 10 games depending on competition.

			distr, err := s.mrsMysql.GoalDistribution(c.Request.Context(), rnID, competitionID)
			if err != nil {
				log.Printf("err : %v failed to query goal distribution ", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
				return
			}

			matchMap, err := s.Validate(c, playerID, oddsSortedSet, distr, competitionID)
			if err != nil {
				log.Printf("err : %v failed to Validate list of matches ", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "bad request"})
				return
			}

		}

	}

	s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"success": "we can start working"})
	return
}*/

func (s *InstantGameServerService) FetchRoundMatches(ctx context.Context, playerID, oddsSortedSet string, distr []mrs.Mrs) ([]string, error) {

	list := []string{}

	m, err := s.GetGoalPattern(oddsSortedSet, distr)
	if err != nil {
		return list, fmt.Errorf("err : %v failed to return GetGoalPattern", err)
	}

	for _, dd := range m {

		sortedSetName := fmt.Sprintf("%s_%s_%s", "CL", playerID, dd.Category)

		data, err := s.redisConn.GetZRangeWithLimit(ctx, sortedSetName, 15)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to read from %s z range", err, sortedSetName)
		}

		categoryCount := len(data)

		if categoryCount > 0 {

			log.Println("*** sel category *** ", sortedSetName, " data returned ", data)
			if len(data) > 0 {

				// shuffle data

				categoryIndex := rand.IntN(categoryCount)
				selectedMatchID := data[categoryIndex]

				// Query this record from db so we know which record save as used

				ss, err := s.playersUsedMatchesMysql.GetMatchDetails(ctx, dd.Category, selectedMatchID)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to return match details of key %s category %s",
						err, selectedMatchID, dd.Category)
				}

				for _, r := range ss {

					log.Printf("category : %s | GoalID %s | MatchID %s", r.Category, r.GoalID, r.MatchID)

					// Add value into used table

					aa, err := playerUsedMatches.NewPlayerUsedMatches(playerID, r.Country, r.ProjectID, r.MatchID, r.Category)
					if err != nil {
						return list, fmt.Errorf("err : %v failed to instantiate UsedMatches", err)
					}

					inserted, err := s.playersUsedMatchesMysql.Save(ctx, *aa)
					if err != nil {
						return list, fmt.Errorf("err : %v failed to save playerUsedMatch", err)
					} else {
						log.Printf("Last inserted playeUsedMatchID %d", inserted)

						list = append(list, selectedMatchID)
						s.RemoveUsedKeys(ctx, sortedSetName, selectedMatchID)
					}

				}

			}
		} else {

			log.Printf("Key ::::: %s does not have enough data ", sortedSetName)

			return list, fmt.Errorf("err : %v does not have enough data %s", err, sortedSetName)
		}

	}

	return list, nil

}

func (s *InstantGameServerService) GetGoalPattern(oddsSortedSet string, distr []mrs.Mrs) (map[int]MatchDetails, error) {

	m := make(map[int]MatchDetails)
	x := 1

	for _, d := range distr {

		log.Printf(">>>>> Raw string >>>>> %s", d.RawScores)

		scores := strings.Split(d.RawScores, "**")
		// raw_scores: 52824375#1#2**52824376#1#1**52824377#2#1**52824378#2#2**52824379#1#0**52824380#0#1**52824381#2#0**52824382#1#1**52824383#3#1**52824384#1#1

		for _, cc := range scores {

			goals := strings.Split(cc, "#")

			if len(goals) == 3 {

				hSc := goals[1]
				aSc := goals[2]

				log.Printf("**** HomeScore %s | Away Score %s ****", hSc, aSc)

				homeGoals, err := strconv.Atoi(hSc)
				if err != nil {
					return m, fmt.Errorf("err : %v failed to convert hsc to int", err)
				}

				awayGoals, err := strconv.Atoi(aSc)
				if err != nil {
					return m, fmt.Errorf("err : %v failed to convert asc to int", err)
				}

				totalGoals := homeGoals + awayGoals

				if totalGoals == 0 {

					sel0 := fmt.Sprintf("%s_%d", oddsSortedSet, 0)
					m[x] = MatchDetails{
						Home:     homeGoals,
						Away:     awayGoals,
						Total:    totalGoals,
						Category: sel0,
					}

				} else if totalGoals == 1 {

					if homeGoals == 1 {

						sel1h := fmt.Sprintf("%s_%s_%s", oddsSortedSet, "1", "h")
						m[x] = MatchDetails{
							Home:     homeGoals,
							Away:     awayGoals,
							Total:    totalGoals,
							Category: sel1h,
						}

					} else {

						sel1a := fmt.Sprintf("%s_%s_%s", oddsSortedSet, "1", "a")
						m[x] = MatchDetails{
							Home:     homeGoals,
							Away:     awayGoals,
							Total:    totalGoals,
							Category: sel1a,
						}

					}

				} else {

					if homeGoals > 0 && awayGoals > 0 {

						// Query from Goal Goal batch

						if homeGoals > awayGoals {

							selggh := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "gg", "h")
							m[x] = MatchDetails{
								Home:     homeGoals,
								Away:     awayGoals,
								Total:    totalGoals,
								Category: selggh,
							}

						} else if awayGoals > homeGoals {

							selgga := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "gg", "a")
							m[x] = MatchDetails{
								Home:     homeGoals,
								Away:     awayGoals,
								Total:    totalGoals,
								Category: selgga,
							}

						} else {

							selggd := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "gg", "d")
							m[x] = MatchDetails{
								Home:     homeGoals,
								Away:     awayGoals,
								Total:    totalGoals,
								Category: selggd,
							}

						}

					} else {

						// Query from No goal

						if homeGoals > awayGoals {

							selngh := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "ng", "h")
							m[x] = MatchDetails{
								Home:     homeGoals,
								Away:     awayGoals,
								Total:    totalGoals,
								Category: selngh,
							}

						} else if awayGoals > homeGoals {

							selnga := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "ng", "a")
							m[x] = MatchDetails{
								Home:     homeGoals,
								Away:     awayGoals,
								Total:    totalGoals,
								Category: selnga,
							}

						} else {

							selngd := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "ng", "d")
							m[x] = MatchDetails{
								Home:     homeGoals,
								Away:     awayGoals,
								Total:    totalGoals,
								Category: selngd,
							}

						}
					}
				}
			}

			x++
		}
	}

	return m, nil
}

func (s *InstantGameServerService) Validate(ctx context.Context, playerID, oddsSortedSet string, distr []mrs.Mrs, competitionID string) (map[int]oddsFiles.CheckKeys, error) {

	m := make(map[int]oddsFiles.CheckKeys)

	selF := []int{}
	if competitionID == "1" || competitionID == "2" || competitionID == "4" {
		selF = append(selF, 10)
	} else {
		selF = append(selF, 9)
	}

	fetched := selF[0]
	log.Printf(" fetched +++++>>>>>> %d", fetched)

	// Get the games ration for over TG25

	keysList := []oddsFiles.CheckKeys{}
	//data, err := s.DecideRatio(ctx, oddsSortedSet, fetched, distr, competitionID)
	data, err := s.FetchRoundMatches(ctx, playerID, oddsSortedSet, distr)

	if err != nil {
		return m, fmt.Errorf("err : %v failed to read from %s z range", err, oddsSortedSet)
	}

	if len(data) < 10 && fetched == 10 {
		return m, fmt.Errorf("*** There are no enough matches ready to create a seen week *** count **** %d", len(data))
	}

	if len(data) < 9 && fetched == 9 {
		return m, fmt.Errorf("*** There are no enough matches ready to create a seen week *** count **** %d", len(data))
	}

	num := 0

	for _, o := range data {

		log.Printf(">>>>>> data fetched >>>>> %s", o)

		if num < fetched {

			yy := oddsFiles.CheckKeys{}
			yy.OddsKey = o

			parentID := strings.Split(o, "IO:") // example tzO:31475633 or keO:31475634
			if len(parentID) == 2 {

				// 1 . Find Odds

				log.Printf("oddKey --> %s", o)
				oddsData, err := s.redisConn.Get(ctx, o)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match odds from redis ", err)
				}

				// 2. Find Winning outcomes

				woKey := fmt.Sprintf("%s%s%s", parentID[0], "IWo:", parentID[1])
				log.Printf("woKey --> %s", woKey)
				woData, err := s.redisConn.Get(ctx, woKey)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match winning outcomes from redis ", err)
				}

				// 3. Find Live scores

				lsKey := fmt.Sprintf("%s%s%s", parentID[0], "ILs:", parentID[1])
				log.Printf("lsKey --> %s", lsKey)
				lsData, err := s.redisConn.Get(ctx, lsKey)
				if err != nil {
					log.Printf("Err : %v failed to get live score from redis ", err)
				}

				if len(oddsData) > 0 && len(woData) > 0 {
					log.Printf("This data is okay to be used")

					ss := oddsFiles.ValidateKeys{
						Odds: oddsData,
						Wo:   woData,
						Ls:   lsData,
					}

					yy.ValidateKeys = ss

					keysList = append(keysList, yy)

					num++

				} else {
					// Delete this record and retry again
					return m, fmt.Errorf("Skip this record oddKey : %s wo %s ls %s", o, woKey, lsKey)
				}

			} else {
				return m, fmt.Errorf("Data saved in bad format : %s", parentID)
			}

		} else {

			//return m, fmt.Errorf("skip %s from adding this game as the season week is full ", o)
			log.Printf("skip %s from adding this game as the season week is full ", o)
		}

	}

	log.Printf("num is ---> %d | fetched ---> %d", num, fetched)

	if num == fetched {

		d := 0
		for _, v := range keysList {
			m[d] = v
			d++
		}

		return m, nil

	}

	return m, fmt.Errorf("final match count : %d not enough to create season week", len(data))
}

// AvailableGames : returns all available games
func (s *InstantGameServerService) AvailableGames(c *gin.Context, allMatches string) ([]string, error) {
	data, err := s.redisConn.GetZRange(c.Request.Context(), allMatches)
	if err != nil {
		log.Printf("Err : %v", err)
		return data, err
	}
	log.Println("id : ", data)
	return data, nil
}

func (s *InstantGameServerService) RemoveUsedKeys(ctx context.Context, key, value string) {
	_, err := s.redisConn.ZRem(ctx, key, value)
	if err != nil {
		log.Printf("Err : %v", err)
	} else {
		log.Printf("###### deleted key %s from zset %s ######", key, value)
	}
}

func (s *InstantGameServerService) TeamInfo(leagueID string, teamID string) (string, string) {

	switch leagueID {
	case "1":
		if teamID == "9" {
			return "EvertonF", "EVE"
		} else if teamID == "1" {
			return "ManchesterC", "MNC"
		} else if teamID == "14" {
			return "BrightonF", "BRT"
		} else if teamID == "6" {
			return "ArsenalG", "ARS"
		} else if teamID == "7" {
			return "BurnleyF", "BUR"
		} else if teamID == "8" {
			return "LeicesterF", "LEI"
		} else if teamID == "4" {
			return "ChelseaB", "CHE"
		} else if teamID == "5" {
			return "TottenhamH", "TOT"
		} else if teamID == "18" {
			return "SouthamptomF", "SOU"
		} else if teamID == "13" {
			return "Newcastle", "NEW"
		} else if teamID == "12" {
			return "WestHamF", "WHU"
		} else if teamID == "3" {
			return "LiverPoolF", "LIV"
		} else if teamID == "15" {
			return "CrystalP", "CRY"
		} else if teamID == "2" {
			return "ManchesterU", "MNU"
		} else if teamID == "19" {
			return "WolverhamptomF", "WOV"
		} else if teamID == "17" {
			return "AstonV", "ARV"
		} else if teamID == "20" {
			return "SheffieldF", "SHE"
		} else if teamID == "16" {
			return "FulHamF", "FUL"
		} else if teamID == "11" {
			return "WestBromF", "WR"
		} else if teamID == "10" {
			return "LeedsF", "LEE"
		} else {
			return "0", "0"
		}
	case "2":
		if teamID == "1" {
			return "BAR", "BAR"
		} else if teamID == "2" {
			return "ATM", "ATM"
		} else if teamID == "3" {
			return "BET", "BET"
		} else if teamID == "4" {
			return "RMA", "RMA"
		} else if teamID == "5" {
			return "GET", "GET"
		} else if teamID == "6" {
			return "EIB", "EIB"
		} else if teamID == "7" {
			return "VAL", "VAL"
		} else if teamID == "8" {
			return "VIL", "VIL"
		} else if teamID == "9" {
			return "SEV", "SEV"
		} else if teamID == "10" {
			return "ESP", "ESP"
		} else if teamID == "11" {
			return "MAL", "MAL"
		} else if teamID == "12" {
			return "CEL", "CEL"
		} else if teamID == "13" {
			return "ALA", "ALA"
		} else if teamID == "14" {
			return "ATH", "ATH"
		} else if teamID == "15" {
			return "LEV", "LEV"
		} else if teamID == "16" {
			return "RSO", "RSO"
		} else if teamID == "17" {
			return "GRA", "GRA"
		} else if teamID == "18" {
			return "OSA", "OSA"
		} else if teamID == "19" {
			return "VLL", "VLL"
		} else if teamID == "20" {
			return "LEG", "LEG"
		} else {
			return "0", "0"
		}
	case "3":
		if teamID == "1" {
			return "GOR", "GOR"
		} else if teamID == "2" {
			return "SOF", "SOF"
		} else if teamID == "3" {
			return "BAN", "BAN"
		} else if teamID == "4" {
			return "KKG", "KKG"
		} else if teamID == "5" {
			return "MAT", "MAT"
		} else if teamID == "6" {
			return "TUS", "TUS"
		} else if teamID == "7" {
			return "SON", "SON"
		} else if teamID == "8" {
			return "ULS", "ULS"
		} else if teamID == "9" {
			return "AFC", "AFC"
		} else if teamID == "10" {
			return "KRS", "KRS"
		} else if teamID == "11" {
			return "KCB", "KCB"
		} else if teamID == "12" {
			return "NZS", "NZS"
		} else if teamID == "13" {
			return "WST", "WST"
		} else if teamID == "14" {
			return "PST", "PST"
		} else if teamID == "15" {
			return "CHM", "CHM"
		} else if teamID == "16" {
			return "VHG", "VHG"
		} else if teamID == "17" {
			return "ZOO", "ZOO"
		} else if teamID == "18" {
			return "NKT", "NKT"
		} else {
			return "0", "0"
		}
	case "4":
		if teamID == "1" {
			return "JUV", "JUV"
		} else if teamID == "2" {
			return "NAP", "NAP"
		} else if teamID == "3" {
			return "INT", "INT"
		} else if teamID == "4" {
			return "MIL", "MIL"
		} else if teamID == "5" {
			return "ATA", "ATA"
		} else if teamID == "6" {
			return "ROM", "ROM"
		} else if teamID == "7" {
			return "LAZ", "LAZ"
		} else if teamID == "8" {
			return "TOR", "TOR"
		} else if teamID == "9" {
			return "SAM", "SAM"
		} else if teamID == "10" {
			return "FIO", "FIO"
		} else if teamID == "11" {
			return "SAS", "SAS"
		} else if teamID == "12" {
			return "CAG", "CAG"
		} else if teamID == "13" {
			return "GEN", "GEN"
		} else if teamID == "14" {
			return "PAR", "PAR"
		} else if teamID == "15" {
			return "UDI", "UDI"
		} else if teamID == "16" {
			return "SPA", "SPA"
		} else if teamID == "17" {
			return "BOL", "BOL"
		} else if teamID == "18" {
			return "BRE", "BRE"
		} else if teamID == "19" {
			return "LEC", "LEC"
		} else if teamID == "20" {
			return "VER", "VER"
		} else {
			return "0", "0"
		}
	default:
		return "0", "0"
	}
}
