package misc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Vector2 struct {
	X float64
	Y float64
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

type TwitchConfig struct {
	// Required
	// Your twitch username (lowercase)
	Broadcaster string `json:"broadcaster"`

	// Required.
	// The name of the channel on twitch to connect to
	ChannelName string `json:"channel_name"`

	// Required
	// Twitch OAUTH Access Token. Keep this secret
	TwitchAccessToken string `json:"twitch_access_token"`

	// Optional.
	// If left empty, then assumes this is the same as the Broadcaster
	// You can set this so this so that the bot respons under a different name
	// other than the broadcasters. Just make sure to set the TwitchAccessToken
	// to that of the Bot instead of the Broadcasters
	TwitchBot string `json:"twitch_bot"`

	// Optional.
	// List of usernames to exclude from getting a chibi
	ExcludeNames []string `json:"exclude_names"`
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

	if len(config.Broadcaster) == 0 {
		return nil, fmt.Errorf("broadcaster not set in twitch config (%s)", path)
	}
	if len(config.ChannelName) == 0 {
		return nil, fmt.Errorf("channel_name not set in twitch config (%s)", path)
	}

	if len(config.TwitchBot) == 0 {
		config.TwitchBot = config.Broadcaster
	}
	return &config, nil
}
