package chat

import (
	"context"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

type ChatMessage struct {
	Username        string
	UserDisplayName string
	TwitchUserId    string
	Message         string
}

type ChatMessageHandler interface {
	HandleMessage(msg ChatMessage) (string, error)
}

type ActorUpdater interface {
	CurrentInfo(ctx context.Context, username string) (operator.OperatorInfo, error)
	UpdateChibi(ctx context.Context, userInfo misc.UserInfo, update *operator.OperatorInfo) error
	FollowChibi(ctx context.Context, userInfo misc.UserInfo, update *operator.OperatorInfo) error
	SaveUserPreferences(ctx context.Context, userInfo misc.UserInfo, update *operator.OperatorInfo) error
	ClearUserPreferences(ctx context.Context, userInfo misc.UserInfo) error
	ShowMessage(ctx context.Context, userInfo misc.UserInfo, msg string) error
}

type ChatCommand interface {
	Reply(c ActorUpdater) string
	UpdateActor(c ActorUpdater) error
}
