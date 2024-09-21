package clientInformation

type ClientInfoApi struct {
	StatusCode        int            `json:"status_code"`
	StatusDescription string         `json:"status_description"`
	Data              ClientInfoData `json:"data"`
}

type ClientInfoData struct {
	FirstName    string  `json:"first_name"`
	MiddleName   string  `json:"middle_name"`
	LastName     string  `json:"last_name"`
	ProfileTag   string  `json:"profile_tag"`
	AuthToken    string  `json:"auth_token"`
	Balance      float64 `json:"balance"`
	BonusBalance float64 `json:"bonus_balance"`
	ExpiresAt    string  `json:"expires_at"`
}

type ClientAuthReqBody struct {
	ProfileTag string `json:"profile_tag"`
}
