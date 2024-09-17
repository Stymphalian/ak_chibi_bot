package twitch_api

type FakeTwitchApiClient struct {
}

func (f *FakeTwitchApiClient) GetUsers(channel string) (*GetUsersResponse, error) {
	response := &GetUsersResponse{
		Data: []GetUsersResponseData{
			{
				Id:              "1",
				Login:           "1",
				Display_name:    "display_name",
				Type:            "1",
				BroadcasterType: "1",
				Description:     "1",
				ProfileImageUrl: "1",
				OfflineImageUrl: "1",
				ViewCount:       1,
				Email:           "test@example.com",
				CreatedAt:       "2022-10-03T14:00:00Z",
			},
		},
	}
	return response, nil
}
func NewFakeTwitchApiClient() TwitchApiClientInterface {
	return &FakeTwitchApiClient{}
}
