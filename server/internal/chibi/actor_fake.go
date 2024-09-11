package chibi

import (
	"errors"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type FakeChibiActor struct {
	Users map[string]string
}

func NewFakeChibiActor() *FakeChibiActor {
	return &FakeChibiActor{
		Users: make(map[string]string, 0),
	}
}

func (f *FakeChibiActor) GiveChibiToUser(userName string, userNameDisplay string) error {
	f.Users[userName] = "Chibi"
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
	f.Users[broadcasterName] = "Chibi"
}

func (f *FakeChibiActor) HandleCommand(msg ChatMessage) (string, error) {
	if strings.HasPrefix(msg.Message, "!") {
		f.Users[msg.Username] = msg.Message
		return "valid", nil
	} else {
		f.Users[msg.Username] = "Invalid"
		return "invalid", errors.New("Error message")
	}
}
