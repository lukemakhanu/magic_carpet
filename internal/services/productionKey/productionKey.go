package productionKey

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches/checkMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matches/matchesMysql"
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
		return fmt.Errorf("Err : %v failed to query upcoming season weeks ", err)
	}

	for _, x := range data {

		log.Printf("x.LeagueID::: %s, x.SeasonWeekID::: %s, x.SeasonID::: %s", x.LeagueID, x.SeasonWeekID, x.SeasonID)

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

				mtk, winningOutcomes, liveScores, err := mts.FormulateOdds(ctx)
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
			return "EVERTON", "EVE"
		} else if teamID == "1" {
			return "MANCHESTER C", "MNC"
		} else if teamID == "14" {
			return "BRIGHTON", "BRT"
		} else if teamID == "6" {
			return "ARSENAL", "ARS"
		} else if teamID == "7" {
			return "BURNLEY", "BUR"
		} else if teamID == "8" {
			return "LEICESTER", "LEI"
		} else if teamID == "4" {
			return "CHELSEA", "CHE"
		} else if teamID == "5" {
			return "TOTTENHAM", "TOT"
		} else if teamID == "18" {
			return "SOUTHAMPTON", "SOU"
		} else if teamID == "13" {
			return "NEWCASTLE", "NEW"
		} else if teamID == "12" {
			return "WEST HAM", "WHU"
		} else if teamID == "3" {
			return "LIVERPOOL", "LIV"
		} else if teamID == "15" {
			return "CRYSTAL PALACE", "CRY"
		} else if teamID == "2" {
			return "MANCHESTER U", "MNU"
		} else if teamID == "19" {
			return "WOLVERHAMPTON", "WOV"
		} else if teamID == "17" {
			return "ASTON V", "ARV"
		} else if teamID == "20" {
			return "SHEFFIELD U", "SHE"
		} else if teamID == "16" {
			return "FULHAM", "FUL"
		} else if teamID == "11" {
			return "WEST BROM", "WR"
		} else if teamID == "10" {
			return "LEEDS", "LEE"
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
			return "ELC", "ELC"
		} else if teamID == "11" {
			return "CAD", "CAD"
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
			return "HUE", "HUE"
		} else {
			return "0", "0"
		}
	case "3":
		if teamID == "1" {
			return "SSC", "SSC"
		} else if teamID == "2" {
			return "AZM", "AZM"
		} else if teamID == "3" {
			return "YNG", "YNG"
		} else if teamID == "4" {
			return "NMG", "NMG"
		} else if teamID == "5" {
			return "CST", "CST"
		} else if teamID == "6" {
			return "PTZ", "PTZ"
		} else if teamID == "7" {
			return "JKT", "JKT"
		} else if teamID == "8" {
			return "TZP", "TZP"
		} else if teamID == "9" {
			return "KGS", "KGS"
		} else if teamID == "10" {
			return "BMU", "BMU"
		} else if teamID == "11" {
			return "RSC", "RSC"
		} else if teamID == "12" {
			return "MFC", "MFC"
		} else if teamID == "13" {
			return "LFC", "LFC"
		} else if teamID == "14" {
			return "MTB", "MTB"
		} else if teamID == "15" {
			return "KMC", "KMC"
		} else if teamID == "16" {
			return "NDA", "NDA"
		} else if teamID == "17" {
			return "MBY", "MBY"
		} else if teamID == "18" {
			return "ALC", "ALC"
		} else if teamID == "19" {
			return "MBC", "MBC"
		} else if teamID == "20" {
			return "SNG", "SNG"
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
			return "SPE", "SPE"
		} else if teamID == "17" {
			return "BOL", "BOL"
		} else if teamID == "18" {
			return "CRO", "CRO"
		} else if teamID == "19" {
			return "BEN", "BEN"
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
		selF = append(selF, 10)
	}

	fetched := selF[0]
	log.Printf(" fetched +++++>>>>>> %d", fetched)

	// Get the games ration for over TG25

	keysList := []oddsFiles.CheckKeys{}
	data, err := s.DecideRatio(ctx, oddsSortedSet, fetched)
	if err != nil {
		return m, fmt.Errorf("Err : %v failed to read from %s z range", err, oddsSortedSet)
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

	return m, fmt.Errorf("Final match count : %d not enough to create season week", len(data))
}

func (s *ProcessKeyService) DecideRatio(ctx context.Context, oddsSortedSet string, totalGames int) ([]string, error) {

	list := []string{}

	goals := s.TotalGoalsPerSession(ctx)
	max := len(goals)
	selectedRations := s.NewRandomIndexes(ctx, max)

	rgO25 := goals[selectedRations[5]] + 1
	rgU25 := totalGames - rgO25 + 1

	log.Printf("Selected ration tgO25 : %d | tgU25 %d", rgO25, rgU25)

	// Now figure out how the matches will be arranged randomly.

	oddsOv25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGO25")
	oddsU25 := fmt.Sprintf("%s_%s", oddsSortedSet, "TGU25")

	data, err := s.redisConn.GetZRangeWithLimit(ctx, oddsOv25, rgO25)
	if err != nil {
		return list, fmt.Errorf("Err : %v failed to read from %s z range", err, oddsOv25)
	}

	data2, err := s.redisConn.GetZRangeWithLimit(ctx, oddsU25, rgU25)
	if err != nil {
		return list, fmt.Errorf("Err : %v failed to read from %s z range", err, oddsOv25)
	}

	for _, x := range data {
		list = append(list, x)
	}

	for _, x := range data2 {
		list = append(list, x)
	}

	log.Println("List before shuffle : ", list)

	rand.Seed(time.Now().UnixNano())
	for i := len(list) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.Intn(i + 1)
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

// NewRandomIndexes : used to create new randomization.
func (s *ProcessKeyService) NewRandomIndexes(ctx context.Context, max int) map[int]int {
	min := 1

	m := make(map[int]int)
	for x := 0; x < 10; x++ {
		rand.Seed(time.Now().UnixNano())
		val := rand.Intn(max-min+1) + min
		m[val] = val
	}

	return m
}






