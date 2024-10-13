package misc

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

type BotConfig struct {
	// Required
	// Twitch Client ID. Can be public. Used to make requests to Twitch API.
	TwitchClientId string `json:"twitch_client_id"`

	// Required
	// Twitch Client Secret. Keep this secret
	TwitchClientSecret string `json:"twitch_client_secret"`

	// Required
	// Twitch OAUTH Access Token. Keep this secret
	TwitchAccessToken string `json:"twitch_access_token"`

	// Required
	// Twitch Oauth Redirect URL. Where to redirect during the OAUTH login
	// procedures. This should be updated for dev and prod environments
	TwitchOauthRedirectUrl string `json:"twitch_oauth_redirect_url"`

	// Required.
	// The twitch name of the bot. This should be the channel in which
	// the bot was registered with and it should match with the TwitchAccessToken
	TwitchBot string `json:"twitch_bot"`

	// Option.
	// List of usernames to exclude from getting a chibi
	ExcludeNames []string `json:"exclude_names"`

	// Option
	// Default to Amiya
	InitialOperator string `json:"initial_operator"`

	// Option
	// Defualt operator details
	// Otherwise set to default, base, Move, 0.5
	OperatorDetails InitialOperatorDetails `json:"operator_details"`

	// Option
	// The interval in minutes in which to remove chibis from the screen if the
	// user have not chatted. Set to -1 to never remove chibis
	// Default: 40 minutes
	RemoveChibiAfterMinutes int `json:"remove_chibi_after_minutes"`

	// Option
	// The interval in minutes to check when to garbage collect unused
	// chat rooms
	// Default: 360 (6 hours)
	RemoveUnusedRoomsAfterMinutes int `json:"remove_unused_rooms_after_minutes"`

	// Optional
	// Default settings for the spine runtime. Min/max/default values
	SpineRuntimeConfig *SpineRuntimeConfig `json:"spine_runtime_config"`

	// Required.
	// A secret key used for encrypting session cookies.
	CookieSecret string `json:"cookie_secret"`

	// Required
	// Secret key used for encrypting JWT tokens. Must be kept safe.
	JwtSecretKey string `json:"jwt_secret_key"`

	// Required
	// CSRF Secret Key
	CsrfSecret string `json:"csrf_secret"`
}

func LoadBotConfig(path string) (*BotConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := BotConfig{}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	if len(config.TwitchBot) == 0 {
		return nil, fmt.Errorf("twitch_bot not set in bot config (%s)", path)
	}
	if len(config.TwitchClientId) == 0 {
		return nil, fmt.Errorf("twitch_client_id not set in bot config (%s)", path)
	}
	if len(config.TwitchClientSecret) == 0 {
		return nil, fmt.Errorf("twitch_client_secret not set in bot config (%s)", path)
	}
	if len(config.TwitchAccessToken) == 0 {
		return nil, fmt.Errorf("twitch_access_token not set in bot config (%s)", path)
	}
	if len(config.TwitchOauthRedirectUrl) == 0 {
		return nil, fmt.Errorf("twitch_oauth_redirect_url not set in bot config (%s)", path)
	}
	if _, err := url.Parse(config.TwitchOauthRedirectUrl); err != nil {
		return nil, fmt.Errorf("twitch_oauth_redirect_url is not a valid URL in bot config (%s): %w", path, err)
	}

	if len(config.CookieSecret) == 0 {
		return nil, fmt.Errorf("cookie_secret not set in bot config (%s)", path)
	}
	if config.RemoveChibiAfterMinutes == 0 {
		config.RemoveChibiAfterMinutes = 40
	}
	if config.RemoveUnusedRoomsAfterMinutes == 0 {
		config.RemoveUnusedRoomsAfterMinutes = 360
	}
	if config.OperatorDetails.Skin == "" {
		config.OperatorDetails.Skin = "default"
	}
	if config.OperatorDetails.Stance == "" {
		config.OperatorDetails.Stance = "base"
	}
	if len(config.OperatorDetails.Animations) == 0 {
		if config.OperatorDetails.Stance == "base" {
			config.OperatorDetails.Animations = []string{"Move"}
		} else {
			config.OperatorDetails.Animations = []string{"Idle"}
		}
	}
	if config.OperatorDetails.PositionX == 0 {
		config.OperatorDetails.PositionX = 0.5
	}
	if config.SpineRuntimeConfig == nil {
		config.SpineRuntimeConfig = DefaultSpineRuntimeConfig()
	} else {
		if err := ValidateSpineRuntimeConfig(config.SpineRuntimeConfig); err != nil {
			return nil, err
		}
	}
	if len(config.CsrfSecret) == 0 {
		return nil, fmt.Errorf("csrf_secret not set in bot config (%s)", path)
	}
	if len(config.CsrfSecret) != 32 {
		return nil, fmt.Errorf("csrf_secret must be 32 characters in bot config (%s)", path)
	}
	if len(config.CookieSecret) == 0 {
		return nil, fmt.Errorf("cookie_secret not set in bot config (%s)", path)
	}
	if len(config.JwtSecretKey) == 0 {
		return nil, fmt.Errorf("jwt_secret_key not set in bot config (%s)", path)
	}

	return &config, nil
}
