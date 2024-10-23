package productionKey

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches/checkMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matches/matchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs/mrsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsConfigs/processOdds"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/seasonWeeks"
	seasonWeekMysql "github.com/lukemakhanu/magic_carpet/internal/domains/seasonWeeks/seasonWeeksMysql"
)

type Job struct {
	JobType  string
	WorkStr  string
	WorkStr1 string
	WorkStr2 string
	WorkStr3 string
}

// ProcessKeyConfiguration is an alias for a function that will take in a pointer to an ProcessKeyService and modify it
type ProcessKeyConfiguration func(os *ProcessKeyService) error

// ProcessKeyService is a implementation of the ProcessKeyService
type ProcessKeyService struct {
	seasonWeekMysql   seasonWeeks.SeasonWeeksRepository
	matchesMysql      matches.MatchesRepository
	mrsMysql          mrs.MrsRepository
	checkMatchesMysql checkMatches.CheckMatchesRepository
	redisConn         processRedis.RunRedis
}

// NewProcessKeyService : instantiate every connection we need to run current game service
func NewProcessKeyService(cfgs ...ProcessKeyConfiguration) (*ProcessKeyService, error) {
	// Create the seasonService
	os := &ProcessKeyService{}
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

// WithMysqlSeasonWeeksRepository : instantiates mysql to connect to season weeks interface
func WithMysqlSeasonWeeksRepository(connectionString string) ProcessKeyConfiguration {
	return func(os *ProcessKeyService) error {
		d, err := seasonWeekMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.seasonWeekMysql = d
		return nil
	}
}

// WithMysqlMatchesRepository : instantiates mysql to connect to matches interface
func WithMysqlMatchesRepository(connectionString string) ProcessKeyConfiguration {
	return func(os *ProcessKeyService) error {
		d, err := matchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.matchesMysql = d
		return nil
	}
}

// WithMysqlCheckMatchesRepository : instantiates mysql to connect to matches interface
func WithMysqlCheckMatchesRepository(connectionString string) ProcessKeyConfiguration {
	return func(os *ProcessKeyService) error {
		d, err := checkMatchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.checkMatchesMysql = d
		return nil
	}
}

// WithMysqlMrsRepository : returns match results
func WithMysqlMrsRepository(connectionString string) ProcessKeyConfiguration {
	return func(os *ProcessKeyService) error {
		d, err := mrsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.mrsMysql = d
		return nil
	}
}

// WithRedisRepository : instantiates redis connections
func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) ProcessKeyConfiguration {
	return func(os *ProcessKeyService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// ReturnParentMatchIDs : returns all parent match ids in batches
func (s *ProcessKeyService) ReturnZRangeData(ctx context.Context, zSetKey string, fetched int) ([]string, error) {
	data, err := s.redisConn.GetZRangeWithLimit(ctx, zSetKey, fetched)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (s *ProcessKeyService) RedisGet(ctx context.Context, key string) (string, error) {
	return s.redisConn.Get(ctx, key)
}

func (s *ProcessKeyService) RedisZRem(ctx context.Context, list string, key string) (interface{}, error) {
	return s.redisConn.ZRem(ctx, list, key)
}

// GetUpcomingSeasonWeeks : used to return upcoming season week.
func (s *ProcessKeyService) GetUpcomingSeasonWeeks(ctx context.Context, oddsSortedSet, woSortedSet, liveScoreSortedSet string, oddsFactor float64) error {

	data, err := s.seasonWeekMysql.UpcomingSsnWeeks(ctx)
	if err != nil {
		return fmt.Errorf("err : %v failed to query upcoming season weeks ", err)
	}

	for _, x := range data {

		log.Printf("x.LeagueID:%s, x.SeasonWeekID:%s, x.SeasonID:%s", x.LeagueID, x.SeasonWeekID, x.SeasonID)

		n := 0

		matchMap, err := s.Validate(ctx, x.LeagueID, oddsSortedSet, woSortedSet, liveScoreSortedSet)
		if err != nil {
			log.Printf("Err : %v unable to create season week >>>>> ", err)
		} else {

			log.Println("matchMap ======>>> ", len(matchMap), " || seasonWeekData  ===> ", len(data))

			//log.Printf("seasonWeekID: %s, StartTime: %s, WeekNumber:%s, Status:%s",
			//	x.SeasonWeekID, x.StartTime, x.WeekNumber, x.Status)

			// Get all matches for this week.

			hh := oddsFiles.FinalSeasonWeek{}
			wo := oddsFiles.FinalSeasonWeekWO{}
			lsc := oddsFiles.FinalSeasonWeekLS{}

			hh.SeasonWeeKID = x.SeasonWeekID
			hh.StartTime = x.StartTime
			hh.EndTime = x.EndTime
			hh.MatchDay = x.WeekNumber
			hh.SeasonID = x.SeasonID

			wo.SeasonWeeKID = x.SeasonWeekID
			wo.StartTime = x.StartTime
			wo.EndTime = x.EndTime
			wo.MatchDay = x.WeekNumber
			wo.SeasonID = x.SeasonID

			lsc.SeasonWeeKID = x.SeasonWeekID
			lsc.StartTime = x.StartTime
			lsc.EndTime = x.EndTime
			lsc.MatchDay = x.WeekNumber
			lsc.SeasonID = x.SeasonID

			matches, err := s.matchesMysql.GetSeasonWeekGames(ctx, x.SeasonWeekID, x.SeasonID)
			if err != nil {
				log.Printf("Err : %v failed to query matches", err)
			}

			log.Println("matches =====>>>>>>>>", matches)

			for _, g := range matches {

				log.Printf("matchID:%s, homeTeamID:%s, awayTeamID:%s, seasonWeekID:%s",
					g.MatchID, g.HomeTeamID, g.AwayTeamID, g.SeasonWeekID)

				// Start creating magic here
				//leagueID := "1"
				homeTeamName, homeTeamAlias := s.TeamInfo(x.LeagueID, g.HomeTeamID)
				awayTeamName, awayTeamAlias := s.TeamInfo(x.LeagueID, g.AwayTeamID)

				matches := oddsFiles.FinalMatches{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				woMatches := oddsFiles.FinalMatchesWO{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				lsMatches := oddsFiles.FinalMatchesLS{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				fd := matchMap[n]

				log.Println("odds ---> ", len(fd.ValidateKeys.Odds))
				mts, err := processOdds.New(fd.ValidateKeys.Odds, fd.ValidateKeys.Wo, fd.ValidateKeys.Ls, oddsFactor)
				if err != nil {
					log.Printf("Err : %v failed to initialize odds ", err)
				}

				mtk, winningOutcomes, liveScores, err := mts.FormulateOdds2(ctx) // FormulateOdds(ctx)
				if err != nil {
					log.Printf("Err : %v failed to formulate odds ", err)
				}

				matches.FinalMarkets = mtk
				woMatches.FinalScore = winningOutcomes
				lsMatches.FinalLiveScores = liveScores

				log.Printf("oddsKey :::--> %s ", fd.OddsKey)
				log.Println("liveScores ---> ", liveScores)

				n++

				hh.FinalMatches = append(hh.FinalMatches, matches)
				wo.FinalMatchesWO = append(wo.FinalMatchesWO, woMatches)
				lsc.FinalMatchesLS = append(lsc.FinalMatchesLS, lsMatches)

				// Start working on winning outcomes

				// Update this match as processed

				status := "active"
				updated, err := s.matchesMysql.UpdateGameStatus(ctx, status, g.MatchID)
				if err != nil {
					log.Printf("Err : %v failed to update processed season week", err)
				}

				log.Printf("updated record : %d", updated)
			}

			// Save Match odds into redis for further use

			sTime, err := time.Parse("2006-01-02 15:04:05", x.StartTime)
			if err != nil {
				log.Printf("Err : %v failed to convert string to time", err)
			}

			keyName := fmt.Sprintf("%s_%s_%s", "pr_odds", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> Odds key saved >>>> %s", keyName)

			oddsData, err := json.Marshal(hh)
			if err != nil {
				log.Printf("Err: %v failed to marshall odds json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyName, string(oddsData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to odds set", err)
				}
			}

			keyNameWO := fmt.Sprintf("%s_%s_%s", "pr_wo", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> WinningOutcome key saved >>>> %s", keyNameWO)

			winningOutcomesData, err := json.Marshal(wo)
			if err != nil {
				log.Printf("Err: %v failed to marshall wo json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyNameWO, string(winningOutcomesData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to winning outcome set", err)
				}
			}

			keyNameLS := fmt.Sprintf("%s_%s_%s", "pr_ls", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> LiveScore key saved >>>> %s", keyNameLS)

			liveScoresData, err := json.Marshal(lsc)
			if err != nil {
				log.Printf("Err: %v failed to marshall wo json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyNameLS, string(liveScoresData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to save live score set", err)
				}
			}

			daysListKeys := fmt.Sprintf("%s_%s", "pr_keys", sTime.Format("2006-01-02"))
			daysListValues := fmt.Sprintf("%s_%s_%s", "pr_keys", sTime.Format("2006-01-02"), x.SeasonWeekID)

			err = s.redisConn.ZAdd(ctx, daysListKeys, "1", daysListValues)
			if err != nil {
				log.Printf("Err: %v failed to save into todays list", err)
			}

			status := "active"
			updated, err := s.seasonWeekMysql.UpdateSsnWeekStatus(ctx, x.SeasonWeekID, x.SeasonID, status)
			if err != nil {
				log.Printf("Err : %v failed to update processed season week", err)
			}

			log.Printf("updated record : %d", updated)
		}
	}

	return nil
}

// GetUpcomingSeasonWeeks : used to return upcoming season week.
func (s *ProcessKeyService) GetUpcomingSeasonWeeks2(ctx context.Context, oddsSortedSet, oddsu15Set string) error {

	data, err := s.seasonWeekMysql.UpcomingSsnWeeks(ctx)
	if err != nil {
		return fmt.Errorf("err : %v failed to query upcoming season weeks ", err)
	}

	for _, x := range data {

		log.Printf("x.LeagueID:%s, x.SeasonWeekID:%s, x.SeasonID:%s", x.LeagueID, x.SeasonWeekID, x.SeasonID)

		n := 0

		matchMap, err := s.Validate3(ctx, x.LeagueID, oddsSortedSet, oddsu15Set)
		if err != nil {
			log.Printf("Err : %v unable to create season week >>>>> ", err)
		} else {

			log.Println("matchMap ======>>> ", len(matchMap), " || seasonWeekData  ===> ", len(data))

			//log.Printf("seasonWeekID: %s, StartTime: %s, WeekNumber:%s, Status:%s",
			//	x.SeasonWeekID, x.StartTime, x.WeekNumber, x.Status)

			// Get all matches for this week.

			hh := oddsFiles.FinalSeasonWeek{}
			wo := oddsFiles.FinalSeasonWeekWO{}
			lsc := oddsFiles.FinalSeasonWeekLS{}

			hh.SeasonWeeKID = x.SeasonWeekID
			hh.StartTime = x.StartTime
			hh.EndTime = x.EndTime
			hh.MatchDay = x.WeekNumber
			hh.SeasonID = x.SeasonID

			wo.SeasonWeeKID = x.SeasonWeekID
			wo.StartTime = x.StartTime
			wo.EndTime = x.EndTime
			wo.MatchDay = x.WeekNumber
			wo.SeasonID = x.SeasonID

			lsc.SeasonWeeKID = x.SeasonWeekID
			lsc.StartTime = x.StartTime
			lsc.EndTime = x.EndTime
			lsc.MatchDay = x.WeekNumber
			lsc.SeasonID = x.SeasonID

			matches, err := s.matchesMysql.GetSeasonWeekGames(ctx, x.SeasonWeekID, x.SeasonID)
			if err != nil {
				log.Printf("Err : %v failed to query matches", err)
			}

			log.Println("matches =====>>>>>>>>", matches)

			for _, g := range matches {

				log.Printf("matchID:%s, homeTeamID:%s, awayTeamID:%s, seasonWeekID:%s",
					g.MatchID, g.HomeTeamID, g.AwayTeamID, g.SeasonWeekID)

				// Start creating magic here
				//leagueID := "1"
				homeTeamName, homeTeamAlias := s.TeamInfo(x.LeagueID, g.HomeTeamID)
				awayTeamName, awayTeamAlias := s.TeamInfo(x.LeagueID, g.AwayTeamID)

				matches := oddsFiles.FinalMatches{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				woMatches := oddsFiles.FinalMatchesWO{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				lsMatches := oddsFiles.FinalMatchesLS{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				fd := matchMap[n]

				log.Println("odds ---> ", len(fd.ValidateKeys.Odds))
				oddsFactor := 0.01
				mts, err := processOdds.New(fd.ValidateKeys.Odds, fd.ValidateKeys.Wo, fd.ValidateKeys.Ls, oddsFactor)
				if err != nil {
					log.Printf("Err : %v failed to initialize odds ", err)
				}

				mtk, winningOutcomes, liveScores, err := mts.FormulateOdds2(ctx) // FormulateOdds(ctx)
				if err != nil {
					log.Printf("Err : %v failed to formulate odds ", err)
				}

				matches.FinalMarkets = mtk
				woMatches.FinalScore = winningOutcomes
				lsMatches.FinalLiveScores = liveScores

				log.Printf("oddsKey :::--> %s ", fd.OddsKey)
				log.Println("liveScores ---> ", liveScores)

				n++

				hh.FinalMatches = append(hh.FinalMatches, matches)
				wo.FinalMatchesWO = append(wo.FinalMatchesWO, woMatches)
				lsc.FinalMatchesLS = append(lsc.FinalMatchesLS, lsMatches)

				// Start working on winning outcomes

				// Update this match as processed

				status := "active"
				updated, err := s.matchesMysql.UpdateGameStatus(ctx, status, g.MatchID)
				if err != nil {
					log.Printf("Err : %v failed to update processed season week", err)
				}

				log.Printf("updated record : %d", updated)
			}

			// Save Match odds into redis for further use

			sTime, err := time.Parse("2006-01-02 15:04:05", x.StartTime)
			if err != nil {
				log.Printf("Err : %v failed to convert string to time", err)
			}

			keyName := fmt.Sprintf("%s_%s_%s", "pr_odds", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> Odds key saved >>>> %s", keyName)

			oddsData, err := json.Marshal(hh)
			if err != nil {
				log.Printf("Err: %v failed to marshall odds json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyName, string(oddsData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to odds set", err)
				}
			}

			keyNameWO := fmt.Sprintf("%s_%s_%s", "pr_wo", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> WinningOutcome key saved >>>> %s", keyNameWO)

			winningOutcomesData, err := json.Marshal(wo)
			if err != nil {
				log.Printf("Err: %v failed to marshall wo json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyNameWO, string(winningOutcomesData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to winning outcome set", err)
				}
			}

			keyNameLS := fmt.Sprintf("%s_%s_%s", "pr_ls", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> LiveScore key saved >>>> %s", keyNameLS)

			liveScoresData, err := json.Marshal(lsc)
			if err != nil {
				log.Printf("Err: %v failed to marshall wo json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyNameLS, string(liveScoresData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to save live score set", err)
				}
			}

			daysListKeys := fmt.Sprintf("%s_%s", "pr_keys", sTime.Format("2006-01-02"))
			daysListValues := fmt.Sprintf("%s_%s_%s", "pr_keys", sTime.Format("2006-01-02"), x.SeasonWeekID)

			err = s.redisConn.ZAdd(ctx, daysListKeys, "1", daysListValues)
			if err != nil {
				log.Printf("Err: %v failed to save into todays list", err)
			}

			status := "active"
			updated, err := s.seasonWeekMysql.UpdateSsnWeekStatus(ctx, x.SeasonWeekID, x.SeasonID, status)
			if err != nil {
				log.Printf("Err : %v failed to update processed season week", err)
			}

			log.Printf("updated record : %d", updated)
		}
	}

	return nil
}

// GetUpcomingSeasonWeeks : used to return upcoming season week.
func (s *ProcessKeyService) GetUpcomingSeasonWeeks3(ctx context.Context, oddsSortedSet, oddsu15Set string) error {

	data, err := s.seasonWeekMysql.UpcomingSsnWeeks2(ctx)
	if err != nil {
		return fmt.Errorf("err : %v failed to query upcoming season weeks ", err)
	}

	for _, x := range data {

		log.Printf("x.LeagueID:%s, x.SeasonWeekID:%s, x.SeasonID:%s competitionID:%s, roundNumberID:%s",
			x.LeagueID, x.SeasonWeekID, x.SeasonID, x.CompetitionID, x.RoundNumberID)

		// Pull the data from our db on how we distribute our goals

		distr, err := s.mrsMysql.GoalDistribution(ctx, x.RoundNumberID, x.CompetitionID)
		if err != nil {
			return fmt.Errorf("err : %v failed to query goal distribution ", err)
		}

		n := 0

		matchMap, err := s.Validate4(ctx, x.LeagueID, oddsSortedSet, oddsu15Set, distr, x.CompetitionID)
		if err != nil {
			log.Printf("Err : %v unable to create season week >>>>> ", err)
		} else {

			log.Println("matchMap ======>>> ", len(matchMap), " || seasonWeekData  ===> ", len(data))

			//log.Printf("seasonWeekID: %s, StartTime: %s, WeekNumber:%s, Status:%s",
			//	x.SeasonWeekID, x.StartTime, x.WeekNumber, x.Status)

			// Get all matches for this week.

			hh := oddsFiles.FinalSeasonWeek{}
			wo := oddsFiles.FinalSeasonWeekWO{}
			lsc := oddsFiles.FinalSeasonWeekLS{}

			hh.SeasonWeeKID = x.SeasonWeekID
			hh.StartTime = x.StartTime
			hh.EndTime = x.EndTime
			hh.MatchDay = x.WeekNumber
			hh.SeasonID = x.SeasonID

			wo.SeasonWeeKID = x.SeasonWeekID
			wo.StartTime = x.StartTime
			wo.EndTime = x.EndTime
			wo.MatchDay = x.WeekNumber
			wo.SeasonID = x.SeasonID

			lsc.SeasonWeeKID = x.SeasonWeekID
			lsc.StartTime = x.StartTime
			lsc.EndTime = x.EndTime
			lsc.MatchDay = x.WeekNumber
			lsc.SeasonID = x.SeasonID

			matches, err := s.matchesMysql.GetSeasonWeekGames(ctx, x.SeasonWeekID, x.SeasonID)
			if err != nil {
				log.Printf("Err : %v failed to query matches", err)
			}

			log.Println("matches =====>>>>>>>>", matches)

			for _, g := range matches {

				log.Printf("matchID:%s, homeTeamID:%s, awayTeamID:%s, seasonWeekID:%s",
					g.MatchID, g.HomeTeamID, g.AwayTeamID, g.SeasonWeekID)

				// Start creating magic here
				//leagueID := "1"
				homeTeamName, homeTeamAlias := s.TeamInfo(x.LeagueID, g.HomeTeamID)
				awayTeamName, awayTeamAlias := s.TeamInfo(x.LeagueID, g.AwayTeamID)

				matches := oddsFiles.FinalMatches{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				woMatches := oddsFiles.FinalMatchesWO{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				lsMatches := oddsFiles.FinalMatchesLS{
					MatchID:   g.MatchID,
					HomeID:    g.HomeTeamID,
					HomeAlias: homeTeamAlias,
					HomeTeam:  homeTeamName,
					AwayID:    g.AwayTeamID,
					AwayAlias: awayTeamAlias,
					AwayTeam:  awayTeamName,
				}

				fd := matchMap[n]

				log.Println("odds ---> ", len(fd.ValidateKeys.Odds))
				oddsFactor := 0.01
				mts, err := processOdds.New(fd.ValidateKeys.Odds, fd.ValidateKeys.Wo, fd.ValidateKeys.Ls, oddsFactor)
				if err != nil {
					log.Printf("Err : %v failed to initialize odds ", err)
				}

				mtk, winningOutcomes, liveScores, err := mts.FormulateOdds2(ctx) // FormulateOdds(ctx)
				if err != nil {
					log.Printf("Err : %v failed to formulate odds ", err)
				}

				matches.FinalMarkets = mtk
				woMatches.FinalScore = winningOutcomes
				lsMatches.FinalLiveScores = liveScores

				log.Printf("oddsKey :::--> %s ", fd.OddsKey)
				log.Println("liveScores ---> ", liveScores)

				n++

				hh.FinalMatches = append(hh.FinalMatches, matches)
				wo.FinalMatchesWO = append(wo.FinalMatchesWO, woMatches)
				lsc.FinalMatchesLS = append(lsc.FinalMatchesLS, lsMatches)

				// Start working on winning outcomes

				// Update this match as processed

				status := "active"
				updated, err := s.matchesMysql.UpdateGameStatus(ctx, status, g.MatchID)
				if err != nil {
					log.Printf("Err : %v failed to update processed season week", err)
				}

				log.Printf("updated record : %d", updated)
			}

			// Save Match odds into redis for further use

			sTime, err := time.Parse("2006-01-02 15:04:05", x.StartTime)
			if err != nil {
				log.Printf("Err : %v failed to convert string to time", err)
			}

			keyName := fmt.Sprintf("%s_%s_%s", "pr_odds", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> Odds key saved >>>> %s", keyName)

			oddsData, err := json.Marshal(hh)
			if err != nil {
				log.Printf("Err: %v failed to marshall odds json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyName, string(oddsData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to odds set", err)
				}
			}

			keyNameWO := fmt.Sprintf("%s_%s_%s", "pr_wo", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> WinningOutcome key saved >>>> %s", keyNameWO)

			winningOutcomesData, err := json.Marshal(wo)
			if err != nil {
				log.Printf("Err: %v failed to marshall wo json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyNameWO, string(winningOutcomesData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to winning outcome set", err)
				}
			}

			keyNameLS := fmt.Sprintf("%s_%s_%s", "pr_ls", sTime.Format("2006-01-02"), x.SeasonWeekID)
			log.Printf(">>> LiveScore key saved >>>> %s", keyNameLS)

			liveScoresData, err := json.Marshal(lsc)
			if err != nil {
				log.Printf("Err: %v failed to marshall wo json", err)
			} else {

				expiry := "108000"
				err := s.redisConn.SetWithExpiry(ctx, keyNameLS, string(liveScoresData), expiry)
				if err != nil {
					log.Printf("Err: %v failed to save live score set", err)
				}
			}

			daysListKeys := fmt.Sprintf("%s_%s", "pr_keys", sTime.Format("2006-01-02"))
			daysListValues := fmt.Sprintf("%s_%s_%s", "pr_keys", sTime.Format("2006-01-02"), x.SeasonWeekID)

			err = s.redisConn.ZAdd(ctx, daysListKeys, "1", daysListValues)
			if err != nil {
				log.Printf("Err: %v failed to save into todays list", err)
			}

			status := "active"
			updated, err := s.seasonWeekMysql.UpdateSsnWeekStatus(ctx, x.SeasonWeekID, x.SeasonID, status)
			if err != nil {
				log.Printf("Err : %v failed to update processed season week", err)
			}

			log.Printf("updated record : %d", updated)
		}
	}

	return nil
}

func (s *ProcessKeyService) TeamInfo(leagueID string, teamID string) (string, string) {
	// home_team,home_alias

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

// Validate : rewrites odds the right way
func (s *ProcessKeyService) Validate(ctx context.Context, leagueID, oddsSortedSet, woSortedSet, liveScoreSortedSet string) (map[int]oddsFiles.CheckKeys, error) {

	m := make(map[int]oddsFiles.CheckKeys)

	selF := []int{}
	if leagueID == "1" || leagueID == "2" || leagueID == "4" {
		selF = append(selF, 10)
	} else {
		selF = append(selF, 9)
	}

	fetched := selF[0]
	log.Printf(" fetched +++++++>>>>>> %d", fetched)

	keysList := []oddsFiles.CheckKeys{}
	data, err := s.redisConn.GetZRangeWithLimit(ctx, oddsSortedSet, 12) // loop through with extra values to ensure each season week has all games.
	if err != nil {
		return m, fmt.Errorf("err : %v failed to read from %s z range", err, oddsSortedSet)
	}

	if len(data) < 10 {
		return m, fmt.Errorf("*** There are no enough matches ready to create a seen week *** count **** %d", len(data))
	}

	num := 0

	for _, o := range data {

		log.Printf(">>>>>> data fetched >>>> %s", o)

		if num < fetched {

			yy := oddsFiles.CheckKeys{}
			yy.OddsKey = o

			parentID := strings.Split(o, "O:") // example tzO:31475633 or keO:31475634
			if len(parentID) == 2 {

				// Sanitize this matches from mysql records. Will skip if match already used in the last 7 days.

				status, count, message, err := s.checkMatchesMysql.MatchExist(ctx, parentID[0], parentID[1])
				if err != nil {
					return m, fmt.Errorf("Err : %v on checking if match exists ", err)
				}

				log.Printf("status : %t, count : %d, message : %s", status, count, message)

				// if count > 0 {

				// 	_, err = s.redisConn.ZRem(ctx, oddsSortedSet, o)
				// 	if err != nil {
				// 		log.Printf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
				// 	}

				// 	return m, fmt.Errorf("Err : %v match already used recently country : %s matchID %s ", err, parentID[0], parentID[1])
				// }

				// 1 . Find Odds

				log.Printf("oddKey --> %s", o)
				oddsData, err := s.redisConn.Get(ctx, o)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match odds from redis ", err)
				}

				// 2. Find Winning outcomes

				woKey := fmt.Sprintf("%s%s%s", parentID[0], "Wo:", parentID[1])
				log.Printf("woKey --> %s", woKey)
				woData, err := s.redisConn.Get(ctx, woKey)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match winning outcomes from redis ", err)
				}

				// 3. Find Live scores

				lsKey := fmt.Sprintf("%s%s%s", parentID[0], "Ls:", parentID[1])
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

				// Save this match as used to avoid repetition in the coming days.

				matchDate := time.Now().Format("2006-01-02")
				cm, err := checkMatches.NewCheckMatches(parentID[0], parentID[1], matchDate)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to instantiate checkMatches struct", err)
				}

				lastID, err := s.checkMatchesMysql.Save(ctx, *cm)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to save into checkMatches table", err)
				}

				log.Printf("Last inserted id %d into checkMatches tbl", lastID)

				_, err = s.redisConn.ZRem(ctx, oddsSortedSet, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
				}

				/*_, err = s.redisConn.ZRem(ctx, woSortedSet, woKey)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, woSortedSet)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, woSortedSet)
				}

				_, err = s.redisConn.ZRem(ctx, liveScoreSortedSet, lsKey)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, liveScoreSortedSet)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, liveScoreSortedSet)
				}*/

				usedOddsSortedSet := fmt.Sprintf("%s_%s", oddsSortedSet, "USED")
				err = s.redisConn.ZAdd(ctx, usedOddsSortedSet, "1", o)
				if err != nil {
					log.Printf("Err : %v failed to add %s sorted set %s", err, o, liveScoreSortedSet)
					return m, fmt.Errorf("Err : %v ailed to add %s sorted set %s", err, o, liveScoreSortedSet)
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

func (s *ProcessKeyService) Validate3(ctx context.Context, leagueID, oddsSortedSet, oddsu15Set string) (map[int]oddsFiles.CheckKeys, error) {

	m := make(map[int]oddsFiles.CheckKeys)

	selF := []int{}
	if leagueID == "1" || leagueID == "2" || leagueID == "4" {
		selF = append(selF, 10)
	} else {
		selF = append(selF, 9)
	}

	fetched := selF[0]
	log.Printf(" fetched +++++>>>>>> %d", fetched)

	// Get the games ration for over TG25

	keysList := []oddsFiles.CheckKeys{}
	data, err := s.DecideRatio(ctx, oddsSortedSet, oddsu15Set, fetched)
	if err != nil {
		return m, fmt.Errorf("err : %v failed to read from %s z range", err, oddsSortedSet)
	}

	if len(data) < 10 {
		return m, fmt.Errorf("*** There are no enough matches ready to create a seen week *** count **** %d", len(data))
	}

	num := 0

	for _, o := range data {

		log.Printf(">>>>>> data fetched >>>>> %s", o)

		if num < fetched {

			yy := oddsFiles.CheckKeys{}
			yy.OddsKey = o

			parentID := strings.Split(o, "O:") // example tzO:31475633 or keO:31475634
			if len(parentID) == 2 {

				// Sanitize this matches from mysql records. Will skip if match already used in the last 7 days.

				status, count, message, err := s.checkMatchesMysql.MatchExist(ctx, parentID[0], parentID[1])
				if err != nil {
					return m, fmt.Errorf("Err : %v on checking if match exists ", err)
				}

				log.Printf("status : %t, count : %d, message : %s", status, count, message)

				// 1 . Find Odds

				log.Printf("oddKey --> %s", o)
				oddsData, err := s.redisConn.Get(ctx, o)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match odds from redis ", err)
				}

				// 2. Find Winning outcomes

				woKey := fmt.Sprintf("%s%s%s", parentID[0], "Wo:", parentID[1])
				log.Printf("woKey --> %s", woKey)
				woData, err := s.redisConn.Get(ctx, woKey)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match winning outcomes from redis ", err)
				}

				// 3. Find Live scores

				lsKey := fmt.Sprintf("%s%s%s", parentID[0], "Ls:", parentID[1])
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

				// Save this match as used to avoid repetition in the coming days.

				matchDate := time.Now().Format("2006-01-02")
				cm, err := checkMatches.NewCheckMatches(parentID[0], parentID[1], matchDate)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to instantiate checkMatches struct", err)
				}

				lastID, err := s.checkMatchesMysql.Save(ctx, *cm)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to save into checkMatches table", err)
				}

				log.Printf("Last inserted id %d into checkMatches tbl", lastID)

				_, err = s.redisConn.ZRem(ctx, oddsSortedSet, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
				}

				// Remove from the Over and Under 2.5 sets

				tgOver25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGO25")
				tgUnder25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGU25")
				tgUnder15 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGU15")
				tgUnder0 := fmt.Sprintf("%s_%s", oddsSortedSet, "0")
				tgUnder1 := fmt.Sprintf("%s_%s", oddsSortedSet, "1")
				tgUnder2 := fmt.Sprintf("%s_%s", oddsSortedSet, "2")
				tgUnder3 := fmt.Sprintf("%s_%s", oddsSortedSet, "3")
				tgUnder4 := fmt.Sprintf("%s_%s", oddsSortedSet, "4")
				tgUnder5 := fmt.Sprintf("%s_%s", oddsSortedSet, "5")
				tgUnder6 := fmt.Sprintf("%s_%s", oddsSortedSet, "6")
				tgUnder7 := fmt.Sprintf("%s_%s", oddsSortedSet, "7")

				_, err = s.redisConn.ZRem(ctx, tgOver25, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgOver25)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgOver25)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder15, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder15)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder15)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder25, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder25)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder25)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder0, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder0)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder0)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder1, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder1)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder1)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder2, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder2)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder2)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder3, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder3)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder3)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder4, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder4)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder4)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder5, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder5)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder5)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder6, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder6)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder6)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder7, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder7)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder7)
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

func (s *ProcessKeyService) Validate4(ctx context.Context, leagueID, oddsSortedSet, oddsu15Set string, distr []mrs.Mrs, competitionID string) (map[int]oddsFiles.CheckKeys, error) {

	m := make(map[int]oddsFiles.CheckKeys)

	selF := []int{}
	if leagueID == "1" || leagueID == "2" || leagueID == "4" {
		selF = append(selF, 10)
	} else {
		selF = append(selF, 9)
	}

	fetched := selF[0]
	log.Printf(" fetched +++++>>>>>> %d", fetched)

	// Get the games ration for over TG25

	keysList := []oddsFiles.CheckKeys{}
	data, err := s.DecideRatio4(ctx, oddsSortedSet, oddsu15Set, fetched, distr, competitionID)
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

			parentID := strings.Split(o, "O:") // example tzO:31475633 or keO:31475634
			if len(parentID) == 2 {

				// Sanitize this matches from mysql records. Will skip if match already used in the last 7 days.

				status, count, message, err := s.checkMatchesMysql.MatchExist(ctx, parentID[0], parentID[1])
				if err != nil {
					return m, fmt.Errorf("Err : %v on checking if match exists ", err)
				}

				log.Printf("status : %t, count : %d, message : %s", status, count, message)

				// 1 . Find Odds

				log.Printf("oddKey --> %s", o)
				oddsData, err := s.redisConn.Get(ctx, o)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match odds from redis ", err)
				}

				// 2. Find Winning outcomes

				woKey := fmt.Sprintf("%s%s%s", parentID[0], "Wo:", parentID[1])
				log.Printf("woKey --> %s", woKey)
				woData, err := s.redisConn.Get(ctx, woKey)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to get match winning outcomes from redis ", err)
				}

				// 3. Find Live scores

				lsKey := fmt.Sprintf("%s%s%s", parentID[0], "Ls:", parentID[1])
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

				// Save this match as used to avoid repetition in the coming days.

				matchDate := time.Now().Format("2006-01-02")
				cm, err := checkMatches.NewCheckMatches(parentID[0], parentID[1], matchDate)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to instantiate checkMatches struct", err)
				}

				lastID, err := s.checkMatchesMysql.Save(ctx, *cm)
				if err != nil {
					return m, fmt.Errorf("Err : %v failed to save into checkMatches table", err)
				}

				log.Printf("Last inserted id %d into checkMatches tbl", lastID)

				_, err = s.redisConn.ZRem(ctx, oddsSortedSet, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, oddsSortedSet)
				}

				// Remove from the Over and Under 2.5 sets

				tgOver25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGO25")
				tgUnder25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGU25")
				tgUnder15 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGU15")
				tgUnder0 := fmt.Sprintf("%s_%s", oddsSortedSet, "0")
				tgUnder1 := fmt.Sprintf("%s_%s", oddsSortedSet, "1")
				tgUnder2 := fmt.Sprintf("%s_%s", oddsSortedSet, "2")
				tgUnder3 := fmt.Sprintf("%s_%s", oddsSortedSet, "3")
				tgUnder4 := fmt.Sprintf("%s_%s", oddsSortedSet, "4")
				tgUnder5 := fmt.Sprintf("%s_%s", oddsSortedSet, "5")
				tgUnder6 := fmt.Sprintf("%s_%s", oddsSortedSet, "6")
				tgUnder7 := fmt.Sprintf("%s_%s", oddsSortedSet, "7")

				_, err = s.redisConn.ZRem(ctx, tgOver25, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgOver25)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgOver25)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder15, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder15)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder15)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder25, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder25)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder25)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder0, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder0)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder0)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder1, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder1)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder1)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder2, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder2)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder2)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder3, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder3)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder3)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder4, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder4)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder4)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder5, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder5)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder5)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder6, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder6)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder6)
				}

				_, err = s.redisConn.ZRem(ctx, tgUnder7, o)
				if err != nil {
					log.Printf("Err : %v failed to delete from %s z range", err, tgUnder7)
					return m, fmt.Errorf("Err : %v failed to delete from %s z range", err, tgUnder7)
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

// TotalGoalsPerSession : this helps with distribution of total goals per match session.
func (s *ProcessKeyService) OddsCombination(ctx context.Context) []int {

	data := []int{
		2, 2, 2, 3, 0, 0, 0, 2, 3, 4, 2, 2,
		1, 1, 2, 2, 2, 3, 3, 3, 2, 2, 0, 3,
		0, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2,
		3, 3, 3, 3, 4, 4, 0, 1, 1, 1, 2, 3,
		4, 2, 3, 3, 1, 1, 0, 2, 2, 3, 3, 2,
		3, 3, 2, 3, 3, 2, 2, 2, 2, 2, 2, 1,
	}

	return data
}

func (s *ProcessKeyService) DecideRatio(ctx context.Context, oddsSortedSet string, oddsu15Set string, totalGames int) ([]string, error) {

	list := []string{}

	combinations := s.OddsCombination(ctx)
	max := len(combinations)
	selectedCombination := rand.IntN(max)
	log.Printf("rand picked index %d", selectedCombination)

	selCombination := combinations[selectedCombination]
	log.Printf("rand picked ov15 index %d | selected combination %d",
		selectedCombination, selCombination)

	//
	// query over under 1.5
	//

	data, err := s.redisConn.GetZRangeWithLimit(ctx, oddsu15Set, selCombination) // To limit loss via this market
	if err != nil {
		return list, fmt.Errorf("err : %v failed to read from %s z range", err, oddsu15Set)
	}

	//
	// Pick scores over 1.5 (ov25 ie 2 goals and more 0 and 1 taken care of above)
	//

	over15Limit := totalGames - selCombination //totalGames - selCombination

	goals := s.TotalGoalsPerSession(ctx)
	maxG := len(goals)
	selectedRations := rand.IntN(maxG)

	rgO25 := goals[selectedRations] // + 1
	rgU25 := over15Limit - rgO25    //+ 1

	log.Printf("Selected ration 1.5 %d tgO25 : %d | tgU25 %d", selCombination, rgO25, rgU25)

	// Now figure out how the matches will be arranged randomly.

	oddsOv25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGO25") // sum of goals more than 2
	oddsU25 := fmt.Sprintf("%s_%s", oddsSortedSet, "2")      // 2 goals in total ie 1-1,0-2,2-0,

	data2, err := s.redisConn.GetZRangeWithLimit(ctx, oddsOv25, rgO25)
	if err != nil {
		return list, fmt.Errorf("err : %v failed to read from %s z range", err, oddsOv25)
	}

	data3, err := s.redisConn.GetZRangeWithLimit(ctx, oddsU25, rgU25)
	if err != nil {
		return list, fmt.Errorf("Err : %v failed to read from %s z range", err, oddsU25)
	}

	log.Println("under 1.5 list >>> ", data, "list ", oddsu15Set)

	log.Println("under 2.5 list >>> ", data2, "list ", oddsU25)

	log.Println("over 2.5 list >>> ", data3, "list ", oddsOv25)

	for _, x := range data {
		list = append(list, x)
	}

	for _, xx := range data2 {
		list = append(list, xx)
	}

	for _, xxx := range data3 {
		list = append(list, xxx)
	}

	log.Println("List before shuffle : ", list)

	for i := len(list) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.IntN(i + 1)
		list[i], list[j] = list[j], list[i]
	}

	log.Println("List after shuffle : ", list)

	return list, nil

}

func (s *ProcessKeyService) DecideRatio2(ctx context.Context, oddsSortedSet string, oddsu15Set string, totalGames int, distr []mrs.Mrs, competitionID string) ([]string, error) {
	list := []string{}

	// Check if the number of games are correct

	ss := 0
	for _, d := range distr {

		matchCount, err := strconv.Atoi(d.GoalCount)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to convert matchCount", err)
		}
		ss += matchCount

	}

	log.Printf("ss returned %d", ss)

	// Sanitize the list before proceeding.

	sanitizedGoals := []mrs.Mrs{}

	if competitionID == "1" || competitionID == "2" || competitionID == "4" {

		if ss != 10 {

			divider := ss / 10

			for _, r := range distr {

				matchCount, err := strconv.Atoi(r.GoalCount)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to convert matchCount", err)
				}

				accurateCount := matchCount / divider
				log.Printf("final Goal count %d", accurateCount)

				goalCount := fmt.Sprintf("%d", accurateCount)

				dd := mrs.Mrs{
					MrID:          r.MrID,
					RoundNumberID: r.RoundNumberID,
					TotalGoals:    r.TotalGoals,
					GoalCount:     goalCount,
					StartTime:     r.StartTime,
					CompetitionID: r.CompetitionID,
					RawScores:     r.RawScores,
					Created:       r.Created,
					Modified:      r.Modified,
				}

				sanitizedGoals = append(sanitizedGoals, dd)
			}

		} else {

			for _, r := range distr {
				sanitizedGoals = append(sanitizedGoals, r)
			}
		}

	} else {

		if ss != 9 {

			divider := ss / 9

			for _, r := range distr {

				matchCount, err := strconv.Atoi(r.GoalCount)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to convert matchCount", err)
				}

				accurateCount := matchCount / divider
				log.Printf("final Goal count %d", accurateCount)

				goalCount := fmt.Sprintf("%d", accurateCount)

				dd := mrs.Mrs{
					MrID:          r.MrID,
					RoundNumberID: r.RoundNumberID,
					TotalGoals:    r.TotalGoals,
					GoalCount:     goalCount,
					StartTime:     r.StartTime,
					CompetitionID: r.CompetitionID,
					RawScores:     r.RawScores,
					Created:       r.Created,
					Modified:      r.Modified,
				}

				sanitizedGoals = append(sanitizedGoals, dd)
			}

		} else {

			for _, r := range distr {
				sanitizedGoals = append(sanitizedGoals, r)
			}

		}
	}

	sanitizedList := 0
	for _, d := range sanitizedGoals {

		matchCount, err := strconv.Atoi(d.GoalCount)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to convert matchCount", err)
		}
		sanitizedList += matchCount

	}

	log.Printf("sanitizedList returned %d", sanitizedList)

	if competitionID == "1" || competitionID == "2" || competitionID == "4" {

		if sanitizedList != 10 {
			return list, fmt.Errorf("number of games %s | available %d", "10", sanitizedList)
		}

	} else {

		if sanitizedList != 9 {
			return list, fmt.Errorf("number of games %s | available %d", "9", sanitizedList)
		}

	}

	log.Printf("*** I get here ***")
	log.Println("*** data returned after cleanup ***", sanitizedGoals)

	// Query the games from Redis now
	for _, d := range sanitizedGoals {

		totalGoalsSet := fmt.Sprintf("%s_%s", oddsSortedSet, d.TotalGoals) // Eg 0,1,2,3,4,5,6

		matchCountLimit, err := strconv.Atoi(d.GoalCount)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to convert matchCount >>> ", err)
		}

		log.Printf("totalGoalsSet : %s, matchCountLimit : %d", totalGoalsSet, matchCountLimit)
		data, err := s.redisConn.GetZRangeWithLimit(ctx, totalGoalsSet, matchCountLimit)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to read from %d z range", err, matchCountLimit)
		}

		log.Println("data returned : ", data)

		for _, xx := range data {
			list = append(list, xx)
		}

	}

	//
	log.Println("List before shuffle : ", list)
	for i := len(list) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.IntN(i + 1)
		list[i], list[j] = list[j], list[i]
	}

	log.Println("List after shuffle : ", list)

	return list, nil

}

func (s *ProcessKeyService) DecideRatio4(ctx context.Context, oddsSortedSet string, oddsu15Set string, totalGames int, distr []mrs.Mrs, competitionID string) ([]string, error) {
	list := []string{}

	// Check if the number of games are correct

	ss := 0
	for _, d := range distr {

		matchCount, err := strconv.Atoi(d.GoalCount)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to convert matchCount", err)
		}
		ss += matchCount

	}

	log.Printf("ss returned %d", ss)

	// Sanitize the list before proceeding.

	sanitizedGoals := []mrs.Mrs{}

	if competitionID == "1" || competitionID == "2" || competitionID == "4" {

		if ss != 10 {

			divider := ss / 10

			for _, r := range distr {

				matchCount, err := strconv.Atoi(r.GoalCount)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to convert matchCount", err)
				}

				accurateCount := matchCount / divider
				log.Printf("final Goal count %d", accurateCount)

				goalCount := fmt.Sprintf("%d", accurateCount)

				dd := mrs.Mrs{
					MrID:          r.MrID,
					RoundNumberID: r.RoundNumberID,
					TotalGoals:    r.TotalGoals,
					GoalCount:     goalCount,
					StartTime:     r.StartTime,
					CompetitionID: r.CompetitionID,
					RawScores:     r.RawScores,
					Created:       r.Created,
					Modified:      r.Modified,
				}

				sanitizedGoals = append(sanitizedGoals, dd)
			}

		} else {

			for _, r := range distr {
				sanitizedGoals = append(sanitizedGoals, r)
			}
		}

	} else {

		if ss != 9 {

			divider := ss / 9

			for _, r := range distr {

				matchCount, err := strconv.Atoi(r.GoalCount)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to convert matchCount", err)
				}

				accurateCount := matchCount / divider
				log.Printf("final Goal count %d", accurateCount)

				goalCount := fmt.Sprintf("%d", accurateCount)

				dd := mrs.Mrs{
					MrID:          r.MrID,
					RoundNumberID: r.RoundNumberID,
					TotalGoals:    r.TotalGoals,
					GoalCount:     goalCount,
					StartTime:     r.StartTime,
					CompetitionID: r.CompetitionID,
					RawScores:     r.RawScores,
					Created:       r.Created,
					Modified:      r.Modified,
				}

				sanitizedGoals = append(sanitizedGoals, dd)
			}

		} else {

			for _, r := range distr {
				sanitizedGoals = append(sanitizedGoals, r)
			}

		}
	}

	sanitizedList := 0
	for _, d := range sanitizedGoals {

		matchCount, err := strconv.Atoi(d.GoalCount)
		if err != nil {
			return list, fmt.Errorf("err : %v failed to convert matchCount", err)
		}
		sanitizedList += matchCount

	}

	log.Printf("sanitizedList returned %d", sanitizedList)

	if competitionID == "1" || competitionID == "2" || competitionID == "4" {

		if sanitizedList != 10 {
			return list, fmt.Errorf("number of games %s | available %d", "10", sanitizedList)
		}

	} else {

		if sanitizedList != 9 {
			return list, fmt.Errorf("number of games %s | available %d", "9", sanitizedList)
		}

	}

	log.Printf("*** I get here ***")
	log.Println("*** data returned after cleanup ***", sanitizedGoals)

	// Query the games from Redis now
	for _, d := range sanitizedGoals {

		scores := strings.Split(d.RawScores, "**")
		// raw_scores: 52824375#1#2**52824376#1#1**52824377#2#1**52824378#2#2**52824379#1#0**52824380#0#1**52824381#2#0**52824382#1#1**52824383#3#1**52824384#1#1

		for _, cc := range scores {

			// Example 52824375#1#2

			goals := strings.Split(cc, "#")
			if len(goals) == 3 {

				hSc := goals[1]
				aSc := goals[2]

				log.Printf("**** HomeScore %s | Away Score %s ****", hSc, aSc)

				homeGoals, err := strconv.Atoi(hSc)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to convert hsc to int", err)
				}

				awayGoals, err := strconv.Atoi(aSc)
				if err != nil {
					return list, fmt.Errorf("err : %v failed to convert asc to int", err)
				}

				totalGoals := homeGoals + awayGoals

				if totalGoals == 0 {

					sel0 := fmt.Sprintf("%s_%d", oddsSortedSet, 0)
					data, err := s.redisConn.GetZRangeWithLimit(ctx, sel0, 1)
					if err != nil {
						return list, fmt.Errorf("err : %v failed to read from %s z range", err, sel0)
					}

					log.Println("*** data returned sel0 *** ", data)
					for _, xx := range data {
						list = append(list, xx)
						s.RemoveUsedKeys(ctx, sel0, xx)
					}

					// Delete this record so we dont reuse it.

				} else if totalGoals == 1 {

					if homeGoals == 1 {

						sel1h := fmt.Sprintf("%s_%s_%s", oddsSortedSet, "1", "h")
						data, err := s.redisConn.GetZRangeWithLimit(ctx, sel1h, 1)
						if err != nil {
							return list, fmt.Errorf("err : %v failed to read from %s z range", err, sel1h)
						}

						log.Println("*** data returned sel1h *** ", data)
						for _, xx := range data {
							list = append(list, xx)
							s.RemoveUsedKeys(ctx, sel1h, xx)
						}

					} else {

						sel1a := fmt.Sprintf("%s_%s_%s", oddsSortedSet, "1", "a")
						data, err := s.redisConn.GetZRangeWithLimit(ctx, sel1a, 1)
						if err != nil {
							return list, fmt.Errorf("err : %v failed to read from %s z range", err, sel1a)
						}

						log.Println("*** data returned sel1a *** ", data)
						for _, xx := range data {
							list = append(list, xx)
							s.RemoveUsedKeys(ctx, sel1a, xx)

						}

					}

				} else {

					if homeGoals > 0 && awayGoals > 0 {

						// Query from Goal Goal batch

						if homeGoals > awayGoals {

							selggh := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "gg", "h")
							data, err := s.redisConn.GetZRangeWithLimit(ctx, selggh, 1)
							if err != nil {
								return list, fmt.Errorf("err : %v failed to read from %s z range", err, selggh)
							}

							log.Println("*** data returned selggh *** ", data)
							for _, xx := range data {
								list = append(list, xx)
								s.RemoveUsedKeys(ctx, selggh, xx)
							}

						} else if awayGoals > homeGoals {

							selgga := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "gg", "a")
							data, err := s.redisConn.GetZRangeWithLimit(ctx, selgga, 1)
							if err != nil {
								return list, fmt.Errorf("err : %v failed to read from %s z range", err, selgga)
							}

							log.Println("*** data returned *** ", data)
							for _, xx := range data {
								list = append(list, xx)
								s.RemoveUsedKeys(ctx, selgga, xx)
							}

						} else {

							selggd := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "gg", "d")
							data, err := s.redisConn.GetZRangeWithLimit(ctx, selggd, 1)
							if err != nil {
								return list, fmt.Errorf("err : %v failed to read from %s z range", err, selggd)
							}

							log.Println("*** data returned *** ", data)
							for _, xx := range data {
								list = append(list, xx)
								s.RemoveUsedKeys(ctx, selggd, xx)
							}

						}

					} else {

						// Query from No goal

						if homeGoals > awayGoals {

							selngh := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "ng", "h")
							data, err := s.redisConn.GetZRangeWithLimit(ctx, selngh, 1)
							if err != nil {
								return list, fmt.Errorf("err : %v failed to read from %s z range", err, selngh)
							}

							log.Println("*** data returned selngh *** ", data)
							for _, xx := range data {
								list = append(list, xx)
								s.RemoveUsedKeys(ctx, selngh, xx)
							}

						} else if awayGoals > homeGoals {

							selnga := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "ng", "a")
							data, err := s.redisConn.GetZRangeWithLimit(ctx, selnga, 1)
							if err != nil {
								return list, fmt.Errorf("err : %v failed to read from %s z range", err, selnga)
							}

							log.Println("*** data returned selnga *** ", data)
							for _, xx := range data {
								list = append(list, xx)
								s.RemoveUsedKeys(ctx, selnga, xx)
							}

						} else {

							selngd := fmt.Sprintf("%s_%d_%s_%s", oddsSortedSet, totalGoals, "ng", "d")
							data, err := s.redisConn.GetZRangeWithLimit(ctx, selngd, 1)
							if err != nil {
								return list, fmt.Errorf("err : %v failed to read from %s z range", err, selngd)
							}

							log.Println("*** data returned *** ", data)
							for _, xx := range data {
								list = append(list, xx)
								s.RemoveUsedKeys(ctx, selngd, xx)
							}

						}
					}
				}

			} else {
				log.Printf("Wrong format %s", cc)
			}

		}

	}

	//
	log.Println("List before shuffle : ", list)
	for i := len(list) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.IntN(i + 1)
		list[i], list[j] = list[j], list[i]
	}

	log.Println("List after shuffle : ", list)

	return list, nil

}

func (s *ProcessKeyService) RemoveUsedKeys(ctx context.Context, key, value string) {
	_, err := s.redisConn.ZRem(ctx, key, value)
	if err != nil {
		log.Printf("Err : %v", err)
	} else {
		log.Printf("deleted key %s from zset %s", key, value)
	}
}

func (s *ProcessKeyService) DecideRatio3(ctx context.Context, oddsSortedSet string, oddsu15Set string, totalGames int) ([]string, error) {

	list := []string{}

	availableList := 3
	selCategory := rand.IntN(availableList)
	log.Printf("selCategory %d", selCategory)

	combinations := s.NewOddsCombination(ctx, selCategory)
	max := len(combinations)

	selectedCombination := rand.IntN(max)

	selCombination := combinations[selectedCombination]
	log.Printf("rand picked ov15 index %d | selected combination %d",
		selectedCombination, selCombination)

	//
	// query over under 1.5
	//

	data, err := s.redisConn.GetZRangeWithLimit(ctx, oddsu15Set, selCombination) // To limit loss via this market
	if err != nil {
		return list, fmt.Errorf("err : %v failed to read from %s z range", err, oddsu15Set)
	}

	//
	// Pick scores over 1.5 (ov25 ie 2 goals and more 0 and 1 taken care of above)
	//

	over15Limit := totalGames - selCombination //totalGames - selCombination

	goals := s.TotalGoalsPerSession(ctx)
	maxG := len(goals)
	selectedRations := rand.IntN(maxG)

	rgO25 := goals[selectedRations] // + 1
	rgU25 := over15Limit - rgO25    //+ 1

	log.Printf("Selected ration 1.5 %d tgO25 : %d | tgU25 %d", selCombination, rgO25, rgU25)

	// Now figure out how the matches will be arranged randomly.

	oddsOv25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGO25") // sum of goals more than 2
	oddsU25 := fmt.Sprintf("%s_%s", oddsSortedSet, "2")      // 2 goals in total ie 1-1,0-2,2-0,

	data2, err := s.redisConn.GetZRangeWithLimit(ctx, oddsOv25, rgO25)
	if err != nil {
		return list, fmt.Errorf("err : %v failed to read from %s z range", err, oddsOv25)
	}

	data3, err := s.redisConn.GetZRangeWithLimit(ctx, oddsU25, rgU25)
	if err != nil {
		return list, fmt.Errorf("err : %v failed to read from %s z range", err, oddsU25)
	}

	//
	// Add others to make sure the randomness is perfect
	//

	//
	// End of addition
	//

	log.Println("under 1.5 list >>> ", data, "list ", oddsu15Set)

	log.Println("under 2.5 list >>> ", data2, "list ", oddsU25)

	log.Println("over 2.5 list >>> ", data3, "list ", oddsOv25)

	for _, x := range data {
		list = append(list, x)
	}

	for _, xx := range data2 {
		list = append(list, xx)
	}

	for _, xxx := range data3 {
		list = append(list, xxx)
	}

	log.Println("List before shuffle : ", list)

	for i := len(list) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.IntN(i + 1)
		list[i], list[j] = list[j], list[i]
	}

	log.Println("List after shuffle : ", list)

	return list, nil

}

// TotalGoalsPerSession : this helps with distribution of total goals per match session.
func (s *ProcessKeyService) TotalGoalsPerSession(ctx context.Context) []int {

	data := []int{
		1, 1, 1, 1, 1,
		2, 2, 2, 2, 2, 2, 2, 2, 2,
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		5, 5, 5, 5, 5, 5,
		6, 6, 6,
		7, 7,
	}

	return data
}

func (s *ProcessKeyService) NewOddsCombination(ctx context.Context, sel int) []int {

	if sel == 1 {

		data := []int{
			3, 2, 2, 3, 4, 1, 4, 2, 3, 4, 2, 2,
			1, 1, 2, 4, 2, 3, 0, 3, 4, 1, 5, 3,
			5, 1, 1, 1, 1, 2, 4, 2, 0, 4, 0, 2,
			4, 3,
		}

		return data

	} else if sel == 2 {

		data := []int{
			1, 2, 1, 5, 4, 1, 3, 5, 3, 4, 5, 1,
			1, 1, 2, 3, 2, 5, 0, 3, 4, 1, 5, 3,
			5, 1, 1, 5, 1, 2, 4, 2, 5, 0, 0, 2,
			4, 3,
		}

		return data

	} else {

		data := []int{
			1, 3, 3, 2, 4, 1, 4, 3, 2, 4, 3, 3,
			1, 1, 3, 4, 3, 2, 0, 2, 4, 1, 5, 2,
			5, 1, 1, 1, 1, 3, 4, 3, 0, 4, 0, 3,
			4, 2,
		}

		return data
	}

}
