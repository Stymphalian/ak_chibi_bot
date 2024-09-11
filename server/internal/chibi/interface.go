package chibi

import (
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
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

type ChibiActorInterface interface {
	ChatMessageHandler
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool

	// TODO: Leaky interface
	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	UpdateChibi(username string, userDisplayName string, opInfo *spine.OperatorInfo) error
	CurrentInfo(userName string) (spine.OperatorInfo, error)
	UpdateChatter(userName string, usernameDisplay string, update *spine.OperatorInfo)
}
