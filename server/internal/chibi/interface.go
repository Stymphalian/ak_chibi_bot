package chibi

import "github.com/Stymphalian/ak_chibi_bot/server/internal/misc"

type ChibiActorInterface interface {
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool
	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	HandleCommand(userName string, userNameDisplay string, msg string) (string, error)
}
