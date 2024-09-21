package instantGameServer

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lukemakhanu/magic_carpet/internal/domain/blackListedTokens"
	"github.com/lukemakhanu/magic_carpet/internal/domain/blackListedTokens/blackListedTokensMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domain/clientBets"
	"github.com/lukemakhanu/magic_carpet/internal/domain/clientBets/clientBetsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domain/clientProfiles"
	"github.com/lukemakhanu/magic_carpet/internal/domain/clientProfiles/clientProfilesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domain/contactUs"
	"github.com/lukemakhanu/magic_carpet/internal/domain/contactUs/contactUsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domain/iframeProfiles"
	"github.com/lukemakhanu/magic_carpet/internal/domain/iframeProfiles/iframeProfilesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domain/processRedis"
	"github.com/lukemakhanu/magic_carpet/internal/domain/processRedis/redisExec"
	"github.com/lukemakhanu/magic_carpet/internal/domain/shared"
	"github.com/lukemakhanu/magic_carpet/internal/domain/shared/sharedFun"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests"
	"github.com/lukemakhanu/magic_carpet/internal/domains/matchRequests/matchRequestsMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/players"
	"github.com/lukemakhanu/magic_carpet/internal/domains/players/playersMysql"
	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches"
	"github.com/lukemakhanu/magic_carpet/internal/domains/selectedMatches/selectedMatchesMysql"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/clientPlaceBet"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp"
	"github.com/lukemakhanu/magic_carpet/internal/httpClients/sharedHttp/sharedHttpConf"
)

var JwtKey = []byte("IR8uuaMHpiqqrj8taK81Ooue_iEsobEIANOUXN")

type Claims struct {
	Msisdn     string  `json:"msisdn"`
	ProfileTag string  `json:"profile_tag"`
	AuthToken  string  `json:"auth_token"`
	FirstName  string  `json:"first_name"`
	MiddleName string  `json:"middle_name"`
	LastName   string  `json:"last_name"`
	Balance    float64 `json:"balance"`
	Bonus      float64 `json:"bonus"`
	jwt.RegisteredClaims
}

// InstantGameServerConfiguration is an alias for a function that will take in a pointer to an InstantGameServerService and modify it
type InstantGameServerConfiguration func(os *InstantGameServerService) error

// InstantGameServerService is a implementation of the InstantGameServerService
type InstantGameServerService struct {
	redisConn        processRedis.RunRedis
	redisProfileConn processRedis.RunRedis
	sharedFunc       shared.SharedFunRepository
	httpConf         sharedHttp.SharedHttpConfRepository
	clientBetsMysql  clientBets.ClientBetsRepository
	// New implentation
	clientProfilesMysql    clientProfiles.ClientProfilesRepository
	iframeProfilesMysql    iframeProfiles.IframeProfilesRepository
	blackListedTokensMysql blackListedTokens.BlackListedTokensRepository
	contactUsMysql         contactUs.ConactUsRepository
	//
	// Instant
	playersMysql       players.PlayersRepository
	matchRequestMysql  matchRequests.MatchRequestsRepository
	selectedMatchMysql selectedMatches.SelectedMatchesRepository

	// end of instant

	key              string // encryption key
	iv               string // initialization vector
	authURL          string
	infoURL          string
	betURL           string
	resultURL        string
	selectedTimeZone *time.Location
}

func NewInstantGameServerService(cfgs ...InstantGameServerConfiguration) (*InstantGameServerService, error) {
	// Create the NewClientAPIService
	os := &InstantGameServerService{}
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

func WithRedisRegisterClientRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisProfileConn = d
		return nil
	}
}

func WithRedisResultsRepository(redisServer string, dbNum int, maxIdle int, maxActive int, idleTimeout time.Duration) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := redisExec.New(redisServer, dbNum, maxIdle, maxActive, idleTimeout)
		if err != nil {
			return err
		}
		os.redisConn = d
		return nil
	}
}

// WithMysqlClientBetsRepository : instantiates mysql to connect to bet slip interface
func WithMysqlClientBetsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		// Create bets repo
		d, err := clientBetsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.clientBetsMysql = d
		return nil
	}
}

// WithSharedFuncRepository : shared functions
func WithSharedFuncRepository() InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		cr, err := sharedFun.New()
		if err != nil {
			return err
		}
		os.sharedFunc = cr
		return nil
	}
}

// WithMysqlClientProfilesRepository : instantiates mysql to connect to client profile interface
func WithMysqlClientProfilesRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		// Create bet slip repo
		d, err := clientProfilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.clientProfilesMysql = d
		return nil
	}
}

// WithMysqlIframeProfilesRepository : instantiates mysql to connect to client profile interface
func WithMysqlIframeProfilesRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		// Create bet slip repo
		d, err := iframeProfilesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.iframeProfilesMysql = d
		return nil
	}
}

// WithMysqlBlackListedTokensRepository
func WithMysqlBlackListedTokensRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		// Create bet slip repo
		d, err := blackListedTokensMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.blackListedTokensMysql = d
		return nil
	}
}

// WithMysqlContactUsRepository
func WithMysqlContactUsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := contactUsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.contactUsMysql = d
		return nil
	}
}

// WithProjectConstants : variables used across the project
func WithProjectConstants(key, iv, authURL, infoURL, betURL, resultURL, selTimeZone string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		os.key = key
		os.iv = iv
		os.betURL = betURL
		os.infoURL = infoURL
		os.authURL = authURL
		os.resultURL = resultURL
		sZone, _ := time.LoadLocation(selTimeZone)
		os.selectedTimeZone = sZone

		return nil
	}
}

/*
********** New implementation
 */

func WithMysqlPlayersRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := playersMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.playersMysql = d
		return nil
	}
}

func WithMysqlMatchesRequestsRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := matchRequestsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.matchRequestMysql = d
		return nil
	}
}

func WithMysqlSelectedMatchesRepository(connectionString string) InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		d, err := selectedMatchesMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.selectedMatchMysql = d
		return nil
	}
}

/*
********** end of new implementation
 */

// WithSharedHttpConfRepository : shared functions
func WithSharedHttpConfRepository() InstantGameServerConfiguration {
	return func(os *InstantGameServerService) error {
		cr, err := sharedHttpConf.New()
		if err != nil {
			return err
		}
		os.httpConf = cr
		return nil
	}
}

// GetCORS : return cors
func (s *InstantGameServerService) GetCORS() gin.HandlerFunc {
	return s.httpConf.CORSMiddleware()
}

// SecureHeaders
func (s *InstantGameServerService) SecureHeaders(allowedHost string) gin.HandlerFunc {
	return s.httpConf.SecurityMiddleware(allowedHost)
}

// New implementation

func (s *InstantGameServerService) RegisterClient(c *gin.Context) {
	var p clientProfiles.ClientProfiles
	err := c.Bind(&p)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create client"})
		return
	}

	count, err := s.clientProfilesMysql.CountProfile(c.Request.Context(), p.Msisdn)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.Error(c.Writer, http.StatusBadRequest, err, err.Error())
		return
	}
	if count > 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "client already registered"})
		return
	}

	var prfID string
	//status := "1"
	primaryKey := false

	passwd, err := s.sharedFunc.HashPassword(p.Password)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create client"})
	}

	randomToken := make([]byte, 32)
	_, err = rand.Read(randomToken)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create client"})
		return
	}

	profileTag := base64.URLEncoding.EncodeToString(randomToken)

	pp, err := clientProfiles.NewParentProfile(prfID, p.Msisdn, p.FirstName, p.MiddleName,
		p.LastName, passwd, profileTag, p.Balance, p.Bonus, primaryKey)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create client"})
		return
	}

	profileID, err := s.clientProfilesMysql.Save(c.Request.Context(), *pp)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusInternalServerError, gin.H{"error": "unable to create client"})
		return
	}

	message := fmt.Sprintf("Client of profile_id %d created successfully", profileID)

	var desc = clientProfiles.RegisterClientData{
		ProfileID: profileID,
		Messsage:  message,
	}

	var data = clientProfiles.RegisterClientResponse{
		StatusCode:        http.StatusOK,
		StatusDescription: "OK",
		Data:              desc,
	}
	s.httpConf.JSON(c.Writer, http.StatusOK, data)

}

func (s *InstantGameServerService) ContactUs(c *gin.Context) {
	var p contactUs.ContactUsAPI
	err := c.Bind(&p)
	if err != nil {
		log.Printf("Err :: %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to send contact request"})
		return
	}

	if len(p.FullNames) == 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "Names not provided"})
		return
	}

	if len(p.EmailAddress) == 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "Email not provided"})
		return
	}

	if len(p.Message) == 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "Message is empty"})
		return
	}

	if len(p.Telephone) == 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "Mobile number not provided"})
		return
	}

	pp, err := contactUs.NewContactUs(p.FullNames, p.EmailAddress, p.Telephone, p.Message)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to submit contact information"})
		return
	}

	contactUsID, err := s.contactUsMysql.Save(c.Request.Context(), *pp)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusInternalServerError, gin.H{"error": "unable to submit contact information"})
		return
	}
	log.Printf("Contact %d saved", contactUsID)

	var data = clientProfiles.RegisterClientResponse{
		StatusCode:        http.StatusOK,
		StatusDescription: "OK",
		Data:              "You have successfully submitted your contact information",
	}
	s.httpConf.JSON(c.Writer, http.StatusOK, data)

}

func (s *InstantGameServerService) GetCLientLoginData(c *gin.Context) {
	type payload struct {
		Msisdn   string `json:"msisdn"`
		Password string `json:"password"`
	}
	var pu payload
	err := c.Bind(&pu)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create client"})
		return
	}
	// Check if empty and return
	if len(pu.Msisdn) == 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "missing mobile number"})
		return
	}

	// Check if empty and return
	if len(pu.Password) == 0 {
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "missing password"})
		return
	}

	loginData, err := s.clientProfilesMysql.GetClientLoginData(c.Request.Context(), pu.Msisdn)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to log in"})
		return
	}

	err = s.sharedFunc.CheckPassword(pu.Password, loginData.Password)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to log in"})
		return
	}

	// Create session variable that will be used later in the other.

	ld, err := s.clientProfilesMysql.ClientGenerateToken(c, pu.Msisdn, pu.Password)
	if err != nil {
		log.Printf("Err : %v", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to log in"})
		return
	}

	token, err := s.GenerateToken(pu.Msisdn, ld.ProfileTag, ld.AuthToken, ld.FirstName, ld.MiddleName, ld.LastName, ld.Balance, ld.Bonus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//c.JSON(http.StatusOK, gin.H{"token": token})

	var desc = clientProfiles.LoginClientData{
		ProfileTag:   ld.ProfileTag,
		FirstName:    ld.FirstName,
		MiddleName:   ld.MiddleName,
		LastName:     ld.LastName,
		Balance:      ld.Balance,
		BonusBalance: ld.Bonus,
		GeneratedAt:  ld.GeneratedAt,
		ExpiresAt:    ld.ExpiresAt,
		Messsage:     "Logged in successfully",
		Token:        token,
	}

	var data = clientProfiles.RegisterClientResponse{
		StatusCode:        http.StatusOK,
		StatusDescription: "OK",
		Data:              desc,
	}
	s.httpConf.JSON(c.Writer, http.StatusOK, data)
}

func (s *InstantGameServerService) GetProfileDetails(c *gin.Context) {

	type payload struct {
		ProfileTag string `json:"profile_tag"`
	}
	var pu payload
	err := c.Bind(&pu)
	if err != nil {
		s.httpConf.Error(c.Writer, http.StatusBadRequest, err, err.Error())
		return
	}

	log.Println("data coming in", pu.ProfileTag)

	if len(pu.ProfileTag) == 0 {
		log.Println("data that come ", pu)
		err := errors.New("wrong profile tag sent")
		s.httpConf.Error(c.Writer, http.StatusBadRequest, err, err.Error())
		return
	}

	pd, err := s.clientProfilesMysql.ProfileByProfileTag(c.Request.Context(), pu.ProfileTag)
	if err != nil {
		log.Printf("Err is : %v \n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("ProfileByProfileTag >> ", pd)

	var desc = clientProfiles.ClientProfileDetails{
		ProfileTag:   pd[0].ProfileTag,
		Msisdn:       pd[0].Msisdn,
		FirstName:    pd[0].FirstName,
		MiddleName:   pd[0].MiddleName,
		LastName:     pd[0].LastName,
		Balance:      pd[0].Balance,
		BonusBalance: pd[0].BonusBalance,
	}

	var data = clientProfiles.RegisterClientResponse{
		StatusCode:        http.StatusOK,
		StatusDescription: "OK",
		Data:              desc,
	}
	log.Println("Data to return ", data)
	s.httpConf.JSON(c.Writer, http.StatusOK, data)
	return
}

// SubmitBetToClient : submits bet to clients side on success we can now update this bet
// to be processed.
func (s *InstantGameServerService) SubmitBetToClient(c *gin.Context, auth string, pb clientBets.SubmitBetToClient) (*clientPlaceBet.SubmitBetResponse, error) {
	d, err := clientPlaceBet.New(s.betURL, auth)
	log.Println("res ", d)
	if err != nil {
		log.Printf("Err : %v", err)
		return nil, err
	}
	data, err := d.SubmitBet(c, pb)
	if err != nil {
		log.Printf("Err : %v", err)
		return nil, err
	}
	return data, nil
}

// ClientPlaceBet : place bet on client database.
func (s *InstantGameServerService) ClientPlaceBet(c *gin.Context) {

	var p clientBets.SubmitBets
	err := c.Bind(&p)
	if err != nil {
		log.Printf("Err : %v", err)

		var rs = iframeProfiles.WebResponse{
			StatusCode:        http.StatusBadRequest,
			StatusDescription: "error please contact system admin",
		}
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, rs)
		return
	}

	log.Printf("profileTag:%s, betAmount:%f, possibleWin:%f, status:%s", p.ProfileTag, p.BetAmount, p.PossibleWin, p.Status)

	pbet, err := s.clientBetsMysql.PlaceBet(c.Request.Context(), &p)
	if err != nil {
		log.Printf("err : %v", err)

		var rs = iframeProfiles.WebResponse{
			StatusCode:        http.StatusBadRequest,
			StatusDescription: pbet.Message,
		}
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, rs)
		return
	}

	var pb = clientBets.PlaceBetAPIResponse{
		Message:    pbet.Message,
		ProfileTag: p.ProfileTag,
		Balance:    pbet.Balance,
		Bonus:      pbet.Bonus,
	}
	log.Printf("respond to send : %v", pb)

	var sss = iframeProfiles.WebResponse{
		StatusCode:        http.StatusAccepted,
		StatusDescription: "OK",
		Data:              pb,
	}

	s.httpConf.JSON(c.Writer, http.StatusAccepted, sss)
}

// ClientResultBet : result bet on client database.
func (s *InstantGameServerService) ClientResultBet(c *gin.Context) {

	var p clientBets.SubmitBets
	err := c.Bind(&p)
	if err != nil {
		log.Printf("Err : %v", err)

		var rs = iframeProfiles.WebResponse{
			StatusCode:        http.StatusBadRequest,
			StatusDescription: "error please contact system admin",
		}
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, rs)
		return
	}

	pbet, err := s.clientBetsMysql.ResultBet(c.Request.Context(), &p)
	if err != nil {
		log.Printf("err : %v", err)

		var rs = iframeProfiles.WebResponse{
			StatusCode:        http.StatusBadRequest,
			StatusDescription: pbet.Message,
		}
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, rs)
		return
	}

	var pb = clientBets.PlaceBetAPIResponse{
		Message:    pbet.Message,
		ProfileTag: p.ProfileTag,
		Balance:    pbet.Balance,
		Bonus:      pbet.Bonus,
	}
	log.Printf("respond to send : %v", pb)

	var sss = iframeProfiles.WebResponse{
		StatusCode:        http.StatusAccepted,
		StatusDescription: "OK",
		Data:              pb,
	}

	s.httpConf.JSON(c.Writer, http.StatusAccepted, sss)
}

func (s *InstantGameServerService) GenerateToken(msisdn, profileTag, authToken, firstName, middleName, lastName string, balance, bonus float64) (string, error) {
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		Msisdn:     msisdn,
		ProfileTag: profileTag,
		AuthToken:  authToken,
		FirstName:  firstName,
		MiddleName: middleName,
		LastName:   lastName,
		Balance:    balance,
		Bonus:      bonus,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *InstantGameServerService) ValidateToken(c *gin.Context, signedToken string) (bool, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		},
	)
	if err != nil {
		return false, err
	}
	claims, ok := token.Claims.(*Claims)

	log.Printf("firstName:%s, middleName:%s, lastName:%s, balance:%f, bonus:%f, authToken:%s, profileTag:%s",
		claims.FirstName, claims.MiddleName, claims.LastName, claims.Balance, claims.Bonus, claims.AuthToken, claims.ProfileTag)

	if !ok {
		return false, err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return false, err
	}

	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "auth_token", signedToken) //claims.AuthToken)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	const timeLayout = "2006-01-02 15:04:05"
	expiryTime := time.Now().Add(time.Minute * 30).In(s.selectedTimeZone).Format(timeLayout)
	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "expiry_time", expiryTime)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	bal := fmt.Sprintf("%f", claims.Balance)
	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "balance", bal)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	bonus := fmt.Sprintf("%f", claims.Bonus)
	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "bonus", bonus)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "first_name", claims.FirstName)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "middle_name", claims.MiddleName)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	err = s.redisProfileConn.HSet(c, claims.ProfileTag, "last_name", claims.LastName)
	if err != nil {
		log.Printf("Err : %v", err)
		return false, err
	}

	return true, nil
}

func (s *InstantGameServerService) RemoveToken(c *gin.Context, signedToken string) error {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		},
	)
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || claims.ExpiresAt.Time.Before(time.Now()) {
		return err
	}

	d, err := blackListedTokens.NewBlackListedToken(signedToken)
	log.Println("res ", d)
	if err != nil {
		log.Printf("Err : %v", err)
		return err
	}
	data, err := s.blackListedTokensMysql.Save(c, *d)
	if err != nil {
		log.Printf("Err : %v", err)
		return err
	}
	log.Printf("id : %d", data)
	return nil
}

func (s *InstantGameServerService) IsBlackListed(c *gin.Context, token string) (bool, error) {
	data, err := s.blackListedTokensMysql.IsBlacklisted(c, token)
	if err != nil {
		log.Printf("Err : %v", err)
		return data, err
	}
	log.Println("id : ", data)
	return data, nil
}

func (s *InstantGameServerService) JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}

		// Asegúrate de que el token tenga el formato "Bearer <token>"
		tokenParts := strings.Split(tokenString, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Verificar si el token está en la lista negra
		isBlacklisted, err := s.IsBlackListed(c, token)
		if err != nil {
			log.Println("err : ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error on token used"})
			c.Abort()
			return
		}

		if isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is blacklisted"})
			c.Abort()
			return
		}

		// Validar el token
		_, err = s.ValidateToken(c, token)
		if err != nil {
			log.Println("err : ", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to log in"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *InstantGameServerService) QueryInstantGames(c *gin.Context) {
	var p players.PlayerRequests
	err := c.Bind(&p)
	if err != nil {
		log.Printf("err : %v unable to create requests for matches", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create requests for matches"})
		return
	}

	player, err := s.playersMysql.PlayerExists(c.Request.Context(), p.ProfileTag)
	if err != nil {
		log.Printf("err : %v unable to return a player information", err)
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to return a player information"})
		return
	}

	selectedPlayerID := []string{}

	if len(player) == 0 {

		// Start by saving the player in our system

		status := "active"
		pp, err := players.NewPlayers(p.ProfileTag, status)
		if err != nil {
			log.Printf("err : %v unable to initialize a player", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to initialize a player"})
			return
		}

		playerID, err := s.playersMysql.Save(c.Request.Context(), *pp)
		if err != nil {
			log.Printf("err : %v unable to create a player", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create a player"})
			return
		}

		// Save match requests

		plyID := fmt.Sprintf("%d", playerID)
		selectedPlayerID = append(selectedPlayerID, plyID)
	} else {

		pID := player[0].PlayerID
		selectedPlayerID = append(selectedPlayerID, pID)

	}

	if len(selectedPlayerID) != 1 {
		log.Printf("Error on capturing playerID")
		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to fetch player Identity"})
		return
	}

	plyID := selectedPlayerID[0]

	cTime := time.Now()

	st := 10
	et := 45
	var sTime = cTime.Add(time.Second * time.Duration(st))
	var eTime = cTime.Add(time.Second * time.Duration(et))

	var startTime = sTime.Format("2006-01-02 15:04:05")
	var endTime = eTime.Format("2006-01-02 15:04:05")

	earlyFinish := "no"
	played := "no"

	competitions := []string{"1", "2", "3", "4"}
	for _, cID := range competitions {

		log.Printf("instantCompetitionID %d", cID)

		pp, err := matchRequests.NewMatchRequests(cID, plyID, startTime, endTime, earlyFinish, played)
		if err != nil {
			log.Printf("err : %v unable to initialize a matchRequests", err)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to initialize a matchRequests"})
			return
		}

		matchRequestID, err := s.matchRequestMysql.Save(c.Request.Context(), *pp)
		if err != nil {
			log.Printf("err : %v unable to create match request for competition id %s", err, cID)
			s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create match request"})
			return
		}

		// Query for matches required for this client.

		for x := 0; x < 10; x++ {

			mrID := fmt.Sprintf("%d", matchRequestID)
			parentMatchID := "0"
			sm, err := selectedMatches.NewSelectedMatches(plyID, mrID, parentMatchID)
			if err != nil {
				log.Printf("err : %v unable to initialize a selectedMatch", err)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to initialize selectedMatch"})
				return
			}

			smID, err := s.selectedMatchMysql.Save(c.Request.Context(), *sm)
			if err != nil {
				log.Printf("err : %v unable to create selected match for competition id %s", err, cID)
				s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "unable to create match request"})
				return
			}
			log.Prin

		}

		s.httpConf.JSON(c.Writer, http.StatusBadRequest, gin.H{"error": "selected match already registered"})
		return

	}

	profileID := 1
	message := fmt.Sprintf("Client of profile_id %d created successfully", profileID)

	var desc = clientProfiles.RegisterClientData{
		ProfileID: profileID,
		Messsage:  message,
	}

	var data = clientProfiles.RegisterClientResponse{
		StatusCode:        http.StatusOK,
		StatusDescription: "OK",
		Data:              desc,
	}
	s.httpConf.JSON(c.Writer, http.StatusOK, data)

}
