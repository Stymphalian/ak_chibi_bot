package room

import (
	"context"
	"log"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type RoomRepositoryPsql struct {
}

func NewRoomRepositoryPsql() *RoomRepositoryPsql {
	return &RoomRepositoryPsql{}
}

func (r *RoomRepositoryPsql) GetOrInsertRoom(ctx context.Context, roomConfig *RoomConfig) (*RoomDb, bool, error) {
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

func (r *RoomRepositoryPsql) GetActiveRooms(ctx context.Context) ([]*RoomDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDbs []*RoomDb
	tx := db.Where("is_active = true").Find(&roomDbs)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return roomDbs, nil
}

func (r *RoomRepositoryPsql) GetRoomByChannelName(ctx context.Context, channelName string) (*RoomDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDb RoomDb
	result := db.Where("channel_name = ?", channelName).First(&roomDb)
	if result.Error != nil {
		return nil, result.Error
	}
	return &roomDb, nil
}

func (r *RoomRepositoryPsql) GetRoomGarbageCollectionPeriodMins(ctx context.Context, roomId uint) int {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDb RoomDb
	result := db.First(&roomDb, roomId)
	if result.Error != nil {
		return 0
	}
	return roomDb.GarbageCollectionPeriodMins
}

func (r *RoomRepositoryPsql) GetSpineRuntimeConfigById(ctx context.Context, roomId uint) (*misc.SpineRuntimeConfig, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDb RoomDb
	result := db.First(&roomDb, roomId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &roomDb.SpineRuntimeConfig, nil
}

func (r *RoomRepositoryPsql) UpdateSpineRuntimeConfigForId(
	ctx context.Context,
	roomId uint,
	config *misc.SpineRuntimeConfig,
) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.
		Model(&RoomDb{}).
		Where("room_id = ?", roomId).
		Select("spine_runtime_config").
		Updates(&RoomDb{SpineRuntimeConfig: *config})
	if result.Error != nil {
		log.Println("Error updating room ", roomId, result.Error)
	}
	return result.Error
}

func (r *RoomRepositoryPsql) IsRoomActiveById(ctx context.Context, roomId uint) bool {
	db := akdb.DefaultDB.WithContext(ctx)
	var roomDb RoomDb
	result := db.First(&roomDb, roomId)
	if result.Error != nil {
		return false
	}
	return roomDb.IsActive
}

func (r *RoomRepositoryPsql) SetRoomActiveById(ctx context.Context, roomId uint, isActive bool) error {
	db := akdb.DefaultDB.WithContext(ctx)
	// roomDb := &RoomDb{RoomId: roomId, IsActive: isActive}
	result := db.
		Model(&RoomDb{}).
		Where("room_id = ?", roomId).
		Select("is_active").
		Updates(&RoomDb{IsActive: isActive})
	if result.Error != nil {
		log.Println("Error updating room ", roomId, result.Error)
	}
	return result.Error
}
