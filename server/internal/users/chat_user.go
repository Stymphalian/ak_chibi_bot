package users

import (
	"context"
	"log"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

type ChatUser struct {
	user    *UserDb
	chatter *ChatterDb

	userId      uint
	chatterId   uint
	usersRepo   UserRepository
	chatterRepo ChatterRepository
}

func NewChatUser(userRepo UserRepository, chatterRepo ChatterRepository, userId uint, chatterId uint) (*ChatUser, error) {
	userDb, err := userRepo.GetById(context.Background(), userId)
	if err != nil {
		return nil, err
	}
	chatterDb, err := chatterRepo.GetById(context.Background(), chatterId)
	if err != nil {
		return nil, err
	}

	return &ChatUser{
		user:        userDb,
		chatter:     chatterDb,
		usersRepo:   userRepo,
		chatterRepo: chatterRepo,
		userId:      userId,
		chatterId:   chatterId,
	}, nil
}

func (c *ChatUser) Close() (err error) {
	err = c.chatterRepo.SetActiveById(context.Background(), c.chatterId, false)
	if err != nil {
		return
	}
	c.chatter.IsActive = false
	return
}

func (c *ChatUser) IsActiveChatter(period time.Duration) bool {
	return misc.Clock.Since(c.GetLastChatTime()) < period
}

func (c *ChatUser) GetUsername() string {
	return c.user.Username
}

func (c *ChatUser) GetUsernameDisplay() string {
	return c.user.UserDisplayName
}

func (c *ChatUser) GetOperatorInfo() *operator.OperatorInfo {
	return &c.chatter.OperatorInfo
}

func (c *ChatUser) SetOperatorInfo(v *operator.OperatorInfo) (err error) {
	err = c.chatterRepo.SetOperatorInfoById(context.Background(), c.chatterId, v)
	if err != nil {
		return
	}
	c.chatter.OperatorInfo = *v
	return
}

func (c *ChatUser) GetLastChatTime() time.Time {
	return c.chatter.LastChatTime
}

func (c *ChatUser) SetLastChatTime(v time.Time) (err error) {
	err = c.chatterRepo.SetLastChatTimeById(context.Background(), c.chatterId, v)
	if err != nil {
		log.Printf("failed to set last chat time for chatter %d: %s\n", c.chatterId, err)
		return
	}
	c.chatter.LastChatTime = v
	return
}

func (c *ChatUser) SetActive(isActive bool) (err error) {
	err = c.chatterRepo.SetActiveById(context.Background(), c.chatterId, isActive)
	if err != nil {
		return
	}
	c.chatter.IsActive = isActive
	return
}

func (c *ChatUser) IsActive() bool {
	return c.chatter.IsActive
}

func (c *ChatUser) UpdateWithLatestChat(update *operator.OperatorInfo) (err error) {
	lastChatTime := misc.Clock.Now()
	err = c.chatterRepo.UpdateLatestChat(
		context.Background(),
		c.chatterId,
		update,
		lastChatTime,
	)
	if err != nil {
		return err
	}
	c.chatter.OperatorInfo = *update
	c.chatter.LastChatTime = lastChatTime
	c.chatter.IsActive = true
	return
}
