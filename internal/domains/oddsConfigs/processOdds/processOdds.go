package processOdds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsConfigs"
	"github.com/lukemakhanu/magic_carpet/internal/domains/oddsFiles"
)

var _ oddsConfigs.OddsConfigsRepository = (*OddsConfigs)(nil)

type OddsConfigs struct {
	oddsPayload string
	woPayload   string
	lsPayload   string
	oddsFactor  float64
}

// New initializes a new instance of odds.
func New(oddsPayload, woPayload, lsPayload string, oddsFactor float64) (*OddsConfigs, error) {

	if oddsPayload == "" {
		return nil, fmt.Errorf("oddsPayload not set")
	}

	if woPayload == "" {
		return nil, fmt.Errorf("woPayload not set")
	}

	if lsPayload == "" {
		log.Printf("Scores must be zero")
		//return nil, fmt.Errorf("lsPayload not set")
	}

	if oddsFactor <= 0.00 {
		return nil, fmt.Errorf("oddsFactor not set")
	}

	c := &OddsConfigs{
		oddsPayload: oddsPayload,
		woPayload:   woPayload,
		lsPayload:   lsPayload,
		oddsFactor:  oddsFactor,
	}

	return c, nil
}

// FormulateOdds : rewrites odds the right way
func (s *OddsConfigs) FormulateOdds(ctx context.Context) ([]oddsFiles.FinalMarkets, oddsFiles.FinalScores, []oddsFiles.FinalLiveScores, error) {

	mkts := []oddsFiles.FinalMarkets{}
	fs := oddsFiles.FinalScores{}
	lsc := []oddsFiles.FinalLiveScores{}

	var o oddsFiles.RawOdds
	err := json.Unmarshal([]byte(s.oddsPayload), &o)
	if err != nil {
		log.Printf("Err on markets : %v", err)
	}

	//log.Printf("ProjectID:%s, parentMatchID:%s, matchID:%s",
	//	o.ProjectID, o.ParentMatchID, o.MatchID)

	for _, x := range o.RawMarkets {
		//log.Printf("name:%s, subTypeID:%s", x.Name, x.SubTypeID)

		updatedName, status := renameMarketName(x.SubTypeID)

		if status == "1" {

			mkt := oddsFiles.FinalMarkets{
				Name: updatedName,
				Code: x.SubTypeID,
			}

			for _, i := range x.RawOutcomes {
				//log.Printf("outcomeID: %s, outcomeName: %s, outcomeAlias: %s, oddValue: %s",
				//	i.OutcomeID, i.OutcomeName, i.OutcomeAlias, i.OddValue)

				oddVl, err := strconv.ParseFloat(i.OddValue, 64)
				if err != nil {
					log.Printf("Err : %v failed to convert string to float64", err)
				} else {

					savedOdds := oddVl - s.oddsFactor
					finalOdd := math.Floor(savedOdds*100) / 100

					outcome := oddsFiles.FinalOutcomes{
						OutcomeID:    i.OutcomeID,
						OutcomeName:  i.OutcomeName,
						OddValue:     finalOdd, //savedOdds,
						OutcomeAlias: i.OutcomeAlias,
					}

					mkt.FinalOutcomes = append(mkt.FinalOutcomes, outcome)

				}
			}

			mkts = append(mkts, mkt)

		} else {
			//log.Printf("Skipped market : %s", x.SubTypeID)
		}

	}

	// Process Winning Outcomes

	var wo oddsFiles.RawWinningOutcomes
	err = json.Unmarshal([]byte(s.woPayload), &wo)
	if err != nil {
		log.Printf("Err on winning outcome : %v", err)
	}

	fs.HomeScore = wo.HomeScore
	fs.AwayScore = wo.AwayScore

	for _, w := range wo.RawWOs {

		_, status := renameMarketName(w.SubTypeID)

		if status == "1" {

			fWo := oddsFiles.FinalWinningOutcomes{
				SubTypeID:   w.SubTypeID,
				OutcomeID:   w.OutcomeID,
				OutcomeName: w.OutcomeName,
				Result:      w.Result,
			}

			fs.FinalWinningOutcomes = append(fs.FinalWinningOutcomes, fWo)

		} else {
			//log.Printf("Skipped market for winning outcome : %s", w.SubTypeID)
		}

	}

	// Process Live scores.

	var ls oddsFiles.RawLS
	err = json.Unmarshal([]byte(s.lsPayload), &ls)
	if err != nil {
		log.Printf("Err on live score : %v", err)
	}

	for _, x := range ls.Goals {

		log.Printf("x.AwayScore : %d, x.HomeScore : %d, x.MinuteScored : %s", x.AwayScore, x.HomeScore, x.MinuteScored)

		l := oddsFiles.FinalLiveScores{
			HomeScore:    x.HomeScore,
			AwayScore:    x.AwayScore,
			MinuteScored: x.MinuteScored,
		}

		lsc = append(lsc, l)

	}

	return mkts, fs, lsc, nil
}

// FormulateOdds2 : rewrites odds the right way
func (s *OddsConfigs) FormulateOdds2(ctx context.Context) ([]oddsFiles.FinalMarkets, oddsFiles.FinalScores, []oddsFiles.FinalLiveScores, error) {

	mkts := []oddsFiles.FinalMarkets{}
	fs := oddsFiles.FinalScores{}
	lsc := []oddsFiles.FinalLiveScores{}

	var o oddsFiles.RawOdds
	err := json.Unmarshal([]byte(s.oddsPayload), &o)
	if err != nil {
		log.Printf("Err on markets : %v", err)
	}

	//log.Printf("ProjectID:%s, parentMatchID:%s, matchID:%s",
	//	o.ProjectID, o.ParentMatchID, o.MatchID)

	for _, x := range o.RawMarkets {
		//log.Printf("name:%s, subTypeID:%s", x.Name, x.SubTypeID)

		updatedName, status := renameMarketName(x.SubTypeID)

		if status == "1" {

			mkt := oddsFiles.FinalMarkets{
				Name: updatedName,
				Code: x.SubTypeID,
			}

			rr := 1
			for _, i := range x.RawOutcomes {
				//log.Printf("outcomeID: %s, outcomeName: %s, outcomeAlias: %s, oddValue: %s",
				//	i.OutcomeID, i.OutcomeName, i.OutcomeAlias, i.OddValue)

				oddVl, err := strconv.ParseFloat(i.OddValue, 64)
				if err != nil {
					log.Printf("Err : %v failed to convert string to float64", err)
				} else {

					goals := s.OddsFactor(ctx)
					max := len(goals)
					selectedRations := rand.IntN(max)

					//selectedRations := s.NewRandomIndexes(ctx, max)

					selectedOdd := goals[selectedRations] //goals[selectedRations[5]]

					savedOdds := oddVl - selectedOdd //s.oddsFactor
					finalOdd := math.Floor(savedOdds*100) / 100
					log.Printf("*** oddVl *** %f | selectedOdd *** %f | savedOdds %f | finalOdd %f",
						oddVl, selectedOdd, savedOdds, finalOdd)

					outcome := oddsFiles.FinalOutcomes{
						OutcomeID:    i.OutcomeID,
						OutcomeName:  i.OutcomeName,
						OddValue:     finalOdd, //savedOdds,
						OutcomeAlias: i.OutcomeAlias,
					}

					mkt.FinalOutcomes = append(mkt.FinalOutcomes, outcome)

				}

				if rr > 7 {
					rr = 1
				} else {
					rr++
				}
			}

			mkts = append(mkts, mkt)

		} else {
			//log.Printf("Skipped market : %s", x.SubTypeID)
		}

	}

	// Process Winning Outcomes

	var wo oddsFiles.RawWinningOutcomes
	err = json.Unmarshal([]byte(s.woPayload), &wo)
	if err != nil {
		log.Printf("Err on winning outcome : %v", err)
	}

	fs.HomeScore = wo.HomeScore
	fs.AwayScore = wo.AwayScore

	for _, w := range wo.RawWOs {

		_, status := renameMarketName(w.SubTypeID)

		if status == "1" {

			fWo := oddsFiles.FinalWinningOutcomes{
				SubTypeID:   w.SubTypeID,
				OutcomeID:   w.OutcomeID,
				OutcomeName: w.OutcomeName,
				Result:      w.Result,
			}

			fs.FinalWinningOutcomes = append(fs.FinalWinningOutcomes, fWo)

		} else {
			//log.Printf("Skipped market for winning outcome : %s", w.SubTypeID)
		}

	}

	// Process Live scores.

	var ls oddsFiles.RawLS
	err = json.Unmarshal([]byte(s.lsPayload), &ls)
	if err != nil {
		log.Printf("Err on live score : %v", err)
	}

	for _, x := range ls.Goals {

		log.Printf("x.AwayScore : %d, x.HomeScore : %d, x.MinuteScored : %s", x.AwayScore, x.HomeScore, x.MinuteScored)

		l := oddsFiles.FinalLiveScores{
			HomeScore:    x.HomeScore,
			AwayScore:    x.AwayScore,
			MinuteScored: x.MinuteScored,
		}

		lsc = append(lsc, l)

	}

	return mkts, fs, lsc, nil
}

// NewRandomIndexes : used to create new randomization.
func (s *OddsConfigs) NewRandomIndexes(ctx context.Context, max int) map[int]int {
	//min := 1

	m := make(map[int]int)
	for x := 0; x < 10; x++ {
		// rand.Seed(time.Now().UnixNano())
		// val := rand.Intn(max-min+1) + min

		val := rand.IntN(max)
		m[val] = val
	}

	return m
}

// TotalGoalsPerSession : this helps with distribution of total goals per match session.
func (s *OddsConfigs) OddsFactor(ctx context.Context) []float64 {

	data := []float64{
		0.0119, 0.0117, 0.011, 0.0114, 0.0118, 0.0117, 0.0114,
		0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01,
		0.012, 0.012, 0.012, 0.012, 0.012, 0.012, 0.012,
		0.02, 0.0204, 0.0208, 0.0209, 0.02, 0.02,
		0.021, 0.021, 0.021, 0.021,
		-0.001, -0.001, -0.001,
	}

	return data
}

func renameMarketName(subTypeID string) (string, string) {

	if subTypeID == "CS" {
		return "Correct Score (FT)", "1"
	} else if subTypeID == "HS" {
		return "Half Time Score", "1"
	} else if subTypeID == "1X2" {
		return "Match Result", "1"
	} else if subTypeID == "H1X2" {
		return "Half Time Result", "1"
	} else if subTypeID == "DC" {
		return "Double Chance", "1"
	} else if subTypeID == "DCH" {
		return "Double Chance (HT)", "1"
	} else if subTypeID == "TG15" {
		return "Over/Under 1.5", "1"
	} else if subTypeID == "TG25" {
		return "Over/Under 2.5", "1"
	} else if subTypeID == "TG35" {
		return "Over/Under 3.5", "1"
	} else if subTypeID == "HX1" {
		return "Handicap -1", "1"
	} else if subTypeID == "HX2" {
		return "Handicap -2", "1"
	} else if subTypeID == "DR" {
		return "Half Time / Full Time", "1"
	} else if subTypeID == "TG" {
		return "Total Goals", "1"
	} else if subTypeID == "GG" {
		return "Goal:Goal FT", "1"
	} else if subTypeID == "HGG" {
		return "Goal:Goal HT", "1"
	} else if subTypeID == "1X2OU15" {
		return "1X2 and Over/Under 1.5", "1"
	} else if subTypeID == "1X2OU25" {
		return "1X2 and Over/Under 2.5", "1"
	} else if subTypeID == "1X2OU35" {
		return "1X2 and Over/Under 3.5", "1"
	} else if subTypeID == "1X2OU45" {
		return "1X2 and Over/Under 4.5", "1"
	} else if subTypeID == "1X2OU55" {
		return "1X2 and Over/Under 5.5", "1"
	} else if subTypeID == "1X2G" {
		return "1X2 and Goal/No Goal", "1"
	} else if subTypeID == "T1OU15" {
		return "Team 1 Over/Under 1.5", "1"
	} else if subTypeID == "T2OU15" {
		return "Team 2 Over/Under 1.5", "1"
	} else if subTypeID == "T1G" {
		return "Team 1 Goal/No Goal", "1"
	} else if subTypeID == "T2G" {
		return "Team 2 Goal/No Goal", "1"
	} else if subTypeID == "TGOE" {
		return "Total Goals Odd/Even", "1"
	} else if subTypeID == "TFG" {
		return "Time of First Goal", "1"
	} else if subTypeID == "FTS" {
		return "First Team to Score", "1"
	} else if subTypeID == "MG" {
		return "Multi-Goals", "1"
	} else {
		return "0", "0"
	}

}
