package misc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSpineRuntimeConfig(t *testing.T) {
	// Default config should validate
	defaultConfig := DefaultSpineRuntimeConfig()
	assert.NoError(t, ValidateSpineRuntimeConfig(defaultConfig))

	// Tets for default animation speed being between min and max
	defaultConfig.MinAnimationSpeed = 1.0
	defaultConfig.MaxAnimationSpeed = 2.0
	defaultConfig.DefaultAnimationSpeed = 2.5
	err := ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_animation_speed must be between")

	// Test min_animation_speed
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.MinAnimationSpeed = 2.0
	defaultConfig.MaxAnimationSpeed = 1.0
	defaultConfig.DefaultAnimationSpeed = 1.5
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min_animation_speed must be less than max_animation_speed")

	// test scale_size
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.MinScaleSize = 1.0
	defaultConfig.MaxScaleSize = 0.5
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min_scale_size must be less than max_scale_size")

	// default scale size must be betweeen min and max
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.DefaultScaleSize = 0.25
	defaultConfig.MinScaleSize = 0.5
	defaultConfig.MaxScaleSize = 1.0
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_scale_size must be between")

	// test max_sprite_pixel_size
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.MaxSpritePixelSize = -1
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_sprite_pixel_size must be greater than 0")

	// test reference_move_speed_px
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.ReferenceMovementSpeedPx = 0
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reference_movement_speed_px must be greater than 0")

	// test default_movement_speed
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.DefaultMovementSpeed = 6.0
	defaultConfig.MaxMovementSpeed = 5.0
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_movement_speed must be between")

	// test min_movement_speed
	defaultConfig = DefaultSpineRuntimeConfig()
	defaultConfig.MinMovementSpeed = 6.0
	defaultConfig.MaxMovementSpeed = 5.0
	err = ValidateSpineRuntimeConfig(defaultConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min_movement_speed must be less than max_movement_speed")
}
