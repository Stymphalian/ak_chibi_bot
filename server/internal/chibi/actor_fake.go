package chibi

import (
	"context"
	"errors"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	spine "github.com/Stymphalian/ak_chibi_bot/server/internal/spine_runtime"
)

type FakeChibiActor struct {
	Users map[string]operator.OperatorInfo
}

func NewFakeChibiActor() *FakeChibiActor {
	return &FakeChibiActor{
		Users: make(map[string]operator.OperatorInfo, 0),
	}
}

func (f *FakeChibiActor) GiveChibiToUser(ctx context.Context, userName string, userNameDisplay string) error {
	opInfo := *operator.EmptyOperatorInfo()
	opInfo.OperatorId = userName
	opInfo.OperatorDisplayName = userName
	f.Users[userName] = opInfo
	return nil
}

func (f *FakeChibiActor) RemoveUserChibi(ctx context.Context, userName string) error {
	delete(f.Users, userName)
	return nil
}

func (f *FakeChibiActor) HasChibi(ctx context.Context, userName string) bool {
	_, ok := f.Users[userName]
	return ok
}

func (f *FakeChibiActor) SetToDefault(ctx context.Context, broadcasterName string, opName string, details misc.InitialOperatorDetails) {
	opInfo := f.Users[broadcasterName]
	opInfo.OperatorId = "DefaultChibi"
	opInfo.OperatorDisplayName = "DefaultChibi"
	f.Users[broadcasterName] = opInfo
}

func (f *FakeChibiActor) HandleMessage(msg chat.ChatMessage) (string, error) {
	if strings.HasPrefix(msg.Message, "!") {
		opInfo := *operator.EmptyOperatorInfo()
		opInfo.OperatorId = msg.Message
		f.Users[msg.Username] = opInfo

		if msg.Message == "!chibi error" {
			return "invalid !chibi", errors.New("error message")
		} else {
			return "valid", nil
		}
	} else {
		opInfo := *operator.EmptyOperatorInfo()
		opInfo.OperatorId = "Invalid"
		f.Users[msg.Username] = opInfo
		return "invalid", errors.New("error message")
	}
}

func (f *FakeChibiActor) UpdateChibi(ctx context.Context, username string, userDisplayName string, opInfo *operator.OperatorInfo) error {
	f.Users[username] = *opInfo
	return nil
}

func (f *FakeChibiActor) FollowChibi(ctx context.Context, username string, userDisplayName string, opInfo *operator.OperatorInfo) error {
	f.Users[username] = *opInfo
	return nil
}

func (f *FakeChibiActor) CurrentInfo(ctx context.Context, userName string) (operator.OperatorInfo, error) {
	if _, ok := f.Users[userName]; !ok {
		return *operator.EmptyOperatorInfo(), spine.NewUserNotFound("User not found: " + userName)
	} else {
		return f.Users[userName], nil
	}
}

func (f *FakeChibiActor) UpdateChatter(
	ctx context.Context,
	username string,
	usernameDisplay string,
	update *operator.OperatorInfo,
) error {
	f.Users[username] = *update
	return nil
}

func (c *FakeChibiActor) ShowMessage(ctx context.Context, userInfo misc.UserInfo, msg string) error {
	return nil
}
