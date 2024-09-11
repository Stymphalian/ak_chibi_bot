package chibi

import (
	"errors"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

type FakeChibiActor struct {
	Users map[string]spine.OperatorInfo
}

func NewFakeChibiActor() *FakeChibiActor {
	return &FakeChibiActor{
		Users: make(map[string]spine.OperatorInfo, 0),
	}
}

func (f *FakeChibiActor) GiveChibiToUser(userName string, userNameDisplay string) error {
	opInfo := *spine.EmptyOperatorInfo()
	opInfo.OperatorId = userName
	opInfo.OperatorDisplayName = userName
	f.Users[userName] = opInfo
	return nil
}

func (f *FakeChibiActor) RemoveUserChibi(userName string) error {
	delete(f.Users, userName)
	return nil
}

func (f *FakeChibiActor) HasChibi(userName string) bool {
	_, ok := f.Users[userName]
	return ok
}

func (f *FakeChibiActor) SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails) {
	opInfo := f.Users[broadcasterName]
	opInfo.OperatorId = "DefaultChibi"
	opInfo.OperatorDisplayName = "DefaultChibi"
	f.Users[broadcasterName] = opInfo
	// f.Users[broadcasterName].OperatorId = "Chibi"
}

func (f *FakeChibiActor) HandleMessage(msg ChatMessage) (string, error) {
	if strings.HasPrefix(msg.Message, "!") {
		opInfo := *spine.EmptyOperatorInfo()
		opInfo.OperatorId = msg.Message
		f.Users[msg.Username] = opInfo
		return "valid", nil
	} else {
		opInfo := *spine.EmptyOperatorInfo()
		opInfo.OperatorId = "Invalid"
		f.Users[msg.Username] = opInfo
		return "invalid", errors.New("Error message")
	}
}

func (f *FakeChibiActor) UpdateChibi(username string, userDisplayName string, opInfo *spine.OperatorInfo) error {
	f.Users[username] = *opInfo
	return nil
}

func (f *FakeChibiActor) CurrentInfo(userName string) (spine.OperatorInfo, error) {
	if _, ok := f.Users[userName]; !ok {
		return *spine.EmptyOperatorInfo(), spine.NewUserNotFound("User not found: " + userName)
	} else {
		return f.Users[userName], nil
	}
}

func (f *FakeChibiActor) UpdateChatter(
	username string,
	usernameDisplay string,
	update *spine.OperatorInfo,
) {
	f.Users[username] = *update
}
