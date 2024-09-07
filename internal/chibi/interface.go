package chibi

type ChibiActorInterface interface {
	HandleCommand(userName string, userNameDisplay string, msg string) (string, error)
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool
	Close()
}
