package twitch_api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type TwitchApiClientInterface interface {
	GetOpenIdConfiguration() (*GetOpenIdConfigurationResponse, error)
	GetUsers(channel string) (*GetUsersResponse, error)
	GetUsersById(userIds ...string) (*GetUsersResponse, error)
	RevokeToken(accessToken string) error
	ValidateToken(token string) (*ValidateTokenResponse, error)
	RefreshToken(clientSecret string, refreshToken string) (*RefreshTokenResponse, error)
}

type TwitchApiClient struct {
	ClientId    string
	AccessToken string
	httpClient  *http.Client
}

func ProvideTwitchApiClient(botConfig *misc.BotConfig) *TwitchApiClient {
	log.Println("ProvideTwitchApiClient")
	return NewTwitchApiClient(
		botConfig.TwitchClientId,
		botConfig.TwitchAccessToken,
	)
}

func NewTwitchApiClient(clientId string, accessToken string) *TwitchApiClient {
	return &TwitchApiClient{
		ClientId:    clientId,
		AccessToken: accessToken,
		httpClient:  &http.Client{Timeout: time.Duration(5) * time.Second},
	}
}

func (c *TwitchApiClient) GetOpenIdConfiguration() (*GetOpenIdConfigurationResponse, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://id.twitch.tv/oauth2/.well-known/openid-configuration",
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %v", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var respBody GetOpenIdConfigurationResponse
	if err = decoder.Decode(&respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *TwitchApiClient) GetUsers(channel string) (*GetUsersResponse, error) {
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

func (c *TwitchApiClient) GetUsersById(userIds ...string) (*GetUsersResponse, error) {
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
	for _, userId := range userIds {
		q.Add("id", userId)
	}
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

func (c *TwitchApiClient) RevokeToken(accessToken string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		"https://id.twitch.tv/oauth2/revoke",
		strings.NewReader(fmt.Sprintf("client_id=%s&token=%s", c.ClientId, accessToken)),
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %v", resp.StatusCode)
	}
	return nil
}

func (c *TwitchApiClient) ValidateToken(token string) (*ValidateTokenResponse, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://id.twitch.tv/oauth2/validate",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyString, err := io.ReadAll(resp.Body)
		if err != nil {
			bodyString = []byte("unable to read body")
		}
		return nil, fmt.Errorf("request failed with status %v:%s", resp.StatusCode, string(bodyString))
	}

	decoder := json.NewDecoder(resp.Body)
	var respBody ValidateTokenResponse
	if err = decoder.Decode(&respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *TwitchApiClient) RefreshToken(clientSecret string, refreshToken string) (*RefreshTokenResponse, error) {
	body := strings.NewReader(
		fmt.Sprintf(
			"client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s",
			c.ClientId,
			clientSecret,
			refreshToken,
		),
	)

	req, err := http.NewRequest(
		http.MethodPost,
		"https://id.twitch.tv/oauth2/token",
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %v", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var respBody RefreshTokenResponse
	if err = decoder.Decode(&respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}
