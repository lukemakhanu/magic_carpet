package clientAuth

type ClientAuth struct {
	StatusCode        int            `json:"status_code"`
	StatusDescription string         `json:"status_description"`
	Data              ClientAuthData `json:"data"`
}
type ClientAuthData struct {
	ProfileTag   string `json:"profile_tag"`
	AuthToken    string `json:"auth_token"`
	Balance      string `json:"balance"`
	BonusBalance string `json:"bonus_balance"`
	ExpiresAt    string `json:"expires_at"`
}

type ClientAuthReqBody struct {
	ProfileTag string `json:"profile_tag"`
}