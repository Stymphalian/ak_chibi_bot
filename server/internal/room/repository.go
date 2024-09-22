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

// func refresh(ctx context.Context, fields ...string) error {
// 	db := akdb.DefaultDB.WithContext(ctx)
// 	result := db.Where("room_id = ?", r.RoomId).Select(fields).First(r)
// 	if result.Error != nil {
// 		log.Println("Error refreshing room", r.RoomId, result.Error)
// 	}
// 	return result.Error
// }

// func update(ctx context.Context, args ...string) error {
// 	db := akdb.DefaultDB.WithContext(ctx)
// 	result := db.Model(r).Where("room_id = ?", r.RoomId).Select(args).Updates(*r)
// 	if result.Error != nil {
// 		log.Println("Error updating room ", r.RoomId, result.Error)
// 	}
// 	return result.Error
// }

// func GetRoomByKey(ctx context.Context, roomId uint) (*RoomDb, error) {
// 	db := akdb.DefaultDB.WithContext(ctx)
// 	var roomDb RoomDb
// 	result := db.First(&roomDb, roomId)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}
// 	return &roomDb, nil
// }

// func GetRoomByChannelName(ctx context.Context, channel string) (*RoomDb, error) {
// 	db := akdb.DefaultDB.WithContext(ctx)
// 	var roomDb RoomDb
// 	result := db.Where("channel_name = ?", channel).First(&roomDb)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}
// 	return &roomDb, nil
// }

// func UpdateRoom(ctx context.Context, roomDb *RoomDb) error {
// 	db := akdb.DefaultDB.WithContext(ctx)
// 	tx := db.Save(roomDb)
// 	if tx.Error != nil {
// 		return tx.Error
// 	}
// 	return nil
// }
