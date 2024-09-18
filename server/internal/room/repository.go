package room

import (
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TODO: Fuck this. Too many columns
type RoomDb struct {
	RoomId                    uint           `gorm:"primarykey"`
	ChannelName               string         `gorm:"column:channel_name"`
	IsActive                  bool           `gorm:"column:is_active"`
	DefaultOperatorName       string         `gorm:"column:default_operator_name"`
	DefaultOperatorSkin       string         `gorm:"column:default_operator_skin"`
	DefaultOperatorStance     string         `gorm:"column:default_operator_stance"`
	DefaultOperatorAnimations pq.StringArray `gorm:"column:default_operator_animations;type:text[]"`
	DefaultOperatorPositionX  float64        `gorm:"column:default_operator_position_x"`

	GarbageCollectionPeriodMins int `gorm:"column:garbage_collection_period_mins"`

	SpineRuntimeConfigDefaultAnimationSpeed    float64 `gorm:"spine_runtime_config_default_animation_speed"`
	SpineRuntimeConfigMinAnimationSpeed        float64 `gorm:"spine_runtime_config_min_animation_speed"`
	SpineRuntimeConfigMaxAnimationSpeed        float64 `gorm:"spine_runtime_config_max_animation_speed"`
	SpineRuntimeConfigDefaultScaleSize         float64 `gorm:"spine_runtime_config_default_scale_size"`
	SpineRuntimeConfigMinScaleSize             float64 `gorm:"spine_runtime_config_min_scale_size"`
	SpineRuntimeConfigMaxScaleSize             float64 `gorm:"spine_runtime_config_max_scale_size"`
	SpineRuntimeConfigMaxSpritePixelSize       int     `gorm:"spine_runtime_config_max_sprite_pixel_size"`
	SpineRuntimeConfigReferenceMovementSpeedPx int     `gorm:"spine_runtime_config_reference_movement_speed_px"`
	SpineRuntimeConfigDefaultMovementSpeed     float64 `gorm:"spine_runtime_config_default_movement_speed"`
	SpineRuntimeConfigMinMovementSpeed         float64 `gorm:"spine_runtime_config_min_movement_speed"`
	SpineRuntimeConfigMaxMovementSpeed         float64 `gorm:"spine_runtime_config_max_movement_speed"`

	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RoomDb) TableName() string {
	return "rooms"
}

func (r *RoomDb) DefaultOperatorConfig() misc.InitialOperatorDetails {
	return misc.InitialOperatorDetails{
		Skin:       r.DefaultOperatorSkin,
		Stance:     r.DefaultOperatorStance,
		Animations: r.DefaultOperatorAnimations,
		PositionX:  r.DefaultOperatorPositionX,
	}
}

func (r *RoomDb) SpineRuntimeConfig() misc.SpineRuntimeConfig {
	return misc.SpineRuntimeConfig{
		DefaultAnimationSpeed:    r.SpineRuntimeConfigDefaultAnimationSpeed,
		MinAnimationSpeed:        r.SpineRuntimeConfigMinAnimationSpeed,
		MaxAnimationSpeed:        r.SpineRuntimeConfigMaxAnimationSpeed,
		DefaultScaleSize:         r.SpineRuntimeConfigDefaultScaleSize,
		MinScaleSize:             r.SpineRuntimeConfigMinScaleSize,
		MaxScaleSize:             r.SpineRuntimeConfigMaxScaleSize,
		MaxSpritePixelSize:       r.SpineRuntimeConfigMaxSpritePixelSize,
		ReferenceMovementSpeedPx: r.SpineRuntimeConfigReferenceMovementSpeedPx,
		DefaultMovementSpeed:     r.SpineRuntimeConfigDefaultMovementSpeed,
		MinMovementSpeed:         r.SpineRuntimeConfigMinMovementSpeed,
		MaxMovementSpeed:         r.SpineRuntimeConfigMaxMovementSpeed,
	}
}

func (r *RoomDb) SetSpineRuntimeConfig(config *misc.SpineRuntimeConfig) {
	r.SpineRuntimeConfigDefaultAnimationSpeed = config.DefaultAnimationSpeed
	r.SpineRuntimeConfigMinAnimationSpeed = config.MinAnimationSpeed
	r.SpineRuntimeConfigMaxAnimationSpeed = config.MaxAnimationSpeed

	r.SpineRuntimeConfigDefaultScaleSize = config.DefaultScaleSize
	r.SpineRuntimeConfigMinScaleSize = config.MinScaleSize
	r.SpineRuntimeConfigMaxScaleSize = config.MaxScaleSize
	r.SpineRuntimeConfigMaxSpritePixelSize = config.MaxSpritePixelSize

	r.SpineRuntimeConfigReferenceMovementSpeedPx = config.ReferenceMovementSpeedPx
	r.SpineRuntimeConfigDefaultMovementSpeed = config.DefaultMovementSpeed
	r.SpineRuntimeConfigMinMovementSpeed = config.MinMovementSpeed
	r.SpineRuntimeConfigMaxMovementSpeed = config.MaxMovementSpeed
}

func GetOrInsertRoom(roomConfig *RoomConfig) (*RoomDb, error) {
	db, err := akdb.Connect()
	if err != nil {
		return nil, err
	}

	var roomDb RoomDb
	tx := db.Where("channel_name = ?", roomConfig.ChannelName).First(&roomDb)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {

			roomDb.ChannelName = roomConfig.ChannelName
			roomDb.IsActive = true

			// This should get updated for each new server load.
			roomDb.DefaultOperatorName = roomConfig.DefaultOperatorName
			roomDb.DefaultOperatorSkin = roomConfig.DefaultOperatorConfig.Skin
			roomDb.DefaultOperatorStance = roomConfig.DefaultOperatorConfig.Stance
			roomDb.DefaultOperatorPositionX = roomConfig.DefaultOperatorConfig.PositionX
			roomDb.DefaultOperatorAnimations = roomConfig.DefaultOperatorConfig.Animations

			roomDb.GarbageCollectionPeriodMins = roomConfig.GarbageCollectionPeriodMins

			roomDb.SpineRuntimeConfigDefaultAnimationSpeed = roomConfig.SpineRuntimeConfig.DefaultAnimationSpeed
			roomDb.SpineRuntimeConfigMinAnimationSpeed = roomConfig.SpineRuntimeConfig.MinAnimationSpeed
			roomDb.SpineRuntimeConfigMaxAnimationSpeed = roomConfig.SpineRuntimeConfig.MaxAnimationSpeed
			roomDb.SpineRuntimeConfigDefaultScaleSize = roomConfig.SpineRuntimeConfig.DefaultScaleSize
			roomDb.SpineRuntimeConfigMinScaleSize = roomConfig.SpineRuntimeConfig.MinScaleSize
			roomDb.SpineRuntimeConfigMaxScaleSize = roomConfig.SpineRuntimeConfig.MaxScaleSize
			roomDb.SpineRuntimeConfigMaxSpritePixelSize = roomConfig.SpineRuntimeConfig.MaxSpritePixelSize
			roomDb.SpineRuntimeConfigReferenceMovementSpeedPx = roomConfig.SpineRuntimeConfig.ReferenceMovementSpeedPx
			roomDb.SpineRuntimeConfigDefaultMovementSpeed = roomConfig.SpineRuntimeConfig.DefaultMovementSpeed
			roomDb.SpineRuntimeConfigMinMovementSpeed = roomConfig.SpineRuntimeConfig.MinMovementSpeed
			roomDb.SpineRuntimeConfigMaxMovementSpeed = roomConfig.SpineRuntimeConfig.MaxMovementSpeed

			tx := db.Create(&roomDb)
			if tx.Error != nil {
				return nil, tx.Error
			}
			return &roomDb, nil
		} else {
			return nil, tx.Error
		}
	} else {
		return &roomDb, nil
	}
}

func UpdateRoom(roomDb *RoomDb) error {
	db, err := akdb.Connect()
	if err != nil {
		return err
	}
	tx := db.Save(roomDb)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func GetActiveRooms() ([]*RoomDb, error) {
	db, err := akdb.Connect()
	if err != nil {
		return nil, err
	}
	var roomDbs []*RoomDb
	tx := db.Where("is_active = true").Find(&roomDbs)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return roomDbs, nil
}
