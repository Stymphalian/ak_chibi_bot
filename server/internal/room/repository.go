package room

import (
	"context"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"gorm.io/gorm"
)

type RoomRepository interface {
	GetActiveRooms(ctx context.Context) ([]*RoomDb, error)
	GetOrInsertRoom(ctx context.Context, roomConfig *RoomConfig) (
		roomDb *RoomDb, isNew bool, err error)

	GetRoomGarbageCollectionPeriodMins(ctx context.Context, roomId uint) int

	GetSpineRuntimeConfigById(ctx context.Context, roomId uint) (
		*misc.SpineRuntimeConfig, error)
	UpdateSpineRuntimeConfigForId(
		ctx context.Context,
		roomId uint,
		config *misc.SpineRuntimeConfig,
	) error

	IsRoomActiveById(ctx context.Context, roomId uint) bool
	SetRoomActiveById(ctx context.Context, roomId uint, isActive bool) error
}

type RoomDb struct {
	RoomId                      uint                        `gorm:"primarykey"`
	ChannelName                 string                      `gorm:"column:channel_name"`
	IsActive                    bool                        `gorm:"column:is_active"`
	DefaultOperatorName         string                      `gorm:"column:default_operator_name"`
	DefaultOperatorConfig       misc.InitialOperatorDetails `gorm:"column:default_operator_config;type:json"`
	SpineRuntimeConfig          misc.SpineRuntimeConfig     `gorm:"column:spine_runtime_config;type:json"`
	GarbageCollectionPeriodMins int                         `gorm:"column:garbage_collection_period_mins"`
	CreatedAt                   time.Time                   `gorm:"column:created_at"`
	UpdatedAt                   time.Time                   `gorm:"column:updated_at"`
	DeletedAt                   gorm.DeletedAt              `gorm:"index"`
}

func (RoomDb) TableName() string {
	return "rooms"
}
