package chibi

import (
	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

type ChibiActorInterface interface {
	chat.ChatMessageHandler
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool

	// TODO: Leaky interface
	UpdateChibi(username string, userDisplayName string, opInfo *spine.OperatorInfo) error
	CurrentInfo(userName string) (spine.OperatorInfo, error)
	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	UpdateChatter(userName string, usernameDisplay string, update *spine.OperatorInfo)
}
