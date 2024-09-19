package chat

import (
	"context"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

type ChatMessage struct {
	Username        string
	UserDisplayName string
	Message         string
}

type ChatMessageHandler interface {
	HandleMessage(msg ChatMessage) (string, error)
}

type ActorUpdater interface {
	CurrentInfo(ctx context.Context, username string) (spine.OperatorInfo, error)
	UpdateChibi(ctx context.Context, username string, usernameDisplay string, update *spine.OperatorInfo) error
}

type ChatCommand interface {
	Reply(c ActorUpdater) string
	UpdateActor(c ActorUpdater) error
}
