package players

type Players struct {
	PlayerID   string
	ProfileTag string
	Status     string
	Created    string
	Modified   string
}

type PlayerRequests struct {
	ProfileTag string `json:"profile_tag"`
}
