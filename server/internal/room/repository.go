package room

import (
	"context"
	"log"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"gorm.io/gorm"
)

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

func (r *RoomDb) GetRoomId() uint {
	return r.RoomId
}

func (r *RoomDb) GetChannelName() string {
	return r.ChannelName
}

func (r *RoomDb) SetChannelName(ctx context.Context, channelName string) error {
	r.ChannelName = channelName
	return r.Update(ctx, "channel_name")
}

func (r *RoomDb) GetIsActive() bool {
	r.Refresh(context.Background(), "is_active")
	return r.IsActive
}

func (r *RoomDb) SetIsActive(ctx context.Context, isActive bool) error {
	r.IsActive = isActive
	return r.Update(ctx, "is_active")
}

func (r *RoomDb) GetDefaultOperatorName() string {
	r.Refresh(context.Background(), "default_operator_name")
	return r.DefaultOperatorName
}

func (r *RoomDb) SetDefaultOperatorName(
	ctx context.Context,
	defaultOperatorName string,
) error {
	r.DefaultOperatorName = defaultOperatorName
	return r.Update(ctx, "default_operator_name")
}

func (r *RoomDb) GetDefaultOperatorConfig() *misc.InitialOperatorDetails {
	r.Refresh(context.Background(), "default_operator_config")
	return &r.DefaultOperatorConfig
}

func (r *RoomDb) SetDefaultOperatorConfig(
	ctx context.Context,
	defaultOperatorConfig *misc.InitialOperatorDetails,
) error {
	r.DefaultOperatorConfig = *defaultOperatorConfig
	return r.Update(ctx, "default_operator_config")
}

func (r *RoomDb) GetSpineRuntimeConfig() *misc.SpineRuntimeConfig {
	r.Refresh(context.Background(), "spine_runtime_config")
	return &r.SpineRuntimeConfig
}

func (r *RoomDb) SetSpineRuntimeConfig(
	ctx context.Context,
	spineRuntimeConfig *misc.SpineRuntimeConfig,
) error {
	r.SpineRuntimeConfig = *spineRuntimeConfig
	return r.Update(ctx, "spine_runtime_config")
}

func (r *RoomDb) GetGarbageCollectionPeriodMins() int {
	r.Refresh(context.Background(), "garbage_collection_period_mins")
	return r.GarbageCollectionPeriodMins
}

func (r *RoomDb) SetGarbageCollectionPeriodMins(
	ctx context.Context,
	garbageCollectionPeriodMins int,
) error {
	r.GarbageCollectionPeriodMins = garbageCollectionPeriodMins
	return r.Update(ctx, "garbage_collection_period_mins")
}

func (RoomDb) TableName() string {
	return "rooms"
}

func (r *RoomDb) Refresh(ctx context.Context, fields ...string) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.Where("room_id = ?", r.RoomId).Select(fields).First(r)
	if result.Error != nil {
		log.Println("Error refreshing room", r.RoomId, result.Error)
	}
	return result.Error
}

func (r *RoomDb) Update(ctx context.Context, args ...string) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.Model(r).Where("room_id = ?", r.RoomId).Select(args).Updates(*r)
	if result.Error != nil {
		log.Println("Error updating room ", r.RoomId, result.Error)
	}
	return result.Error
}

func GetRoomByKey(ctx context.Context, roomId uint) (*RoomDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDb RoomDb
	result := db.First(&roomDb, roomId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &roomDb, nil
}

func GetRoomByChannelName(ctx context.Context, channel string) (*RoomDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDb RoomDb
	result := db.Where("channel_name = ?", channel).First(&roomDb)
	if result.Error != nil {
		return nil, result.Error
	}
	return &roomDb, nil
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
