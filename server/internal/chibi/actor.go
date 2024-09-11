package chibi

import (
	"log"
	"slices"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

type ChibiActor struct {
	spineService         *spine.SpineService
	ChatUsers            map[string]*spine.ChatUser
	LastChatterTime      time.Time
	client               spine.SpineClient
	chatCommandProcessor *ChatCommandProcessor
	excludeNames         []string
}

func NewChibiActor(
	spineService *spine.SpineService,
	client spine.SpineClient,
	excludeNames []string,
) *ChibiActor {
	a := &ChibiActor{
		spineService:    spineService,
		ChatUsers:       make(map[string]*spine.ChatUser, 0),
		LastChatterTime: time.Now(),
		client:          client,
		excludeNames:    excludeNames,
	}
	a.chatCommandProcessor = &ChatCommandProcessor{a, spineService, client}
	return a
}

func (c *ChibiActor) GiveChibiToUser(userName string, userNameDisplay string) error {
	// Skip giving chibis to these Users
	if slices.Contains(c.excludeNames, userName) {
		return nil
	}

	_, err := c.chatCommandProcessor.addRandomChibi(userName, userNameDisplay)
	if err == nil {
		log.Println("User joined. Adding a chibi for them ", userName)
		misc.Monitor.NumUsers += 1
	}
	return err
}

func (c *ChibiActor) RemoveUserChibi(userName string) error {
	if slices.Contains(c.excludeNames, userName) {
		return nil
	}
	_, err := c.client.RemoveOperator(
		&spine.RemoveOperatorRequest{UserName: userName},
	)
	if err != nil {
		log.Printf("Error removing chibi for %s: %s\n", userName, err)
	}

	delete(c.ChatUsers, userName)
	return nil
}

func (c *ChibiActor) HasChibi(userName string) bool {
	_, err := c.CurrentInfo(userName)
	if err != nil {
		_, ok := err.(*spine.UserNotFound)
		return !ok
	}
	return true
}

// TODO: Move this to command processor?
// TODO: Leaking operatorDetails
func (c *ChibiActor) SetToDefault(
	broadcasterName string,
	opName string,
	details misc.InitialOperatorDetails,
) {
	opInfo := c.spineService.OperatorFromDefault(opName, details)
	c.UpdateChatter(broadcasterName, broadcasterName, opInfo)
}

func (c *ChibiActor) HandleMessage(msg ChatMessage) (string, error) {
	if !c.HasChibi(msg.Username) {
		c.GiveChibiToUser(msg.Username, msg.UserDisplayName)
	}
	if msg.Message[0] != '!' {
		return "", nil
	}
	c.ChatUsers[msg.Username].LastChatTime = time.Now()

	return c.chatCommandProcessor.HandleMessage(
		msg.Username,
		msg.UserDisplayName,
		msg.Message,
	)
}

// TODO: Leaky interface
func (c *ChibiActor) UpdateChibi(username string, userDisplayName string, opInfo *spine.OperatorInfo) error {
	return c.chatCommandProcessor.UpdateChibi(
		username,
		userDisplayName,
		opInfo,
	)
}

func (c *ChibiActor) CurrentInfo(userName string) (spine.OperatorInfo, error) {
	chatUser, ok := c.ChatUsers[userName]
	if !ok {
		return *spine.EmptyOperatorInfo(), spine.NewUserNotFound("User not found: " + userName)
	}

	return chatUser.CurrentOperator, nil
}

func (c *ChibiActor) UpdateChatter(
	username string,
	usernameDisplay string,
	update *spine.OperatorInfo,
) {
	chatUser, ok := c.ChatUsers[username]
	if !ok {
		c.ChatUsers[username] = spine.NewChatUser(
			username,
			usernameDisplay,
			time.Now(),
		)
		chatUser = c.ChatUsers[username]
	}
	chatUser.UserNameDisplay = usernameDisplay
	chatUser.LastChatTime = time.Now()
	chatUser.CurrentOperator = *update
}
