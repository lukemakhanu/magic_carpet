package results

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/lukemakhanu/veimu_apps/internal/domain/resultStatuses"
)

// Compile time interface assertion.
var _ ResultFetcher = (*ResultClient)(nil)

// ResultClient contains methods for interacting with Kovacic APIs.
type ResultClient struct {
	resultEndPoint string
	roundNumberID  string
	competitionID  string
	startTime      string
	timeouts       time.Duration
	client         *http.Client
	loc            *time.Location
}

// New initializes a new instance of Live Score Client.
func New(resultEndPoint, roundNumberID, competitionID, startTime string, timeouts time.Duration, client *http.Client, loc *time.Location) (*ResultClient, error) {

	resultURL, err := url.Parse(resultEndPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result endpoint: %w", err)
	}

	if roundNumberID == "" {
		return nil, fmt.Errorf("Round number ID not set")
	}

	if competitionID == "" {
		return nil, fmt.Errorf("competitionID not set")
	}

	if startTime == "" {
		return nil, fmt.Errorf("startTime not set")
	}

	if timeouts <= 0 {
		return nil, fmt.Errorf("Timeout not set")
	}

	c := &ResultClient{
		resultEndPoint: resultURL.String(),
		roundNumberID:  roundNumberID,
		competitionID:  competitionID,
		startTime:      startTime,
		timeouts:       timeouts,
		client:         client,
		loc:            loc,
	}
	if c.client == nil {
		c.client = defaultHTTPClient
	}

	return c, nil
}

// GetResults : returns results from third party provider
func (s *ResultClient) GetResults(ctx context.Context) ([]resultStatuses.Results, []resultStatuses.MatchResults, error) {

	var g []resultStatuses.Results
	var m []resultStatuses.MatchResults

	var resultUrl = fmt.Sprintf("%s%s", s.resultEndPoint, s.roundNumberID)
	log.Printf("Calling... %s | start time : [ %s ]", resultUrl, s.startTime)

	response, err := s.client.Get(resultUrl)
	if err != nil {
		return g, m, fmt.Errorf("Failed to call result API")
	}

	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return g, m, fmt.Errorf("failed to read response body: %w | responseBody : %s ", err, string(responseBody))
	}

	if response.StatusCode == http.StatusOK {
		//
		// Start Parsing data.
		cc := resultStatuses.ResultStatuses{
			StartTime:     s.startTime,
			CompetitionID: s.competitionID,
			RoundNumberID: s.roundNumberID,
			Status:        "pending",
		}
		return parseResults(responseBody, cc, s.loc)
	}

	return g, m, fmt.Errorf("failed to get results : status: %d, error body: %s", response.StatusCode, responseBody)

}

func parseResults(content []byte, c resultStatuses.ResultStatuses, loc *time.Location) ([]resultStatuses.Results, []resultStatuses.MatchResults, error) {

	var g []resultStatuses.Results
	var m []resultStatuses.MatchResults

	var r resultStatuses.WoJson
	err := json.Unmarshal(content, &r)
	if err != nil {
		return g, m, fmt.Errorf("failed to unmarshal xml : %w", err)
	}

	log.Printf("statusCode : %s | statusDesc : %s | matchDay %s | seasonID %s | seasonWeekID  %s | startTime %s",
		r.StatusCode, r.StatusDescription, r.OutcomeDetailsJson.MatchDay, r.OutcomeDetailsJson.SeasonID,
		r.OutcomeDetailsJson.SeasonWeekID, r.OutcomeDetailsJson.StartTime)

	for _, d := range r.OutcomeDetailsJson.MatchesJson {

		log.Printf("matchID %s | homeTeam %s | awayTeam %s | HomeScore : %s | AwayScore %s",
			d.MatchID, d.HomeTeam, d.AwayTeam, d.FinalScoreJson.HomeScore, d.FinalScoreJson.AwayScore)

		for _, o := range d.FinalScoreJson.OutcomesJson {

			log.Printf("subTypeID %s | outcomeID %s | outcomeName %s | result %s",
				o.SubTypeID, o.OutcomeID, o.OutcomeName, o.Result)

			dd := resultStatuses.Results{
				RoundNumberID: r.OutcomeDetailsJson.SeasonWeekID,
				ParentMatchID: d.MatchID,   // parent_match_id
				SubTypeID:     o.SubTypeID, // sub_type_id
				StartTime:     r.OutcomeDetailsJson.StartTime,
				OutcomeID:     o.OutcomeID,
				CompetitionID: c.CompetitionID,
				OutcomeName:   o.OutcomeName, //mrk.Description,
				Result:        o.Result,
				VoidFactor:    "1",
				Producer:      "RESULT-HTTP-CLIENT",
				MatchType:     "matches",
			}

			g = append(g, dd)

			/*d := resultStatuses.MatchResults{
				ParentMatchID: d.MatchID,
				HomeTeamID:    d.HomeID,
				AwayTeamID:    d.AwayID,
				StartTime:     r.OutcomeDetailsJson.StartTime, //  starTime,
				EventTime:     r.OutcomeDetailsJson.StartTime, //  mkt.EventTime,
				RoundNumberID: c.RoundNumberID,
				CompetitionID: c.CompetitionID,
				HomeScore:     d.FinalScoreJson.HomeScore,
				AwayScore:     d.FinalScoreJson.AwayScore,
			}

			m = append(m, d) */

		}

		d := resultStatuses.MatchResults{
			ParentMatchID: d.MatchID,
			HomeTeamID:    d.HomeID,
			AwayTeamID:    d.AwayID,
			StartTime:     r.OutcomeDetailsJson.StartTime, //  starTime,
			EventTime:     r.OutcomeDetailsJson.StartTime, //  mkt.EventTime,
			RoundNumberID: c.RoundNumberID,
			CompetitionID: c.CompetitionID,
			HomeScore:     d.FinalScoreJson.HomeScore,
			AwayScore:     d.FinalScoreJson.AwayScore,
		}

		m = append(m, d)

	}

	if len(g) == 0 || len(m) == 0 {
		return g, m, fmt.Errorf("winning outcome not found")
	}

	return g, m, nil
}

var defaultHTTPClient = &http.Client{
	Timeout: time.Second * 15,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Second * 15,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}
