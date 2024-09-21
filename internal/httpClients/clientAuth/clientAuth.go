package clientAuth

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
)

// Compile time interface assertion.
var _ ClientAuthFetcher = (*AuthClient)(nil)

// AuthClient contains methods for client interphase
type AuthClient struct {
	clientAuthEndPoint string
	profileTag         string
}

func New(clientAuthEndPoint, profileTag string) (*AuthClient, error) {

	clientAuthURL, err := url.Parse(clientAuthEndPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clientAuthURL endpoint: %w", err)
	}

	if profileTag == "" {
		return nil, fmt.Errorf("profileTag not set")
	}

	c := &AuthClient{
		clientAuthEndPoint: clientAuthURL.String(),
		profileTag:         profileTag,
	}

	return c, nil
}

// GetLiveScores : returns live scores from third party provider
func (s *AuthClient) GetClientAuth(ctx context.Context) (*ClientAuth, error) {

	log.Printf("Calling... %s ", s.clientAuthEndPoint)

	method := "POST"

	rb := ClientAuthReqBody{
		ProfileTag: s.profileTag,
	}

	authPayload, err := json.Marshal(rb)
	if err != nil {
		log.Println("failed to marshall request payload ", err)
		return nil, fmt.Errorf("failed to marshall request payload")
	}
	payload := strings.NewReader(string(authPayload))

	log.Println("auth request body >>> ", string(authPayload))

	req, err := http.NewRequest(method, s.clientAuthEndPoint, payload)

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to initialize new request")
	}

	authorizationHeader := fmt.Sprintf("%s %s", "Bearer", s.profileTag)
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

	var ss *ClientAuth
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
