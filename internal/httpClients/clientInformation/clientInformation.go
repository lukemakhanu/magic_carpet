package clientInformation

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
var _ ClientInformationFetcher = (*InfoClient)(nil)

// AuthClient contains methods for client interphase
type InfoClient struct {
	clientInformationEndPoint string
	profileTag                string
	signedToken               string
}

func New(clientInformationEndPoint, profileTag, signedToken string) (*InfoClient, error) {

	clientInformationURL, err := url.Parse(clientInformationEndPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clientInformationURL endpoint: %w", err)
	}

	if profileTag == "" {
		return nil, fmt.Errorf("profileTag not set")
	}

	if signedToken == "" {
		return nil, fmt.Errorf("signedToken not set")
	}

	c := &InfoClient{
		clientInformationEndPoint: clientInformationURL.String(),
		profileTag:                profileTag,
		signedToken:               signedToken,
	}

	return c, nil
}

// GetLiveScores : returns live scores from third party provider
func (s *InfoClient) GetClientInfo(ctx context.Context) (*ClientInfoApi, error) {

	log.Printf("Calling... %s ", s.clientInformationEndPoint)

	method := "POST"

	rb := ClientAuthReqBody{
		ProfileTag: s.profileTag,
	}

	betPayload, err := json.Marshal(rb)
	if err != nil {
		log.Println("failed to marshall request payload ", err)
		return nil, fmt.Errorf("failed to marshall request payload")
	}

	log.Println("betPayload >>> ", string(betPayload))

	payload := strings.NewReader(string(betPayload))
	req, err := http.NewRequest(method, s.clientInformationEndPoint, payload)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to initialize new request")
	}

	req.Header.Add("Content-Type", "application/json")
	authorizationHeader := fmt.Sprintf("%s %s", "Bearer", s.signedToken)
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

	var ss *ClientInfoApi
	err = json.Unmarshal(responseBody, &ss)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json : %w", err)
	}
	log.Println("statusCode : ", ss)

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
