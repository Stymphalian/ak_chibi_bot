package users

import (
	"context"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

type ChatUser struct {
	user    *UserDb
	chatter *ChatterDb
}

func NewChatUser(User *UserDb, Chatter *ChatterDb) (*ChatUser, error) {
	return &ChatUser{
		user:    User,
		chatter: Chatter,
	}, nil
}

func (c *ChatUser) Close() error {
	return c.chatter.SetIsActive(context.Background(), false)
}

func (c *ChatUser) IsActiveChatter(period time.Duration) bool {
	return misc.Clock.Since(c.chatter.GetLastChatTime()) < period
}

func (c *ChatUser) GetUsername() string {
	return c.user.Username
}

func (c *ChatUser) GetUsernameDisplay() string {
	return c.user.UserDisplayName
}

func (c *ChatUser) GetOperatorInfo() *operator.OperatorInfo {
	return c.chatter.GetOperatorInfo()
}

func (c *ChatUser) SetOperatorInfo(v *operator.OperatorInfo) error {
	return c.chatter.SetOperatorInfo(context.Background(), v)
}

func (c *ChatUser) GetLastChatTime() time.Time {
	return c.chatter.GetLastChatTime()
}

func (c *ChatUser) SetLastChatTime(v time.Time) {
	c.chatter.SetLastChatTime(context.Background(), v)
}

func (c *ChatUser) SetActive(isActive bool) {
	c.chatter.SetIsActive(context.Background(), isActive)
}

func (c *ChatUser) IsActive() bool {
	return c.chatter.GetIsActive()
}
