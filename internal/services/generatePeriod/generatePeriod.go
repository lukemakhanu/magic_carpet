package generatePeriod

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns"
	"github.com/lukemakhanu/magic_carpet/internal/domains/goalPatterns/goalPatternsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domains/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domains/scheduledTimes"
	scheduledTimeMysql "github.com/lukemakhanu/magic_carpet/internal/domains/scheduledTimes/scheduledTimesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/snwkpts"
	"github.com/lukemakhanu/magic_carpet/internal/domains/snwkpts/snwkptsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/ssns"
	"github.com/lukemakhanu/magic_carpet/internal/domains/ssns/ssnsMysql"
)

// GeneratePeriodConfiguration is an alias for a function that will take in a pointer to an GeneratePeriodService and modify it
type GeneratePeriodConfiguration func(os *GeneratePeriodService) error

type GeneratePeriodService struct {
	ssnsMysql          ssns.SsnsRepository
	scheduledTimeMysql scheduledTimes.ScheduledTimesRepository
	redisConn          processRedis.RunRedis
	goalPatternsMysql  goalPatterns.GoalPatternsRepository
	snwkptsMysql       snwkpts.SnWkPtsRepository
}

func NewGeneratePeriodService(cfgs ...GeneratePeriodConfiguration) (*GeneratePeriodService, error) {
	os := &GeneratePeriodService{}
	for _, cfg := range cfgs {
		err := cfg(os)
		if err != nil {
			return nil, err
		}
	}
	return os, nil
}

func WithMysqlSsnsRepository(connectionString string) GeneratePeriodConfiguration {
	return func(os *GeneratePeriodService) error {
		d, err := ssnsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.ssnsMysql = d
		return nil
	}
}

func WithMysqlScheduledTimeRepository(connectionString string) GeneratePeriodConfiguration {
	return func(os *GeneratePeriodService) error {
		d, err := scheduledTimeMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.scheduledTimeMysql = d
		return nil
	}
}

func WithRedisRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) GeneratePeriodConfiguration {
	return func(os *GeneratePeriodService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

func WithMysqlGoalPatternsRepository(connectionString string) GeneratePeriodConfiguration {
	return func(os *GeneratePeriodService) error {
		d, err := goalPatternsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.goalPatternsMysql = d
		return nil
	}
}

func WithMysqlSnWkPtsRepository(connectionString string) GeneratePeriodConfiguration {
	return func(os *GeneratePeriodService) error {
		d, err := snwkptsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.snwkptsMysql = d
		return nil
	}
}

// CreateScheduledTime : used to create scheduled start time for each date
func (s *GeneratePeriodService) CreateScheduledTime(ctx context.Context, locale *time.Location, competitionID, status string, addTime int64) error {

	count, err := s.scheduledTimeMysql.CountScheduledTime(ctx, competitionID)
	if err != nil {
		return fmt.Errorf("Err :: %v ", err)
	}

	if count < 5 {

		log.Printf("About to create new scheduled time, remaining periods ::: %d", count)

		// Create scheduled times for a few days to come.

		now := time.Now()
		fmt.Println("Now: " + now.String())

		for i := 0; i < 10; i++ {

			scheduled := time.Date(now.Year(), now.Month(), now.Day(), 03, 0, 0, 0, time.UTC).AddDate(0, 0, i)

			//var schTime = scheduled.Add(time.Minute * time.Duration(addTime)).In(locale).Format("2006-01-02 15:04:00")
			var schTime = scheduled.Add(time.Minute * time.Duration(addTime)).Format("2006-01-02 15:04:00")

			status := "inactive"
			sc, err := scheduledTimes.NewScheduledTime(schTime, competitionID, status)
			if err != nil {
				log.Printf("Err : %v unable to initiate scheduled time", err)
			}

			lastID, err := s.scheduledTimeMysql.Save(ctx, *sc)
			if err != nil {
				log.Printf("Err : %v unable to save scheduled time", err)
			}

			log.Printf("Last scheduled time saved : %d", lastID)

		}

	}

	log.Printf("Still scheduled times : %d", count)

	return nil

}

// CreateScheduledTime : used to create scheduled start time for each date
func (s *GeneratePeriodService) PrepareGames(ctx context.Context, locale *time.Location, competitionID, status string) error {

	count, err := s.ssnsMysql.CountRemainingPeriods(ctx, competitionID)
	if err != nil {
		return fmt.Errorf("err :: %v ", err)
	}

	x := 0

	if count < 20 {

		log.Printf("About to initiate a process to create new matches, remaining periods : %d", count)

		// Select the top scheduled time that is still inactive. This record must be updated to active after its periods
		// are created.

		schID, scheduledTime, err := s.scheduledTimeMysql.FirstActiveScheduledTime(ctx, competitionID, status)
		if err != nil {
			log.Printf("%v", err)
		} else {
			log.Printf("ScheduledTimeID %s | scheduledTime %s", schID, scheduledTime)

			// Proceed to create periods

			matchStartTime := 120

			gTime, err := time.Parse("2006-01-02 15:04:05", scheduledTime)
			if err != nil {
				return fmt.Errorf("err : %v on converting string to time..", err)
			}

			ssn, err := ssns.NewSsns(competitionID, status)
			if err != nil {
				return fmt.Errorf("err : %v failed to instantiate a new season", err)
			}

			ssnID, err := s.ssnsMysql.Save(ctx, *ssn)
			if err != nil {
				return fmt.Errorf("err : %v failed to save a new ssns", err)
			}

			if ssnID > 0 {

				for i := 1; i <= 19; i++ {

					log.Printf("Proceed to saving the rest of the data..")

					// For each season create 38 season weeks.

					// At this point select the patterns to use for this games
					// We get 38 patterns, put them in a map and use them below
					// to decide how goals will be destributed.

					goalDistribution, err := s.goalPatternsMysql.GoalDistributions(ctx, competitionID)
					if err != nil {
						return fmt.Errorf("err : %v failed to return goal distribution", err)
					}

					if len(goalDistribution) == 0 {
						return fmt.Errorf("err : %v no goal distribution returned %d", err, len(goalDistribution))
					}

					availableList := len(goalDistribution)
					selCategory := rand.IntN(availableList)

					selectedBatch := goalDistribution[selCategory]

					log.Println("selCategory", selCategory, "selectedBatch >>>> ", selectedBatch)

					parentIDs := strings.Split(selectedBatch.RoundNumberID, ",")

					if len(goalDistribution) != 38 {
						return fmt.Errorf("err : %v there are no enough goal distribution | available %d", err, len(goalDistribution))
					}

					if len(parentIDs) != 38 {
						return fmt.Errorf("err : %v there round number ids returned arent enough %d", err, len(parentIDs))
					}

					distr := make(map[int]string)
					kk := 1
					for _, f := range parentIDs {
						distr[kk] = f
						kk++
					}

					for h := 1; h <= 38; h++ {

						if x < 715 {

							// Instantiate season week for all upcoming games.

							var startTime = gTime.Add(time.Second * time.Duration(matchStartTime)).Format("2006-01-02 15:04:05")

							enTime, err := time.Parse("2006-01-02 15:04:05", startTime)
							if err != nil {
								return fmt.Errorf("err getting minute scored : %v", err)
							}

							matchDuration := 35
							var endTime = enTime.Add(time.Second * time.Duration(matchDuration)).Format("2006-01-02 15:04:05")

							seasonID := fmt.Sprintf("%d", ssnID)
							weekNumber := fmt.Sprintf("%d", h)
							status := "inactive"

							swID, err := s.ssnsMysql.SaveSsnWeek(ctx, competitionID, seasonID, weekNumber, status, startTime, endTime) //seasonWeeks.NewSeasonWeeks(seasonID, weekNumber, status, startTime, endTime)
							if err != nil {
								return fmt.Errorf("err : %v failed to save season week.", err)
							}

							ssnWkID := fmt.Sprintf("%d", swID)

							// Add goal distribution

							rnID := distr[h]
							gg, err := snwkpts.NewSnWkPts(ssnWkID, rnID, competitionID)
							if err != nil {
								return fmt.Errorf("err : %v failed to instantiate snWkPts", err)
							}

							snwkptID, err := s.snwkptsMysql.Save(ctx, *gg)
							if err != nil {
								return fmt.Errorf("err : %v failed to save snWkPts", err)
							}

							log.Printf("last snwkptID %d", snwkptID)

							// Get this data from redis (ssn weeks)

							zkey := fmt.Sprintf("%s_%d_%s", competitionID, i, weekNumber)

							data, err := s.redisConn.GetZRangeWithLimit(ctx, zkey, 12)
							if err != nil {
								log.Printf("Err : %v failed to read from %s z range **", err, zkey)
							} else {

								for _, k := range data {

									tvt := strings.Split(k, ",")
									if len(tvt) == 2 {

										homeTeamID := tvt[0]
										AwayTeamID := tvt[1]
										matchStatus := "inactive"

										matchID, err := s.ssnsMysql.SaveGames(ctx, competitionID, seasonID, ssnWkID, weekNumber, homeTeamID, AwayTeamID, matchStatus) // matchesMysql.Save(ctx, *mt)
										if err != nil {
											return fmt.Errorf("Err : %v failed to save match", err)
										}

										log.Printf("Last match id : %d", matchID)

									} else {
										return fmt.Errorf("Wrong format for teams facing each other %s | weekNumber %s", k, weekNumber)
									}

								}

								matchStartTime += 120

							}

						}

						x++

					}

				}

			} else {
				log.Printf("ssnID return 0 --> %d", ssnID)
			}

			// Update used scheduled Time

			updatedStatus := "active"
			updated, err := s.scheduledTimeMysql.UpdateScheduleTime(ctx, updatedStatus, schID)
			if err != nil {
				log.Fatalf("Err : %v unable to updated used scheduled time", err)
			}

			log.Printf("Updated record : %d", updated)

		}
	}

	log.Printf("Still enough active games : %d", count)
	return nil

}

func (s *GeneratePeriodService) PrepareGamesOriginal(ctx context.Context, locale *time.Location, competitionID, status string, addTime int64) error {

	check, selTime, err := s.CheckTime(ctx, competitionID)
	if err != nil {
		return fmt.Errorf("Err :: %v ", err)
	}

	if check == true {

		matchStartTime := 120

		var schTime = selTime.Add(time.Minute * time.Duration(addTime)).Format("2006-01-02 15:04:00") //15:04:05")

		gTime, err := time.Parse("2006-01-02 15:04:05", schTime)
		if err != nil {
			return fmt.Errorf("Err : %v on converting string to time..", err)
		}

		ssn, err := ssns.NewSsns(competitionID, status)
		if err != nil {
			return fmt.Errorf("Err : %v failed to instantiate a new season", err)
		}

		ssnID, err := s.ssnsMysql.Save(ctx, *ssn)
		if err != nil {
			return fmt.Errorf("Err : %v failed to save a new ssns", err)
		}

		if ssnID > 0 {

			for i := 1; i <= 19; i++ {

				log.Printf("Proceed to saving the rest of the data..")

				// For each season create 38 season weeks.

				for h := 1; h <= 38; h++ {

					// Instantiate season week for all upcoming games.

					var startTime = gTime.Add(time.Second * time.Duration(matchStartTime)).Format("2006-01-02 15:04:05")

					enTime, err := time.Parse("2006-01-02 15:04:05", startTime)
					if err != nil {
						return fmt.Errorf("Err getting minute scored : %v", err)
					}

					matchDuration := 35 //30
					var endTime = enTime.Add(time.Second * time.Duration(matchDuration)).Format("2006-01-02 15:04:05")

					seasonID := fmt.Sprintf("%d", ssnID)
					weekNumber := fmt.Sprintf("%d", h)
					status := "inactive"

					swID, err := s.ssnsMysql.SaveSsnWeek(ctx, competitionID, seasonID, weekNumber, status, startTime, endTime) //seasonWeeks.NewSeasonWeeks(seasonID, weekNumber, status, startTime, endTime)
					if err != nil {
						return fmt.Errorf("Err : %v failed to save season week.", err)
					}

					ssnWkID := fmt.Sprintf("%d", swID)

					// Get this data from redis (ssn weeks)

					zkey := fmt.Sprintf("%s_%d_%s", competitionID, i, weekNumber)

					data, err := s.redisConn.GetZRangeWithLimit(ctx, zkey, 12)
					if err != nil {
						log.Printf("Err : %v failed to read from %s z range **", err, zkey)
					} else {

						for _, k := range data {

							tvt := strings.Split(k, ",")
							if len(tvt) == 2 {

								homeTeamID := tvt[0]
								AwayTeamID := tvt[1]
								matchStatus := "inactive"

								matchID, err := s.ssnsMysql.SaveGames(ctx, competitionID, seasonID, ssnWkID, weekNumber, homeTeamID, AwayTeamID, matchStatus) // matchesMysql.Save(ctx, *mt)
								if err != nil {
									return fmt.Errorf("Err : %v failed to save match", err)
								}

								log.Printf("Last match id : %d", matchID)

							} else {
								return fmt.Errorf("Wrong format for teams facing each other %s | weekNumber %s", k, weekNumber)
							}

						}

						matchStartTime += 120

					}

				}

			}

		} else {
			log.Printf("ssnID return 0 --> %d", ssnID)
		}

	} else {
		log.Printf("*** Time not due ****")
	}

	return nil
}

func (s *GeneratePeriodService) CheckTime(ctx context.Context, competitionID string) (bool, time.Time, error) {

	selTime := time.Now()
	lastGateTime, err := s.ssnsMysql.GetLastGame(ctx, competitionID)
	if err != nil {
		return false, selTime, fmt.Errorf("Err : %v on getting last game time", err)
	}

	if len(lastGateTime) == 0 {
		log.Printf("The first seasons being created")
		return true, selTime, nil
	}

	cTime := time.Now()
	var systemCurrentTime = cTime.Format("2006-01-02 15:04:05") //In(loc).Format("2006-01-02 15:04:05")

	currentTime, err := time.Parse("2006-01-02 15:04:05", systemCurrentTime)
	if err != nil {
		return false, selTime, fmt.Errorf("Err getting system time : %v", err)
	}

	lastGame, err := time.Parse("2006-01-02 15:04:05", lastGateTime[0].StartTime)
	if err != nil {
		return false, selTime, fmt.Errorf("Err getting system time : %v", err)
	}

	hr := 5
	var execTime = currentTime.Add(time.Hour * time.Duration(hr))

	if execTime.After(lastGame) == true {
		log.Printf("***** Time is due ******")
		return true, lastGame, nil
	}

	return false, selTime, fmt.Errorf("Err : %v time not due.. ", err)

}
