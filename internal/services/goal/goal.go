package goal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/checkMatches/checkMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goals"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goals/goalsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/slowRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/slowRedis/rExec"
)

type Job struct {
	JobType  string
	WorkStr  string
	WorkStr1 string
	WorkStr2 string
	WorkStr3 string
}

// GoalConfiguration is an alias for a function that will take in a pointer to an GoalService and modify it
type GoalConfiguration func(os *GoalService) error

// GoalService is a implementation of the GoalService
type GoalService struct {
	checkMatchesMysql checkMatches.CheckMatchesRepository
	redisConn         processRedis.RunRedis
	slowRedisConn     slowRedis.SlowRedis
	goalMysql         goals.GoalsRepository
}

// NewGoalService : instantiate every connection we need to run current game service
func NewGoalService(cfgs ...GoalConfiguration) (*GoalService, error) {
	// Create the seasonService
	os := &GoalService{}
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

// WithMysqlCheckMatchesRepository : instantiates mysql to connect to matches interface
func WithMysqlCheckMatchesRepository(connectionString string) GoalConfiguration {
	return func(os *GoalService) error {
		d, err := checkMatchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.checkMatchesMysql = d
		return nil
	}
}

// WithRedisRepository : instantiates redis connections
func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) GoalConfiguration {
	return func(os *GoalService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// WithSlowRedisRepository : instantiates redis connections
func WithSlowRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) GoalConfiguration {
	return func(os *GoalService) error {
		d, err := rExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.slowRedisConn = d
		return nil
	}
}

// WithMysqlGoalsRepository : instantiates mysql to connect to goals interface
func WithMysqlGoalsRepository(connectionString string) GoalConfiguration {
	return func(os *GoalService) error {
		d, err := goalsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.goalMysql = d
		return nil
	}
}

// RandomIndexes : generate random numbers
func (s *GoalService) RandomIndexes(ctx context.Context) map[int]int {
	min := 1
	max := 900

	m := make(map[int]int)
	for x := 0; x < 50; x++ {
		rand.Seed(time.Now().UnixNano())
		val := rand.Intn(max-min+1) + min
		m[val] = val
	}

	return m
}

// SelectKeys :
func (s *GoalService) SelectKeys(ctx context.Context, oddsSortedSet, sanitizedKeysSet string, matches []string, projectID string) error {

	tgOver25 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "TGO25")
	tgUnder25 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "TGU25")
	tgUnder15 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "TGU15")

	tg0 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "0")
	tg1 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "1")

	tg2 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "2")
	// tg2gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "2", "gg")
	// tg2ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "2", "ng")

	tg3 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "3")
	// tg3gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "3", "gg")
	// tg3ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "3", "ng")

	tg4 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "4")
	// tg4gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "4", "gg")
	// tg4ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "4", "ng")

	tg5 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "5")
	// tg5gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "5", "gg")
	// tg5ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "5", "ng")

	tg6 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "6")
	/* 	tg6gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "6", "gg")
	   	tg6ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "6", "ng") */

	sanitizedSetLen, err := s.redisConn.SortedSetLen(ctx, sanitizedKeysSet)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, sanitizedKeysSet)
	}
	log.Printf("sanitizedSetLen :: %d", sanitizedSetLen)

	tg0Len, err := s.redisConn.SortedSetLen(ctx, tg0)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg0)
	}

	tg1Len, err := s.redisConn.SortedSetLen(ctx, tg1)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg1)
	}

	tg2Len, err := s.redisConn.SortedSetLen(ctx, tg2)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg2)
	}

	tg3Len, err := s.redisConn.SortedSetLen(ctx, tg3)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg3)
	}

	tg4Len, err := s.redisConn.SortedSetLen(ctx, tg4)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg4)
	}

	tg5Len, err := s.redisConn.SortedSetLen(ctx, tg5)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg5)
	}

	tg6Len, err := s.redisConn.SortedSetLen(ctx, tg6)
	if err != nil {
		return fmt.Errorf("err : %v failed to get zcard for key %s ", err, tg6)
	}

	if tg0Len > 100000 &&
		tg1Len > 100000 &&
		tg2Len > 100000 &&
		tg3Len > 50000 &&
		tg4Len > 30000 &&
		tg5Len > 20000 &&
		tg6Len > 20000 {
		return fmt.Errorf("there are enough games sanitized tg0Len [%d] | tg1Len [%d] |tg2Len [%d] |tg3Len [%d] |tg4Len [%d] |tg5Len [%d] |tg6Len [%d] | ",
			tg0Len, tg1Len, tg2Len, tg3Len, tg4Len, tg5Len, tg6Len)
	}

	log.Printf("oddsSortedSet len is %d", len(matches))

	if len(matches) < 1000 {
		return fmt.Errorf("there are not enough games [%d] from %s set to generate sanitized list ", len(matches), oddsSortedSet)
	}

	gamesMap := make(map[int]string)

	for i, v := range matches {
		gamesMap[i] = v
	}

	selLen := len(matches)
	selectedGames := s.NewRandomIndexes(ctx, selLen)
	log.Println("selectedGames ", selectedGames)

	// Check if this games can be saved.

	for i, v := range selectedGames {

		// get matches from the selected batch
		matchID := gamesMap[v]
		log.Printf("Selected match is : %s", matchID)

		parentID := strings.Split(matchID, "O:") // example tzO:31475633 or keO:31475634
		if len(parentID) == 2 {

			// Check if winning out is set for this game. if all is good save otherwise skip and go to the next
			// This will avoid the pending games as winning outcome is not returned.

			matchWoKey := fmt.Sprintf("%s%s%s", parentID[0], "Wo:", parentID[1])

			selectedWo, err := s.redisConn.Get(ctx, matchWoKey)
			if err != nil {
				log.Printf("Err : %v unable to get wo %s from redis ", err, selectedWo)
			} else {

				var wo oddsFiles.RawWinningOutcomes
				err = json.Unmarshal([]byte(selectedWo), &wo)
				if err != nil {
					log.Printf("Err : %v unable to un marshall winning outcome ", err)
				} else {

					// Proceed with normal processes.

					log.Printf("RoundNumberID : %s, HomeScore : %s, AwayScore : %s", wo.RoundNumberID, wo.HomeScore, wo.AwayScore)

					selWo := []string{}
					for _, i := range wo.RawWOs {
						log.Printf("SubTypeID : %s, OutcomeID : %s, ParentMatchID : %s", i.SubTypeID, i.OutcomeID, i.ParentMatchID)
						selWo = append(selWo, i.Result)
					}

					if len(selWo) > 25 {

						selCountry := parentID[0]
						//selMatchID := parentID[1]

						hScore, err := strconv.Atoi(wo.HomeScore)
						if err != nil {
							log.Printf("Err %v on converting home score `%s from string to int ", err, wo.HomeScore)
						} else {

							aScore, err := strconv.Atoi(wo.AwayScore)
							if err != nil {
								log.Printf("Err %v on converting away score `%s from string to int ", err, wo.AwayScore)
							} else {

								priority := fmt.Sprintf("%d", i)
								err := s.redisConn.ZAdd(ctx, sanitizedKeysSet, priority, matchID)
								if err != nil {
									log.Printf("Err : %v unable to add %s into %s set ", err, matchID, oddsSortedSet)
								}

								totalGoals := hScore + aScore

								log.Printf("Total goals %d", totalGoals)

								// To control over under 2.5 market
								if totalGoals > 2 {

									err := s.redisConn.ZAdd(ctx, tgOver25, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, tgOver25)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, tgOver25)
									if err != nil {
										log.Printf("err : %v ", err)
									}

								} else {

									err := s.redisConn.ZAdd(ctx, tgUnder25, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, tgUnder25)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, tgUnder25)
									if err != nil {
										log.Printf("err : %v ", err)
									}

								}

								// For future iteration

								if totalGoals == 0 {
									// Save total 0

									// Can only be ng under (goal goal market)

									s0 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "0")
									err := s.redisConn.ZAdd(ctx, s0, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s0)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s0)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									err = s.redisConn.ZAdd(ctx, tgUnder15, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, tgUnder15)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, tgUnder15)
									if err != nil {
										log.Printf("err : %v ", err)
									}

								} else if totalGoals == 1 {

									s1 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "1")
									err := s.redisConn.ZAdd(ctx, s1, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s1)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s1)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									err = s.redisConn.ZAdd(ctx, tgUnder15, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, tgUnder15)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, tgUnder15)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									// Categorise gg and ng (goal goal // no goal)
									// total goal 1 cant be categoried, can only be ng

									if hScore == 1 {

										s1h := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "1", "h")
										err := s.redisConn.ZAdd(ctx, s1h, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s1h)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s1h)
										if err != nil {
											log.Printf("err : %v ", err)
										}

									} else {

										s1a := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "1", "a")
										err := s.redisConn.ZAdd(ctx, s1a, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s1a)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s1a)
										if err != nil {
											log.Printf("err : %v ", err)
										}

									}

								} else if totalGoals == 2 {

									s2 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "2")
									err := s.redisConn.ZAdd(ctx, s2, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s2)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s2)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									// Categorise goal goal // no goal market

									if hScore > 0 && aScore > 0 {

										s2gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "2", "gg")
										err := s.redisConn.ZAdd(ctx, s2gg, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s2gg)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s2gg)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s2gg, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s2gg)
										}

									} else {

										s2ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "2", "ng")
										err := s.redisConn.ZAdd(ctx, s2ng, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s2ng)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s2ng)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s2ng, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s2ng)
										}

									}

								} else if totalGoals == 3 {

									s3 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "3")
									err := s.redisConn.ZAdd(ctx, s3, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s3)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s3)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									// Categorise goal goal // no goal market

									if hScore > 0 && aScore > 0 {

										s3gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "3", "gg")
										err := s.redisConn.ZAdd(ctx, s3gg, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s3gg)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s3gg)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s3gg, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s3gg)
										}

									} else {

										s3ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "3", "ng")
										err := s.redisConn.ZAdd(ctx, s3ng, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s3ng)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s3ng)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s3ng, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s3ng)
										}

									}

								} else if totalGoals == 4 {

									s4 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "4")
									err := s.redisConn.ZAdd(ctx, s4, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s4)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s4)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									// Categorise goal goal // no goal market

									if hScore > 0 && aScore > 0 {

										s4gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "4", "gg")
										err := s.redisConn.ZAdd(ctx, s4gg, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s4gg)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s4gg)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s4gg, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s4gg)
										}

									} else {

										s4ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "4", "ng")
										err := s.redisConn.ZAdd(ctx, s4ng, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s4ng)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s4ng)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s4ng, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s4ng)
										}

									}

								} else if totalGoals == 5 {

									s5 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "5")
									err := s.redisConn.ZAdd(ctx, s5, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s5)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s5)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									// Categorise goal goal // no goal market

									if hScore > 0 && aScore > 0 {

										s5gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "5", "gg")
										err := s.redisConn.ZAdd(ctx, s5gg, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s5gg)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s5gg)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s5gg, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s5gg)
										}

									} else {

										s5ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "5", "ng")
										err := s.redisConn.ZAdd(ctx, s5ng, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s5ng)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s5ng)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s5ng, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s5ng)
										}

									}

								} else if totalGoals == 6 {

									s6 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "6")
									err := s.redisConn.ZAdd(ctx, s6, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s6)
									}

									// save Goal category
									err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s6)
									if err != nil {
										log.Printf("err : %v ", err)
									}

									// Categorise goal goal // no goal market

									if hScore > 0 && aScore > 0 {

										s6gg := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "6", "gg")
										err := s.redisConn.ZAdd(ctx, s6gg, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s6gg)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s6gg)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s6gg, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s6gg)
										}

									} else {

										s6ng := fmt.Sprintf("%s_%s_%s", sanitizedKeysSet, "6", "ng")
										err := s.redisConn.ZAdd(ctx, s6ng, priority, matchID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s6ng)
										}

										// save Goal category
										err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, s6ng)
										if err != nil {
											log.Printf("err : %v ", err)
										}

										err = s.Save1X2(ctx, s6ng, priority, matchID, hScore, aScore, selCountry, projectID)
										if err != nil {
											log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s6ng)
										}

									}

								} else {

									s7 := fmt.Sprintf("%s_%s", sanitizedKeysSet, "7")
									err := s.redisConn.ZAdd(ctx, s7, priority, matchID)
									if err != nil {
										log.Printf("Err : %v unable to add %s into %s set ", err, matchID, s7)
									}

								}

								// Save record

								matchDate := time.Now().Format("2006-01-02")
								cm, err := checkMatches.NewCheckMatches(parentID[0], parentID[1], matchDate)
								if err != nil {
									log.Printf("Err : %v failed to instantiate checkMatches struct", err)
								}

								lastID, err := s.checkMatchesMysql.Save(ctx, *cm)
								if err != nil {
									log.Printf("Err : %v failed to save into checkMatches table", err)
								}

								log.Printf("Last inserted id %d into checkMatches tbl", lastID)

							}

						}

					} else {
						// Push this value as it is not correct

						usedKeys := fmt.Sprintf("%s_%s", oddsSortedSet, "WRONG_FORMAT")
						log.Printf("this key %s is is not formatted correctly ", usedKeys)

						priority := fmt.Sprintf("%d", i)
						err = s.redisConn.ZAdd(ctx, usedKeys, priority, matchID)
						if err != nil {
							log.Printf("Err : %v unable to add %s into %s set ", err, matchID, usedKeys)
						}
					}

				}

			}

		} else {
			log.Printf("Match saved with wrong format : %s", parentID)
		}

	}

	return nil
}

// Save1X2 :
func (s *GoalService) Save1X2(ctx context.Context, baseKey, priority, matchID string, homeScore, awayScore int, selCountry, projectID string) error {

	if homeScore > awayScore {

		hm := fmt.Sprintf("%s_%s", baseKey, "h")
		err := s.redisConn.ZAdd(ctx, hm, priority, matchID)
		if err != nil {
			return fmt.Errorf("err : %v unable to add %s into %s set ", err, matchID, hm)
		}

		// save Goal category
		err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, hm)
		if err != nil {
			log.Printf("err : %v ", err)
		}

	} else if awayScore > homeScore {

		hm := fmt.Sprintf("%s_%s", baseKey, "a")
		err := s.redisConn.ZAdd(ctx, hm, priority, matchID)
		if err != nil {
			return fmt.Errorf("err : %v unable to add %s into %s set ", err, matchID, hm)
		}

		// save Goal category
		err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, hm)
		if err != nil {
			log.Printf("err : %v ", err)
		}

	} else {

		// Draw

		hm := fmt.Sprintf("%s_%s", baseKey, "d")
		err := s.redisConn.ZAdd(ctx, hm, priority, matchID)
		if err != nil {
			return fmt.Errorf("err : %v unable to add %s into %s set ", err, matchID, hm)
		}

		// save Goal category
		err = s.SaveGoalCategory(ctx, matchID, selCountry, projectID, hm)
		if err != nil {
			log.Printf("err : %v ", err)
		}

	}

	return nil
}

// NewRandomIndexes : used to create new randomization.
func (s *GoalService) NewRandomIndexes(ctx context.Context, max int) map[int]int {
	min := 1

	m := make(map[int]int)
	for x := 0; x < 50; x++ {
		rand.Seed(time.Now().UnixNano())
		val := rand.Intn(max-min+1) + min
		m[val] = val
	}

	return m
}

func (s *GoalService) ReturnOdds(ctx context.Context, oddsSortedSet string) ([]string, error) {

	matches, err := s.slowRedisConn.GetZRange(ctx, oddsSortedSet)
	if err != nil {
		return []string{}, fmt.Errorf("err : %v on querying zrange for key : %s", err, oddsSortedSet)
	}

	return matches, nil

}

func (s *GoalService) SaveGoalCategory(ctx context.Context, matchID, country, projectID, category string) error {

	goals, err := goals.NewGoals(matchID, country, projectID, category)
	if err != nil {
		return fmt.Errorf("err : %v initiating goals: ", err)
	}

	id, err := s.goalMysql.Save(ctx, *goals)
	if err != nil {
		return fmt.Errorf("err : %v saving new goal ", err)
	}
	log.Printf("Created id %d", id)

	return nil

}
