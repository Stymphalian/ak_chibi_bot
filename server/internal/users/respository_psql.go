package users

import (
	"context"
	"errors"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"gorm.io/gorm"
)

type UserRepositoryPsql struct {
}

func NewUserRepositoryPsql() *UserRepositoryPsql {
	return &UserRepositoryPsql{}
}

func (r *UserRepositoryPsql) GetById(ctx context.Context, userId uint) (*UserDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var userDb UserDb
	result := db.First(&userDb, userId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &userDb, nil
}

func (r *UserRepositoryPsql) GetByTwitchId(ctx context.Context, twitchUserId string) (*UserDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var userDb UserDb
	result := db.First(&userDb, "twitch_user_id = ?", twitchUserId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &userDb, nil
}

func (r *UserRepositoryPsql) GetOrInsertUser(ctx context.Context, userinfo misc.UserInfo) (*UserDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	// TODO: Should check via twitchUserId instead of just username
	var userDb UserDb
	result := db.
		Where("username = ?", userinfo.Username).
		Attrs(
			UserDb{
				Username:        userinfo.Username,
				UserDisplayName: userinfo.UsernameDisplay,
				TwitchUserId:    userinfo.TwitchUserId,
			},
		).
		FirstOrCreate(&userDb)

	if result.Error != nil {
		return nil, result.Error
	} else {
		return &userDb, nil
	}
}

type ChatterRepositoryPsql struct {
}

func NewChatterRepositoryPsql() *ChatterRepositoryPsql {
	return &ChatterRepositoryPsql{}
}

func (r *ChatterRepositoryPsql) GetById(ctx context.Context, chatterId uint) (*ChatterDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var chatterDb ChatterDb
	result := db.First(&chatterDb, chatterId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chatterDb, nil
}

func (r *ChatterRepositoryPsql) GetOrInsertChatter(
	ctx context.Context,
	roomId uint,
	user *UserDb,
	lastChatTime time.Time,
	operatorInfo *operator.OperatorInfo,
) (*ChatterDb, error) {
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

func (r *ChatterRepositoryPsql) GetActiveChatters(ctx context.Context, roomId uint) ([]*UserChatterDb, error) {
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

func (r *ChatterRepositoryPsql) SetOperatorInfoById(ctx context.Context, chatterId uint, operatorInfo *operator.OperatorInfo) error {
	db := akdb.DefaultDB.WithContext(ctx)

	result := db.
		Model(&ChatterDb{}).
		Where("chatter_id = ?", chatterId).
		Update("operator_info", operatorInfo)
	return result.Error
}

func (r *ChatterRepositoryPsql) SetLastChatTimeById(ctx context.Context, chatterId uint, lastChatTime time.Time) error {
	db := akdb.DefaultDB.WithContext(ctx)

	result := db.
		Model(&ChatterDb{}).
		Where("chatter_id = ?", chatterId).
		Update("last_chat_time", lastChatTime)
	return result.Error
}

func (r *ChatterRepositoryPsql) SetActiveById(ctx context.Context, chatterId uint, isActive bool) error {
	db := akdb.DefaultDB.WithContext(ctx)

	result := db.
		Model(&ChatterDb{}).
		Where("chatter_id = ?", chatterId).
		Update("is_active", isActive)
	return result.Error
}

func (r *ChatterRepositoryPsql) GetOperatorInfoById(ctx context.Context, chatterId uint) (*operator.OperatorInfo, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var chatterDb ChatterDb
	result := db.First(&chatterDb, chatterId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chatterDb.OperatorInfo, nil
}

func (r *ChatterRepositoryPsql) GetLastChatTimeById(ctx context.Context, chatterId uint) (time.Time, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var chatterDb ChatterDb
	result := db.First(&chatterDb, chatterId)
	if result.Error != nil {
		return time.Time{}, result.Error
	}
	return chatterDb.LastChatTime, nil
}

func (r *ChatterRepositoryPsql) GetIsActiveById(ctx context.Context, chatterId uint) (bool, error) {
	db := akdb.DefaultDB.WithContext(ctx)
	var chatterDb ChatterDb
	result := db.First(&chatterDb, chatterId)
	if result.Error != nil {
		return false, result.Error
	}
	return chatterDb.IsActive, nil
}

func (r *ChatterRepositoryPsql) UpdateLatestChat(ctx context.Context, chatterId uint, opInfo *operator.OperatorInfo, lastChatTime time.Time) error {
	db := akdb.DefaultDB.WithContext(ctx)
	result := db.
		Model(&ChatterDb{}).
		Where("chatter_id = ?", chatterId).
		Select("operator_info", "last_chat_time", "is_active").
		Updates(&ChatterDb{
			OperatorInfo: *opInfo,
			IsActive:     true,
			LastChatTime: lastChatTime,
		})
	return result.Error
}

type UserPreferencesRepositoryPsql struct{}

func NewUserPreferencesRepositoryPsql() *UserPreferencesRepositoryPsql {
	return &UserPreferencesRepositoryPsql{}
}

func (r *UserPreferencesRepositoryPsql) GetByUserIdOrNil(ctx context.Context, userId uint) (*UserPreferencesDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var userDb UserPreferencesDb
	result := db.
		Where("user_id = ?", userId).
		First(&userDb)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &userDb, nil
}

func (r *UserPreferencesRepositoryPsql) GetByTwitchIdOrNil(ctx context.Context, twitchUserId string) (*UserPreferencesDb, error) {
	db := akdb.DefaultDB.WithContext(ctx)

	var prefDb UserPreferencesDb
	result := db.
		Table("user_preferences as p").
		Select("p.*").
		Joins("LEFT JOIN users as u ON u.user_id = p.user_id").
		Where("u.twitch_user_id = ?", twitchUserId).
		First(&prefDb)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &prefDb, nil
}

func (r *UserPreferencesRepositoryPsql) SetByUserId(ctx context.Context, userId uint, opInfo *operator.OperatorInfo) error {
	db := akdb.DefaultDB.WithContext(ctx)

	var prefDb UserPreferencesDb
	result := db.
		Where("user_id = ?", userId).
		Assign(
			UserPreferencesDb{
				UserId:       userId,
				OperatorInfo: *opInfo,
			},
		).
		FirstOrCreate(&prefDb)
	return result.Error
}

func (r *UserPreferencesRepositoryPsql) DeleteByUserId(ctx context.Context, userId uint) error {
	db := akdb.DefaultDB.WithContext(ctx)

	result := db.
		Where("user_id = ?", userId).
		Delete(&UserPreferencesDb{})
	return result.Error
}
