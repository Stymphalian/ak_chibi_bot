package chibi

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

type ChibiActor struct {
	spineService         *spine.SpineService
	ChatUsers            map[string]*users.ChatUser
	LastChatterTime      time.Time
	client               spine.SpineClient
	chatCommandProcessor *chat.ChatCommandProcessor
	excludeNames         []string

	// TODO: Find a better way to get the roomId into the ChibiActors/ChatUsers
	roomId *uint
}

func NewChibiActor(
	spineService *spine.SpineService,
	client spine.SpineClient,
	excludeNames []string,
) *ChibiActor {
	a := &ChibiActor{
		spineService:         spineService,
		ChatUsers:            make(map[string]*users.ChatUser, 0),
		LastChatterTime:      misc.Clock.Now(),
		client:               client,
		chatCommandProcessor: chat.NewChatCommandProcessor(spineService),
		excludeNames:         excludeNames,
	}
	return a
}

func (c *ChibiActor) Close() error {
	// No need to send the remove Operators to the clients.
	// This room is going down on the server anyways and the WS connections
	// will be closed.
	for _, chatUser := range c.ChatUsers {
		chatUser.Close()
	}
	return nil
}

func (c *ChibiActor) SetRoomId(roomId uint) {
	c.roomId = &roomId
}

func (c *ChibiActor) GiveChibiToUser(ctx context.Context, userName string, userNameDisplay string) error {
	// Skip giving chibis to these Users
	if slices.Contains(c.excludeNames, userName) {
		return nil
	}

	operatorInfo, err := c.spineService.GetRandomOperator()
	if err != nil {
		return err
	}
	log.Printf("Giving %s the chibi %s\n", userName, operatorInfo.OperatorId)
	_, err = c.client.SetOperator(
		&spine.SetOperatorRequest{
			UserName:        userName,
			UserNameDisplay: userNameDisplay,
			Operator:        *operatorInfo,
		})
	if err != nil {
		log.Printf("Failed to set chibi (%s)", err.Error())
		return nil
	}
	c.UpdateChatter(ctx, userName, userNameDisplay, operatorInfo)

	// _, err := c.chatCommandProcessor.addRandomChibi(userName, userNameDisplay)
	if err == nil {
		log.Println("User joined. Adding a chibi for them ", userName)
		misc.Monitor.NumUsers += 1
	}
	return err
}

func (c *ChibiActor) RemoveUserChibi(ctx context.Context, userName string) error {
	if slices.Contains(c.excludeNames, userName) {
		return nil
	}
	_, err := c.client.RemoveOperator(
		&spine.RemoveOperatorRequest{UserName: userName},
	)
	if err != nil {
		log.Printf("Error removing chibi for %s: %s\n", userName, err)
	}
	c.ChatUsers[userName].SetActive(false)
	delete(c.ChatUsers, userName)
	return nil
}

func (c *ChibiActor) HasChibi(ctx context.Context, userName string) bool {
	_, err := c.CurrentInfo(ctx, userName)
	if err != nil {
		_, ok := err.(*spine.UserNotFound)
		return !ok
	}
	return true
}

// TODO: Move this to command processor?
// TODO: Leaking operatorDetails
func (c *ChibiActor) SetToDefault(
	ctx context.Context,
	broadcasterName string,
	opName string,
	details misc.InitialOperatorDetails,
) {
	opInfo := c.spineService.OperatorFromDefault(opName, details)
	c.UpdateChatter(ctx, broadcasterName, broadcasterName, opInfo)
}

func (c *ChibiActor) HandleMessage(msg chat.ChatMessage) (string, error) {
	ctx := context.Background()
	if !c.HasChibi(ctx, msg.Username) {
		c.GiveChibiToUser(ctx, msg.Username, msg.UserDisplayName)
	}
	if msg.Message[0] != '!' {
		return "", nil
	}
	c.ChatUsers[msg.Username].SetLastChatTime(misc.Clock.Now())

	current, err := c.CurrentInfo(ctx, msg.Username)
	if err != nil {
		switch err.(type) {
		case *spine.UserNotFound:
			log.Println("Chibi not found for user ", msg.Username)
		}
		return "", nil
	}

	chatCommand, err := c.chatCommandProcessor.HandleMessage(&current, msg)
	chatCommand.UpdateActor(c)
	return chatCommand.Reply(c), err
}

// TODO: Leaky interface
func (c *ChibiActor) UpdateChibi(ctx context.Context, username string, userDisplayName string, opInfo *spine.OperatorInfo) error {
	c.spineService.ValidateUpdateSetDefaultOtherwise(opInfo)

	_, err := c.client.SetOperator(
		&spine.SetOperatorRequest{
			UserName:        username,
			UserNameDisplay: userDisplayName,
			Operator:        *opInfo,
		})
	if err != nil {
		log.Printf("Failed to set chibi (%s)", err.Error())
		return nil
	}

	c.UpdateChatter(ctx, username, userDisplayName, opInfo)
	return nil
}

func (c *ChibiActor) CurrentInfo(ctx context.Context, userName string) (spine.OperatorInfo, error) {
	chatUser, ok := c.ChatUsers[userName]
	if !ok {
		return *spine.EmptyOperatorInfo(), spine.NewUserNotFound("User not found: " + userName)
	}

	return *chatUser.GetOperatorInfo(), nil
}

func (c *ChibiActor) UpdateChatter(
	ctx context.Context,
	username string,
	usernameDisplay string,
	update *spine.OperatorInfo,
) error {
	_, ok := c.ChatUsers[username]
	if !ok {
		userDb, err := users.GetOrInsertUser(ctx, username, usernameDisplay)
		if err != nil {
			return err
		}
		if c.roomId == nil {
			return fmt.Errorf("roomId must be set in order to update a ChatUser")
		}
		chatterDb, err := users.GetOrInsertChatter(ctx, *c.roomId, userDb, misc.Clock.Now(), update)
		if err != nil {
			return err
		}

		chatUser, err := users.NewChatUser(userDb, chatterDb)
		if err != nil {
			return err
		}
		c.ChatUsers[username] = chatUser
	}

	chatUser := c.ChatUsers[username]
	chatUser.SetLastChatTime(misc.Clock.Now())
	chatUser.SetOperatorInfo(update)
	chatUser.SetActive(true)
	// TODO: Make this more efficient. no need to save to DB if things haven't changed
	// chatUser.Save()
	return nil
}
