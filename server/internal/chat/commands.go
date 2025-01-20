package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

type ChatCommandNoOp struct{}

func (c *ChatCommandNoOp) Reply(a ActorUpdater) string      { return "" }
func (c *ChatCommandNoOp) UpdateActor(a ActorUpdater) error { return nil }

type ChatCommandSimpleMessage struct {
	replyMessage string
}

func (c *ChatCommandSimpleMessage) Reply(a ActorUpdater) string      { return c.replyMessage }
func (c *ChatCommandSimpleMessage) UpdateActor(a ActorUpdater) error { return nil }

type ChatCommandInfo struct {
	username string
	info     string
}

func (c *ChatCommandInfo) Reply(chibiActor ActorUpdater) string {
	ctx := context.Background()
	current, err := chibiActor.CurrentInfo(ctx, c.username)
	if err != nil {
		return ""
	}

	var msg string
	switch c.info {
	case "skins":
		msg = fmt.Sprintf("%s skins: %s", current.OperatorDisplayName, strings.Join(current.Skins, ", "))
	case "anims":
		msg = fmt.Sprintf("%s animations: %s", current.OperatorDisplayName, strings.Join(current.AvailableAnimations, ","))
	case "info":
		currentAnimations := current.Action.GetAnimations(current.CurrentAction)
		msg = fmt.Sprintf(
			"%s: %s, %s, %s, (%s)",
			current.OperatorDisplayName,
			current.Skin,
			current.ChibiStance,
			current.Facing,
			strings.Join(currentAnimations, ","),
		)
	default:
		return ""
	}
	return msg
}
func (c *ChatCommandInfo) UpdateActor(a ActorUpdater) error { return nil }

type ChatCommandUpdateActor struct {
	replyMessage    string
	username        string
	usernameDisplay string
	twitchUserId    string
	update          *operator.OperatorInfo
}

func (c *ChatCommandUpdateActor) Reply(a ActorUpdater) string { return c.replyMessage }
func (c *ChatCommandUpdateActor) UpdateActor(a ActorUpdater) error {
	ctx := context.Background()
	return a.UpdateChibi(ctx, misc.UserInfo{
		Username:        c.username,
		UsernameDisplay: c.usernameDisplay,
		TwitchUserId:    c.twitchUserId,
	}, c.update)
}

type ChatCommandSavePrefsAction int

const (
	ChatCommandSaveChibi_Save   = ChatCommandSavePrefsAction(0)
	ChatCommandSaveChibi_Remove = ChatCommandSavePrefsAction(1)
)

type ChatCommandSavePrefs struct {
	replyMessage    string
	username        string
	usernameDisplay string
	twitchUserId    string
	update          *operator.OperatorInfo
	action          ChatCommandSavePrefsAction
}

func (c *ChatCommandSavePrefs) Reply(a ActorUpdater) string { return c.replyMessage }
func (c *ChatCommandSavePrefs) UpdateActor(a ActorUpdater) error {
	ctx := context.Background()
	ui := misc.UserInfo{
		Username:        c.username,
		UsernameDisplay: c.usernameDisplay,
		TwitchUserId:    c.twitchUserId,
	}

	if c.action == ChatCommandSaveChibi_Save {
		return a.SaveUserPreferences(ctx, ui, c.update)
	} else {
		return a.ClearUserPreferences(ctx, ui)
	}
}

type ChatCommandShowMessage struct {
	replyMessage    string
	username        string
	usernameDisplay string
	twitchUserId    string
	message         string
}

func (c *ChatCommandShowMessage) Reply(a ActorUpdater) string { return c.replyMessage }
func (c *ChatCommandShowMessage) UpdateActor(a ActorUpdater) error {
	ctx := context.Background()
	ui := misc.UserInfo{
		Username:        c.username,
		UsernameDisplay: c.usernameDisplay,
		TwitchUserId:    c.twitchUserId,
	}
	return a.ShowMessage(ctx, ui, c.message)
}

type ChatCommandFollow struct {
	replyMessage    string
	username        string
	usernameDisplay string
	twitchUserId    string
	update          *operator.OperatorInfo
}

func (c *ChatCommandFollow) Reply(a ActorUpdater) string { return c.replyMessage }
func (c *ChatCommandFollow) UpdateActor(a ActorUpdater) error {
	ctx := context.Background()
	return a.FollowChibi(ctx, misc.UserInfo{
		Username:        c.username,
		UsernameDisplay: c.usernameDisplay,
		TwitchUserId:    c.twitchUserId,
	}, c.update)
}

type ChatCommandFindMe struct {
	replyMessage    string
	username        string
	usernameDisplay string
	twitchUserId    string
}

func (c *ChatCommandFindMe) Reply(a ActorUpdater) string { return c.replyMessage }
func (c *ChatCommandFindMe) UpdateActor(a ActorUpdater) error {
	ctx := context.Background()
	return a.FindOperator(ctx, misc.UserInfo{
		Username:        c.username,
		UsernameDisplay: c.usernameDisplay,
		TwitchUserId:    c.twitchUserId,
	})
}
