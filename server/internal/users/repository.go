package users

import (
	"errors"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"

	"gorm.io/gorm"
)

type UserDb struct {
	UserId          uint           `gorm:"primarykey"`
	Username        string         `gorm:"column:username"`
	UserDisplayName string         `gorm:"column:user_display_name"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (UserDb) TableName() string {
	return "users"
}

func GetUserById(userId uint) (*UserDb, error) {
	db := akdb.DefaultDB

	var userDb UserDb
	result := db.First(&userDb, userId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &userDb, nil
}

func GetOrInsertUser(username string, userDisplayName string) (*UserDb, error) {
	db := akdb.DefaultDB

	var userDb UserDb
	result := db.Where("username = ?", username).First(&userDb)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Row doesn't exist yet, so create the user
			userDb.Username = username
			userDb.UserDisplayName = username
			result = db.Create(&userDb)
			if result.Error != nil {
				return nil, result.Error
			}
			return &userDb, nil
		} else {
			return nil, result.Error
		}
	} else {
		return &userDb, nil
	}
}

func UpdateUser(user *UserDb) error {
	db := akdb.DefaultDB
	return db.Save(user).Error
}

type ChatterDb struct {
	ChatterId    uint               `gorm:"primarykey"`
	RoomId       uint               `gorm:"column:room_id"`
	UserId       uint               `gorm:"column:user_id"`
	IsActive     bool               `gorm:"column:is_active"`
	OperatorInfo spine.OperatorInfo `gorm:"operator_info;type:json"`
	LastChatTime time.Time          `gorm:"column:last_chat_time"`

	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// User         *UserDb            `gorm:"foreignKey:UserId;references:UserId"`
}

func (ChatterDb) TableName() string {
	return "chatters"
}

func GetOrInsertChatter(roomId uint, user *UserDb, lastChatTime time.Time, operatorInfo *spine.OperatorInfo) (*ChatterDb, error) {
	// func GetOrInsertChatter(roomId uint, user *UserDb) (*ChatterDb, error) {
	db := akdb.DefaultDB

	var chatterDb ChatterDb
	result := db.Where("room_id = ? AND user_id = ?", roomId, user.UserId).First(&chatterDb)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Row doesn't exist yet, so create the user
			chatterDb.RoomId = roomId
			chatterDb.UserId = user.UserId
			chatterDb.IsActive = true
			chatterDb.OperatorInfo = *operatorInfo
			chatterDb.LastChatTime = lastChatTime

			result = db.Create(&chatterDb)
			if result.Error != nil {
				return nil, result.Error
			}
			return &chatterDb, nil
		} else {
			return nil, result.Error
		}
	} else {
		return &chatterDb, nil
	}
}

func UpdateChatter(c *ChatterDb) error {
	db := akdb.DefaultDB
	return db.Save(c).Error
}

func GetActiveChatters(roomId uint) ([]*ChatterDb, error) {
	db := akdb.DefaultDB
	var chatterDbs []*ChatterDb
	tx := db.Where("room_id = ? AND is_active = true", roomId).Find(&chatterDbs)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return chatterDbs, nil
}
