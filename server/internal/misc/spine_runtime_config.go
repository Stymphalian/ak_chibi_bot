package misc

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type SpineRuntimeConfig struct {
	DefaultAnimationSpeed float64 `json:"default_animation_speed"`
	MinAnimationSpeed     float64 `json:"min_animation_speed"`
	MaxAnimationSpeed     float64 `json:"max_animation_speed"`

	DefaultScaleSize   float64 `json:"default_scale_size"`
	MinScaleSize       float64 `json:"min_scale_size"`
	MaxScaleSize       float64 `json:"max_scale_size"`
	MaxSpritePixelSize int     `json:"max_sprite_pixel_size"`

	ReferenceMovementSpeedPx int     `json:"reference_movement_speed_px"`
	DefaultMovementSpeed     float64 `json:"default_movement_speed"`
	MinMovementSpeed         float64 `json:"min_movement_speed"`
	MaxMovementSpeed         float64 `json:"max_movement_speed"`
}

func DefaultSpineRuntimeConfig() *SpineRuntimeConfig {
	return &SpineRuntimeConfig{
		DefaultAnimationSpeed: 1.0,
		MinAnimationSpeed:     0.1,
		MaxAnimationSpeed:     5.0,

		DefaultScaleSize:   1.0,
		MinScaleSize:       0.5,
		MaxScaleSize:       1.5,
		MaxSpritePixelSize: 350,

		ReferenceMovementSpeedPx: 80,
		DefaultMovementSpeed:     1.0,
		MinMovementSpeed:         0.1,
		MaxMovementSpeed:         2.0,
	}
}

func ValidateSpineRuntimeConfig(config *SpineRuntimeConfig) error {
	if config.MinAnimationSpeed > config.MaxAnimationSpeed {
		return fmt.Errorf("min_animation_speed must be less than max_animation_speed")
	}
	if config.DefaultAnimationSpeed < config.MinAnimationSpeed || config.DefaultAnimationSpeed > config.MaxAnimationSpeed {
		return fmt.Errorf("default_animation_speed must be between %f and %f", config.MinAnimationSpeed, config.MaxAnimationSpeed)
	}

	if config.MinScaleSize > config.MaxScaleSize {
		return fmt.Errorf("min_scale_size must be less than max_scale_size")
	}
	if config.DefaultScaleSize < config.MinScaleSize || config.DefaultScaleSize > config.MaxScaleSize {
		return fmt.Errorf("default_scale_size must be between %f and %f", config.MinScaleSize, config.MaxScaleSize)
	}
	if config.MaxSpritePixelSize < 0 {
		return fmt.Errorf("max_sprite_pixel_size must be greater than 0")
	}

	if config.ReferenceMovementSpeedPx <= 0 {
		return fmt.Errorf("reference_movement_speed_px must be greater than 0")
	}
	if config.MinMovementSpeed > config.MaxMovementSpeed {
		return fmt.Errorf("min_movement_speed must be less than max_movement_speed")
	}
	if config.DefaultMovementSpeed < config.MinMovementSpeed || config.DefaultMovementSpeed > config.MaxMovementSpeed {
		return fmt.Errorf("default_movement_speed must be between %f and %f", config.MinMovementSpeed, config.MaxMovementSpeed)
	}

	return nil
}

func (oi *SpineRuntimeConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal SpineRuntimeConfig value:", value))
	}

	err := json.Unmarshal(bytes, oi)
	if err != nil {
		return err
	}
	return nil
}

func (oi SpineRuntimeConfig) Value() (driver.Value, error) {
	jsonData, err := json.Marshal(oi)
	return string(jsonData), err
}
