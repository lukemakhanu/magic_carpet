// This package fetches matches and odds from kiron
package matchData

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

	"github.com/lukemakhanu/veimu_apps/internal/domain/matches"
)

// Compile time interface assertion.
var _ MatchFetcher = (*MatchClient)(nil)

// MatchClient contains methods for interacting with Kiron APIs.
type MatchClient struct {
	feedsEndpoint string
	competition   string
	hours         string
	timezone      string
	timeouts      time.Duration
	client        *http.Client
}

// New initializes a new instance of Match Client.
func New(feedsEndpoint string, competition string, hours string, timezone string, timeouts time.Duration, client *http.Client) (*MatchClient, error) {
	feedsURL, err := url.Parse(feedsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feeds endpoint: %w", err)
	}

	if competition == "" {
		return nil, fmt.Errorf("Competition not set")
	}

	if hours == "" {
		return nil, fmt.Errorf("Hours not set")
	}

	if timezone == "" {
		return nil, fmt.Errorf("Timezone not set")
	}

	if timeouts <= 0 {
		return nil, fmt.Errorf("Timeout not set")
	}

	c := &MatchClient{
		feedsEndpoint: feedsURL.String(),
		competition:   competition,
		hours:         hours,
		timezone:      timezone,
		timeouts:      timeouts,
		client:        client,
	}
	if c.client == nil {
		c.client = defaultHTTPClient
	}

	return c, nil
}

// GetMatch : returns match and odds from third party provider
func (s *MatchClient) GetMatch(ctx context.Context) (matches.MatchJson, error) {

	var feedsUrl = fmt.Sprintf("%s%s%s%s", s.feedsEndpoint, s.competition, "&hours=", s.hours)
	log.Println("Calling... ", feedsUrl)

	response, err := s.client.Get(feedsUrl)
	if err != nil {
		return matches.MatchJson{}, fmt.Errorf("Failed to call match API : %v", err)
	}

	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return matches.MatchJson{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode == http.StatusOK {
		//
		// Start Parsing data.
		return parseMatch(responseBody)
	}

	return matches.MatchJson{}, fmt.Errorf("failed to get match and odds : status: %d, error body: %s", response.StatusCode, responseBody)

}

var defaultHTTPClient = &http.Client{
	Timeout: time.Second * 45,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Second * 45,
		}).Dial,
		TLSHandshakeTimeout: 20 * time.Second,
	},
}

func parseMatch(content []byte) (matches.MatchJson, error) {
	var up matches.MatchJson
	err := json.Unmarshal(content, &up)
	if err != nil {
		return up, fmt.Errorf("failed to unmarshal xml : %w", err)
	}
	return up, nil
}
