package users

import (
	"context"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetById(ctx context.Context, userId uint) (*UserDb, error)
	GetOrInsertUser(ctx context.Context, username string, userDisplayName string) (*UserDb, error)
}

type ChatterRepository interface {
	GetById(ctx context.Context, chatterId uint) (*ChatterDb, error)
	GetOrInsertChatter(
		ctx context.Context,
		roomId uint,
		user *UserDb,
		lastChatTime time.Time,
		operatorInfo *operator.OperatorInfo,
	) (*ChatterDb, error)

	SetOperatorInfoById(ctx context.Context, chatterId uint, operatorInfo *operator.OperatorInfo) error
	SetLastChatTimeById(ctx context.Context, chatterId uint, lastChatTime time.Time) error
	SetActiveById(ctx context.Context, chatterId uint, isActive bool) error
	UpdateLatestChat(ctx context.Context, chatterId uint, opInfo *operator.OperatorInfo, lastChatTime time.Time) error
	GetActiveChatters(ctx context.Context, roomId uint) ([]*UserChatterDb, error)
}

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

type ChatterDb struct {
	ChatterId    uint                  `gorm:"primarykey"`
	RoomId       uint                  `gorm:"column:room_id"`
	UserId       uint                  `gorm:"column:user_id"`
	IsActive     bool                  `gorm:"column:is_active"`
	OperatorInfo operator.OperatorInfo `gorm:"operator_info;type:json"`
	LastChatTime time.Time             `gorm:"column:last_chat_time"`

	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ChatterDb) TableName() string {
	return "chatters"
}

type UserChatterDb struct {
	ChatterDb
	User UserDb
}

func (UserChatterDb) TableName() string {
	return "chatters"
}
