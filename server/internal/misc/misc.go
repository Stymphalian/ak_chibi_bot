package misc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func ClampF64(val float64, min float64, max float64) float64 {
	if val < min {
		return min
	} else if val > max {
		return max
	} else {
		return val
	}
}

type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (v Vector2) String() string {
	return fmt.Sprintf("(%f, %f)", v.X, v.Y)
}

func MatchesKeywords(str string, keywords []string) (string, bool) {
	for _, keyword := range keywords {
		if strings.EqualFold(str, keyword) {
			return keyword, true
		}
	}
	return "", false
}

type BotConfig struct {
	// Required
	// Twitch Client ID. Can be public. Used to make requests to Twitch API.
	TwitchClientId string `json:"twitch_client_id"`

	// Required
	// Twitch OAUTH Access Token. Keep this secret
	TwitchAccessToken string `json:"twitch_access_token"`

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
	// chat rooms. Set to -1 to never cleanup rooms
	// Default: 360 (6 hours)
	RemoveUnusedRoomsAfterMinutes int `json:"remove_unused_rooms_after_minutes"`

	// Option
	// Secret key used to give access to /admin endpoints.
	// I can't be bothered with an proper ACL system right now
	AdminSecret string `json:"admin_secret"`

	// Optional
	// Google Cloud Project Id used for creating/restore restore points during
	// the production reloading of the server
	GoogleCloudProjectId string `json:"google_cloud_project_id"`

	// Optional
	// Path to the Google Cloud Credential json file. Should be an absolute path
	// If not provided then Save/Restore won't work. Which is fine unless this
	// this is the production server.
	GoogleCloudProjectCredentialsFilePath string `json:"google_cloud_project_credentials_file_path"`
}

type InitialOperatorDetails struct {
	Skin       string   `json:"skin"`
	Stance     string   `json:"stance"`
	Animations []string `json:"animations"`
	PositionX  float64  `json:"position_x"`
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
		return nil, fmt.Errorf("twitch_bot not set in twitch config (%s)", path)
	}
	if len(config.TwitchClientId) == 0 {
		return nil, fmt.Errorf("twitch_client_id not set in twitch config (%s)", path)
	}
	if len(config.TwitchAccessToken) == 0 {
		return nil, fmt.Errorf("twitch_access_token not set in twitch config (%s)", path)
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

	return &config, nil
}