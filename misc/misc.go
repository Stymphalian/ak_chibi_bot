package misc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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

type BroadcasterConfig struct {
	// Required
	// Your twitch username (lowercase)
	Broadcaster string `json:"broadcaster"`

	// Required.
	// The name of the channel on twitch to connect to
	ChannelName string `json:"channel_name"`

	// // Required
	// // Twitch OAUTH Access Token. Keep this secret
	// TwitchAccessToken string `json:"twitch_access_token"`
}

type TwitchConfig struct {
	// // Required
	// // Your twitch username (lowercase)
	// Broadcaster string `json:"broadcaster"`

	// // Required.
	// // The name of the channel on twitch to connect to
	// ChannelName string `json:"channel_name"`

	// Required
	// Twitch OAUTH Access Token. Keep this secret
	TwitchAccessToken string `json:"twitch_access_token"`

	// Option.
	// If left empty, then assumes this is the same as the Broadcaster
	// You can set this so this so that the bot responds under a different name
	// other than the broadcasters. Just make sure to set the TwitchAccessToken
	// to that of the Bot instead of the Broadcasters
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
}

type InitialOperatorDetails struct {
	Skin       string   `json:"skin"`
	Stance     string   `json:"stance"`
	Animations []string `json:"animations"`
	PositionX  float64  `json:"position_x"`
}

func LoadTwitchConfig(path string) (*TwitchConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := TwitchConfig{}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	// if len(config.Broadcaster) == 0 {
	// 	return nil, fmt.Errorf("broadcaster not set in twitch config (%s)", path)
	// }
	// if len(config.ChannelName) == 0 {
	// 	return nil, fmt.Errorf("channel_name not set in twitch config (%s)", path)
	// }

	// if len(config.TwitchBot) == 0 {
	// 	config.TwitchBot = config.Broadcaster
	// }
	if config.RemoveChibiAfterMinutes == 0 {
		config.RemoveChibiAfterMinutes = 40
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

func LoadBroadcasterConfig(path string) (*BroadcasterConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := BroadcasterConfig{}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	if len(config.Broadcaster) == 0 {
		return nil, fmt.Errorf("broadcaster not set in twitch config (%s)", path)
	}
	if len(config.ChannelName) == 0 {
		return nil, fmt.Errorf("channel_name not set in twitch config (%s)", path)
	}

	return &config, nil
}
