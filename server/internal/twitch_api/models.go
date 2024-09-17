package twitch_api

type GetUsersResponseData struct {
	Id              string `json:"id"`
	Login           string `json:"login"`
	Display_name    string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
	Email           string `json:"email"`
	CreatedAt       string `json:"created_at"`
}

type GetUsersResponse struct {
	Data []GetUsersResponseData `json:"data"`
}
