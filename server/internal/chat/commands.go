package chat

import (
	"context"
	"fmt"
	"strings"

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
	update          *operator.OperatorInfo
}

func (c *ChatCommandUpdateActor) Reply(a ActorUpdater) string { return c.replyMessage }
func (c *ChatCommandUpdateActor) UpdateActor(a ActorUpdater) error {
	ctx := context.Background()
	return a.UpdateChibi(ctx, c.username, c.usernameDisplay, c.update)
}
