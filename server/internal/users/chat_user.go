package users

import (
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
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
	c.chatter.IsActive = false
	return c.Save()
}

func (c *ChatUser) IsActiveChatter(period time.Duration) bool {
	return misc.Clock.Since(c.chatter.LastChatTime) < period
}

func (c *ChatUser) GetUsername() string {
	return c.user.Username
}
func (c *ChatUser) GetUsernameDisplay() string {
	return c.user.UserDisplayName
}

func (c *ChatUser) GetOperatorInfo() spine.OperatorInfo {
	return c.chatter.OperatorInfo
}
func (c *ChatUser) SetOperatorInfo(v *spine.OperatorInfo) {
	c.chatter.OperatorInfo = *v
	c.Save()
}
func (c *ChatUser) GetLastChatTime() time.Time {
	return c.chatter.LastChatTime
}
func (c *ChatUser) SetLastChatTime(v time.Time) {
	c.chatter.LastChatTime = v
	c.Save()
}

func (c *ChatUser) SetActive(isActive bool) {
	c.chatter.IsActive = isActive
	c.Save()
}
func (c *ChatUser) IsActive() bool {
	return c.chatter.IsActive
}

func (c *ChatUser) Save() error {
	if err := UpdateUser(c.user); err != nil {
		return err
	}
	return UpdateChatter(c.chatter)
}
