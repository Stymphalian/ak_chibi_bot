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

type ChibiActorInterface interface {
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool
	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	HandleCommand(msg ChatMessage) (string, error)

	// TODO: Leaky interface
	UpdateChibi(username string, userDisplayName string, opInfo *spine.OperatorInfo) error
	CurrentInfo(userName string) (spine.OperatorInfo, error)
}
