package users

import (
	"context"
	"log"
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

func (r *UserDb) GetUserId() uint {
	return r.UserId
}

func (r *UserDb) GetUsername(ctx context.Context) string {
	r.Refresh(ctx, "username")
	return r.Username
}

func (r *UserDb) SetUsername(ctx context.Context, username string) error {
	r.Username = username
	return r.Update(ctx, "username")
}

func (r *UserDb) GetUserDisplayName(ctx context.Context) string {
	r.Refresh(ctx, "user_display_name")
	return r.UserDisplayName
}

func (r *UserDb) SetUserDisplayName(ctx context.Context, userDisplayName string) error {
	r.UserDisplayName = userDisplayName
	return r.Update(ctx, "user_display_name")
}

func (UserDb) TableName() string {
	return "users"
}

func (r *UserDb) Refresh(ctx context.Context, fields ...string) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.Where("user_id = ?", r.UserId).Select(fields).First(r)
	if result.Error != nil {
		log.Println("Error refreshing UserDb", r.UserId, result.Error)
	}
	return result.Error
}

func (r *UserDb) Update(ctx context.Context, args ...string) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.Model(r).Where("user_id = ?", r.UserId).Select(args).Updates(*r)
	if result.Error != nil {
		log.Println("Error updating UserDb ", r.UserId, result.Error)
	}
	return result.Error
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

func (r *ChatterDb) GetRoomId() uint {
	r.Refresh(context.Background(), "room_id")
	return r.RoomId
}

func (r *ChatterDb) SetRoomId(ctx context.Context, roomId uint) error {
	r.RoomId = roomId
	return r.Update(ctx, "room_id")
}

func (r *ChatterDb) GetUserId() uint {
	r.Refresh(context.Background(), "user_id")
	return r.UserId
}

func (r *ChatterDb) SetUserId(ctx context.Context, userId uint) error {
	r.UserId = userId
	return r.Update(ctx, "user_id")
}

func (r *ChatterDb) GetIsActive() bool {
	r.Refresh(context.Background(), "is_active")
	return r.IsActive
}

func (r *ChatterDb) SetIsActive(ctx context.Context, isActive bool) error {
	r.IsActive = isActive
	return r.Update(ctx, "is_active")
}

func (r *ChatterDb) GetOperatorInfo() *spine.OperatorInfo {
	r.Refresh(context.Background(), "operator_info")
	return &r.OperatorInfo
}

func (r *ChatterDb) SetOperatorInfo(ctx context.Context, opInfo *spine.OperatorInfo) error {
	r.OperatorInfo = *opInfo
	return r.Update(ctx, "operator_info")
}

func (r *ChatterDb) GetLastChatTime() time.Time {
	r.Refresh(context.Background(), "last_chat_time")
	return r.LastChatTime
}

func (r *ChatterDb) SetLastChatTime(ctx context.Context, lastChatTime time.Time) error {
	r.LastChatTime = lastChatTime
	return r.Update(ctx, "last_chat_time")
}

func (r *ChatterDb) Refresh(ctx context.Context, fields ...string) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.Where("chatter_id = ?", r.ChatterId).Select(fields).First(r)
	if result.Error != nil {
		log.Println("Error refreshing ChatterDb", r.ChatterId, result.Error)
	}
	return result.Error
}

func (r *ChatterDb) Update(ctx context.Context, args ...string) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.Model(r).Where("chatter_id = ?", r.ChatterId).Select(args).Updates(*r)
	if result.Error != nil {
		log.Println("Error updating ChatterDb ", r.ChatterId, result.Error)
	}
	return result.Error
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
