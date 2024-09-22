package chat

import (
	"context"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
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
	CurrentInfo(ctx context.Context, username string) (operator.OperatorInfo, error)
	UpdateChibi(ctx context.Context, username string, usernameDisplay string, update *operator.OperatorInfo) error
}

type ChatCommand interface {
	Reply(c ActorUpdater) string
	UpdateActor(c ActorUpdater) error
}
