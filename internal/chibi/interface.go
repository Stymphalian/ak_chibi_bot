package chibi

import "github.com/Stymphalian/ak_chibi_bot/internal/misc"

type ChibiActorInterface interface {
	HandleCommand(userName string, userNameDisplay string, msg string) (string, error)
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool

	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	Close()
}
