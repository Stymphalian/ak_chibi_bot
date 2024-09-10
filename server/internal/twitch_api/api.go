package twitch_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	ClientId    string
	AccessToken string
	httpClient  *http.Client
}

func NewClient(clientId string, accessToken string) *Client {
	return &Client{
		ClientId:    clientId,
		AccessToken: accessToken,
		httpClient:  &http.Client{Timeout: time.Duration(5) * time.Second},
	}
}

func (c *Client) GetUsers(channel string) (*GetUsersResponse, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.twitch.tv/helix/users",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Add("Client-Id", c.ClientId)
	q := req.URL.Query()
	q.Add("login", channel)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %v", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var respBody GetUsersResponse
	if err = decoder.Decode(&respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}
