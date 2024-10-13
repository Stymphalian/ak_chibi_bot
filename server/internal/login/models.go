package login

type LoggedInResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"user_name"`
	TwitchUserId  string `json:"user_id"`
	IsAdmin       bool   `json:"is_admin"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
