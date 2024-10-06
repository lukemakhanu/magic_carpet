package instantGameServer

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests/matchRequestsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/players"
	"github.com/lukemakhanu/magic_carpet/internal/domains/players/playersMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches/selectedMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp/sharedHttpConf"
)

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

// GetCORS : return cors
func (s *InstantGameServerService) GetCORS() gin.HandlerFunc {
	return s.httpConf.CORSMiddleware()
}

func (s *InstantGameServerService) QueryInstantGames(c *gin.Context) {
	var p players.PlayerRequests
	err := c.Bind(&p)
	if err != nil {
		log.Printf("err : %v unable to create requests for matches", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create requests for matches"})
		return
	}

	player, err := s.playersMysql.PlayerExists(c.Request.Context(), p.ProfileTag)
	if err != nil {
		log.Printf("err : %v unable to return a player information", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to return a player information"})
		return
	}

	selectedPlayerID := []string{}

	if len(player) == 0 {

		// Start by saving the player in our system

		status := "active"
		pp, err := players.NewPlayers(p.ProfileTag, status)
		if err != nil {
			log.Printf("err : %v unable to initialize a player", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to initialize a player"})
			return
		}

		playerID, err := s.playersMysql.Save(c.Request.Context(), *pp)
		if err != nil {
			log.Printf("err : %v unable to create a player", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create a player"})
			return
		}

		// Save match requests

		plyID := fmt.Sprintf("%d", playerID)
		selectedPlayerID = append(selectedPlayerID, plyID)
	} else {

		pID := player[0].PlayerID
		selectedPlayerID = append(selectedPlayerID, pID)

	}

	if len(selectedPlayerID) != 1 {
		log.Printf("Error on capturing playerID")
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to fetch player Identity"})
		return
	}

	plyID := selectedPlayerID[0]

	cTime := time.Now()

	st := 10
	et := 45
	var sTime = cTime.Add(time.Second * time.Duration(st))
	var eTime = cTime.Add(time.Second * time.Duration(et))

	var startTime = sTime.Format("2006-01-02 15:04:05")
	var endTime = eTime.Format("2006-01-02 15:04:05")

	earlyFinish := "no"
	played := "no"

	// We have to have matches for all
	// competition

	competitions := []string{"1", "2", "3", "4"}
	for _, cID := range competitions {

		// create five next games
		// that a client can stroll through

		matchRound := 5

		for x := 0; x < matchRound; x++ {

			keyCreated := "pending"
			pp, err := matchRequests.NewMatchRequests(cID, plyID, startTime, endTime, earlyFinish, played, keyCreated)
			if err != nil {
				log.Printf("err : %v unable to initialize a matchRequests", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to initialize a matchRequests"})
				return
			}

			matchRequestID, err := s.matchRequestMysql.Save(c.Request.Context(), *pp)
			if err != nil {
				log.Printf("err : %v unable to create match request for competition id %s", err, cID)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create match request"})
				return
			}

			// Query for matches required for this client.

			allMatches := s.availableMatches

			// Get matches the client has had recently

			usedParentMatchIDs, err := s.matchRequestMysql.PlayerUsedMatchesDesc(c.Request.Context(), plyID)
			if err != nil {
				log.Printf("err : %v unable to return used parent match ids", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to return ids"})
				return
			}

			// remove any parent_match_id used by this client
			for _, v := range usedParentMatchIDs {
				log.Printf("remove %s for player %s", v.ParentMatchID, plyID)
				delete(allMatches, v.ParentMatchID)
			}

			for x := 0; x < 10; x++ {

				mrID := fmt.Sprintf("%d", matchRequestID)
				parentMatchID := "0"

				sm, err := selectedMatches.NewSelectedMatches(plyID, mrID, parentMatchID)
				if err != nil {
					log.Printf("err : %v unable to initialize a selectedMatch", err)
					s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to initialize selectedMatch"})
					return
				}

				smID, err := s.selectedMatchMysql.Save(c.Request.Context(), *sm)
				if err != nil {
					log.Printf("err : %v unable to create selected match for competition id %s", err, cID)
					s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create match request"})
					return
				}

				log.Printf("selected match_id %d", smID)

			}
		}

	}

	// Query the pending ids from the db table request_matches and
	// selected_matches and create the game for the client.

	keyCreated := "pending"
	reqMatches, err := s.matchRequestMysql.PendingRequestedMatchDesc(c.Request.Context(), plyID, p.CompetitionID, keyCreated)
	if err != nil {
		log.Printf("err : %v unable to return pending requested matches", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "err on loading matches"})
		return
	}

	for _, m := range reqMatches {
		log.Printf("matchRequestID : %s | keyCreated : %s", m.MatchRequestID, m.KeyCreated)

		selMatch, err := s.selectedMatchMysql.GetMatchesbyMatchRequestID(c.Request.Context(), m.MatchRequestID)
		if err != nil {
			log.Printf("err : %v unable to return selected matches", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "err on loading selected matches"})
			return
		}

		for _, sm := range selMatch {
			log.Printf("sm.SelectedMatchesID, sm.ParentMatchID", sm.SelectedMatchesID, sm.ParentMatchID)

		}
	}

	// 	s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "selected match already registered"})
	// return

	// profileID := 1
	// message := fmt.Sprintf("Client of profile_id %d created successfully", profileID)

	// var desc = clientProfiles.RegisterClientData{
	// 	ProfileID: profileID,
	// 	Messsage:  message,
	// }

	// var data = clientProfiles.RegisterClientResponse{
	// 	StatusCode:        http.StatusOK,
	// 	StatusDescription: "OK",
	// 	Data:              desc,
	// }
	// s.httpConf.JSON(c.Writer, http.StatusOK, data)

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
