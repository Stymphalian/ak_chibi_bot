package users

import (
	"context"
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

func GetUserById(ctx context.Context, userId uint) (*UserDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var userDb UserDb
	result := db.First(&userDb, userId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &userDb, nil
}

func GetOrInsertUser(ctx context.Context, username string, userDisplayName string) (*UserDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var userDb UserDb
	result := db.
		Where("username = ?", username).
		Attrs(
			UserDb{
				Username:        username,
				UserDisplayName: userDisplayName,
			},
		).
		FirstOrCreate(&userDb)

	if result.Error != nil {
		return nil, result.Error
	} else {
		return &userDb, nil
	}
}

func UpdateUser(ctx context.Context, user *UserDb) error {
	db := akdb.DefaultDB.WithContext(ctx)
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
}

func (ChatterDb) TableName() string {
	return "chatters"
}

func GetOrInsertChatter(ctx context.Context, roomId uint, user *UserDb, lastChatTime time.Time, operatorInfo *spine.OperatorInfo) (*ChatterDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var chatterDb ChatterDb
	result := db.
		Where("room_id = ? AND user_id = ?", roomId, user.UserId).
		Attrs(
			ChatterDb{
				RoomId:       roomId,
				UserId:       user.UserId,
				IsActive:     true,
				OperatorInfo: *operatorInfo,
				LastChatTime: lastChatTime,
			},
		).
		FirstOrCreate(&chatterDb)

	if result.Error != nil {
		return nil, result.Error
	} else {
		return &chatterDb, nil
	}
}

func UpdateChatter(ctx context.Context, c *ChatterDb) error {
	db := akdb.DefaultDB.WithContext(ctx)
	return db.Save(c).Error
}

type UserChatterDb struct {
	ChatterDb
	User UserDb
}

func (UserChatterDb) TableName() string {
	return "chatters"
}

func GetActiveChatters(ctx context.Context, roomId uint) ([]*UserChatterDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var chatUsers []*UserChatterDb
	tx := db.
		Where("room_id = ? AND is_active = true", roomId).
		Preload("User").
		Find(&chatUsers)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return chatUsers, nil
}
