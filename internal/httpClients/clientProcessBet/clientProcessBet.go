package clientProcessBet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lukemakhanu/veimu_apps/internal/domain/clientBets"
)

// Compile time interface assertion.
var _ ClientProcessBetFetcher = (*InfoClient)(nil)

// AuthClient contains methods for client interphase
type InfoClient struct {
	submitWonBetEndPoint string
	authStr              string
}

func New(submitWonBetEndPoint, auth string) (*InfoClient, error) {

	wonBetURL, err := url.Parse(submitWonBetEndPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wonBetURL endpoint: %w", err)
	}

	if auth == "" {
		return nil, fmt.Errorf("auth not set")
	}

	c := &InfoClient{
		submitWonBetEndPoint: wonBetURL.String(),
		authStr:              auth,
	}

	return c, nil
}

// GetLiveScores : returns live scores from third party provider
func (s *InfoClient) SubmitWonBet(ctx context.Context, pb clientBets.SubmitBetToClient) (*SubmitBetResponse, error) {

	log.Printf("Calling... %s ", s.submitWonBetEndPoint)

	method := "POST"

	betPayload, err := json.Marshal(pb)
	if err != nil {
		log.Println("failed to marshall request payload ", err)
		return nil, fmt.Errorf("failed to marshall request payload")
	}

	log.Println("wonbetPayload >>> ", string(betPayload))

	payload := strings.NewReader(string(betPayload))
	req, err := http.NewRequest(method, s.submitWonBetEndPoint, payload)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to initialize new request")
	}

	req.Header.Add("Content-Type", "application/json")
	authorizationHeader := fmt.Sprintf("%s %s", "Bearer", s.authStr)
	log.Printf("authorizationHeader : %s", authorizationHeader)
	req.Header.Add("Authorization", authorizationHeader)

	res, err := defaultHTTPClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to call client auth API")
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to read response body: %w | responseBody : %s ", err, string(responseBody))
	}
	log.Println(string(responseBody))

	var ss *SubmitBetResponse
	err = json.Unmarshal(responseBody, &ss)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json : %w", err)
	}
	log.Printf("statusCode : %d", ss.StatusCode)

	return ss, nil

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
