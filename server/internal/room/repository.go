package room

import (
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

func GetOrInsertRoom(roomConfig *RoomConfig) (*RoomDb, bool, error) {
	db, err := akdb.Connect()
	if err != nil {
		return nil, false, err
	}

	var roomDb RoomDb
	tx := db.Where("channel_name = ?", roomConfig.ChannelName).First(&roomDb)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {

			roomDb.ChannelName = roomConfig.ChannelName
			roomDb.IsActive = true

			// TODO: default oeprator config should get updated for each new server reloading
			// so that the new operators with each banner release can be the default op.
			roomDb.DefaultOperatorName = roomConfig.DefaultOperatorName
			roomDb.DefaultOperatorConfig = roomConfig.DefaultOperatorConfig
			roomDb.SpineRuntimeConfig = *roomConfig.SpineRuntimeConfig
			roomDb.GarbageCollectionPeriodMins = roomConfig.GarbageCollectionPeriodMins

			tx := db.Create(&roomDb)
			if tx.Error != nil {
				return nil, false, tx.Error
			}
			return &roomDb, true, nil
		} else {
			return nil, false, tx.Error
		}
	} else {
		return &roomDb, false, nil
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
