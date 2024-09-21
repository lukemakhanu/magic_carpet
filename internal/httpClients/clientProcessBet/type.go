package clientProcessBet

type SubmitBetResponse struct {
	StatusCode        int          `json:"status_code"`
	StatusDescription string       `json:"status_description"`
	Data              ResponseData `json:"data"`
}

type ResponseData struct {
	Message      string  `json:"message"`
	ProfileTag   string  `json:"profile_tag"`
	AuthToken    string  `json:"auth_token"`
	Balance      float64 `json:"balance"`
	BonusBalance float64 `json:"bonus_balance"`
}
