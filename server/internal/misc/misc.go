package misc

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
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

type InitialOperatorDetails struct {
	Skin       string   `json:"skin"`
	Stance     string   `json:"stance"`
	Animations []string `json:"animations"`
	PositionX  float64  `json:"position_x"`
}

func (oi *InitialOperatorDetails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal InitialOperatorDetails value:", value))
	}

	err := json.Unmarshal(bytes, oi)
	if err != nil {
		return err
	}
	return nil
}

func (oi InitialOperatorDetails) Value() (driver.Value, error) {
	jsonData, err := json.Marshal(oi)
	return string(jsonData), err
}

var channelRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`)

func ValidateChannelName(channelName string) error {
	channelName = strings.TrimSpace(channelName)
	if !channelRegex.MatchString(channelName) {
		return fmt.Errorf("channel name must be alphanumeric and between 1 and 100 characters, was '%s'", channelName)
	}
	return nil
}
