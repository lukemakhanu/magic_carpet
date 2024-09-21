package liveScore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/lukemakhanu/veimu_apps/internal/domain/liveScoreStatuses"
)

// Compile time interface assertion.
var _ LiveScoreFetcher = (*LiveScoreClient)(nil)

// LiveScoreClient contains methods for interacting with Kovacic APIs.
type LiveScoreClient struct {
	liveScoreEndPoint string
	roundNumberID     string
	competitionID     string
	seasonID          string
	prjSeason         string
	seasonWeek        string
	startTime         string
	endTime           string
	timeouts          time.Duration
	client            *http.Client
}

// New initializes a new instance of Live Score Client.
func New(liveScoreEndPoint, roundNumberID, competitionID, seasonID, prjSeason, seasonWeek, startTime, endTime string, timeouts time.Duration, client *http.Client) (*LiveScoreClient, error) {

	liveScoreURL, err := url.Parse(liveScoreEndPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse live score endpoint: %w", err)
	}

	if roundNumberID == "" {
		return nil, fmt.Errorf("Round number ID not set")
	}

	if competitionID == "" {
		return nil, fmt.Errorf("Competition ID not set")
	}

	if seasonID == "" {
		return nil, fmt.Errorf("seasonID not set")
	}

	if prjSeason == "" {
		return nil, fmt.Errorf("prjSeason not set")
	}

	if seasonWeek == "" {
		return nil, fmt.Errorf("seasonWeek not set")
	}

	if startTime == "" {
		return nil, fmt.Errorf("startTime not set")
	}

	if endTime == "" {
		return nil, fmt.Errorf("endTime not set")
	}

	if timeouts <= 0 {
		return nil, fmt.Errorf("Timeout not set")
	}

	c := &LiveScoreClient{
		liveScoreEndPoint: liveScoreURL.String(),
		roundNumberID:     roundNumberID,
		competitionID:     competitionID,
		seasonID:          seasonID,
		prjSeason:         prjSeason,
		seasonWeek:        seasonWeek,
		startTime:         startTime,
		endTime:           endTime,
		timeouts:          timeouts,
		client:            client,
	}
	if c.client == nil {
		c.client = defaultHTTPClient
	}

	return c, nil
}

// GetLiveScores : returns live scores from third party provider
func (s *LiveScoreClient) GetLiveScores(ctx context.Context) ([]liveScoreStatuses.ScoreLine, error) {

	var liveScoreUrl = fmt.Sprintf("%s%s", s.liveScoreEndPoint, s.roundNumberID)
	log.Printf("Calling... %s | start time : %s", liveScoreUrl, s.startTime)

	response, err := s.client.Get(liveScoreUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to call match API")
	}

	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w | responseBody : %s ", err, string(responseBody))
	}

	if response.StatusCode == http.StatusOK {
		//
		// Start Parsing data.
		cc := liveScoreStatuses.LiveScoreStatuses{
			StartTime:     s.startTime,
			EndTime:       s.endTime,
			RoundNumberID: s.roundNumberID,
			CompetitionID: s.competitionID,
			SeasonID:      s.seasonID,
			PrjSeason:     s.prjSeason,
			SeasonWeek:    s.seasonWeek,
			Status:        "pending",
		}
		return parseLiveScores(responseBody, cc)
	}

	return nil, fmt.Errorf("failed to get live scores : status: %d, error body: %s", response.StatusCode, responseBody)

}

func parseLiveScores(content []byte, c liveScoreStatuses.LiveScoreStatuses) ([]liveScoreStatuses.ScoreLine, error) {

	var g []liveScoreStatuses.ScoreLine

	var s liveScoreStatuses.RawLiveScores
	err := json.Unmarshal(content, &s)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json : %w", err)
	}

	log.Printf("statusCode %s | statusDescription %s | MatchDay %s | seasonID %s | SeasonWeekID %s | StartTime %s | EndTime %s",
		s.StatusCode, s.StatusDescription, s.MatchLiveScores.MatchDay, s.MatchLiveScores.SeasonID, s.MatchLiveScores.SeasonWeekID,
		s.MatchLiveScores.StartTime, s.MatchLiveScores.EndTime)

	for _, i := range s.MatchLiveScores.LsMatches {
		log.Printf("matchID %s | homeID %s awayID %s", i.MatchID, i.HomeID, i.AwayID)

		for _, x := range i.LiveScores {
			log.Printf("homeScore %d | awayScore %d | minuteScored %s",
				x.HomeScore, x.AwayScore, x.MinuteScored)

			st, err := time.Parse("2006-01-02 15:04:05", c.StartTime)
			if err != nil {
				return nil, fmt.Errorf("Could not parse start time : %w", err)
			}

			minScored, err := strconv.Atoi(x.MinuteScored)
			if err != nil {
				return nil, fmt.Errorf("Could not convert string to int : %w", err)
			}

			var addT = st.Add(time.Second * time.Duration(minScored))

			d := liveScoreStatuses.ScoreLine{
				CompetitionID: c.CompetitionID,
				HomeScore:     x.HomeScore, //hTeam,
				AwayScore:     x.AwayScore, //aTeam,
				ScoreTime:     addT,
				ParentMatchID: i.MatchID, //rr.ID,
				RoundNumberID: c.RoundNumberID,
				SeasonID:      c.SeasonID,
				OdiSeason:     c.PrjSeason,
				SeasonWeek:    c.SeasonWeek,
				StartTime:     c.StartTime,
				EndTime:       c.EndTime,
				MinScored:     x.MinuteScored, //scoreTime,
			}

			g = append(g, d)

		}
	}

	if len(g) == 0 {
		return g, fmt.Errorf("live score not found")
	}

	return g, nil
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
