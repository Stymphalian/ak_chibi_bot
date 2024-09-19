package room

import (
	"context"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"gorm.io/gorm"
)

// TODO: Fuck this. Too many columns
type RoomDb struct {
	RoomId                      uint                        `gorm:"primarykey"`
	ChannelName                 string                      `gorm:"column:channel_name"`
	IsActive                    bool                        `gorm:"column:is_active"`
	DefaultOperatorName         string                      `gorm:"column:default_operator_name"`
	DefaultOperatorConfig       misc.InitialOperatorDetails `gorm:"column:default_operator_config;type:json"`
	SpineRuntimeConfig          misc.SpineRuntimeConfig     `gorm:"column:spine_runtime_config;type:json"`
	GarbageCollectionPeriodMins int                         `gorm:"column:garbage_collection_period_mins"`

	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RoomDb) TableName() string {
	return "rooms"
}

func (r *RoomDb) SetSpineRuntimeConfig(config *misc.SpineRuntimeConfig) {
	r.SpineRuntimeConfig = *config
}

func GetOrInsertRoom(ctx context.Context, roomConfig *RoomConfig) (*RoomDb, bool, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var roomDb RoomDb
	result := db.Where("channel_name = ?", roomConfig.ChannelName).Attrs(
		RoomDb{
			ChannelName:                 roomConfig.ChannelName,
			IsActive:                    true,
			DefaultOperatorName:         roomConfig.DefaultOperatorName,
			DefaultOperatorConfig:       roomConfig.DefaultOperatorConfig,
			SpineRuntimeConfig:          *roomConfig.SpineRuntimeConfig,
			GarbageCollectionPeriodMins: roomConfig.GarbageCollectionPeriodMins,
		},
	).FirstOrCreate(&roomDb)
	if result.Error != nil {
		return nil, false, result.Error
	}
	if result.RowsAffected == 0 {
		return &roomDb, false, nil
	} else {
		return &roomDb, true, nil
	}
}

func UpdateRoom(ctx context.Context, roomDb *RoomDb) error {
	db := akdb.DefaultDB.WithContext(ctx)
	tx := db.Save(roomDb)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func GetActiveRooms(ctx context.Context) ([]*RoomDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDbs []*RoomDb
	tx := db.Where("is_active = true").Find(&roomDbs)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return roomDbs, nil
}
