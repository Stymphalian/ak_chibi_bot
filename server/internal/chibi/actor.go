package chibi

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	spine "github.com/Stymphalian/ak_chibi_bot/server/internal/spine_runtime"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

type ChibiActor struct {
	spineService *operator.OperatorService
	usersRepo    users.UserRepository
	chattersRepo users.ChatterRepository

	ChatUsers            map[string]*users.ChatUser
	lastChatterTime      time.Time
	client               spine.SpineClient
	chatCommandProcessor *chat.ChatCommandProcessor
	excludeNames         []string

	// TODO: Find a better way to get the roomId into the ChibiActors/ChatUsers
	roomId uint
}

func NewChibiActor(
	roomId uint,
	spineService *operator.OperatorService,
	usersRepo users.UserRepository,
	chattersRepo users.ChatterRepository,
	client spine.SpineClient,
	excludeNames []string,
) *ChibiActor {
	a := &ChibiActor{
		spineService: spineService,
		usersRepo:    usersRepo,
		chattersRepo: chattersRepo,

		ChatUsers:            make(map[string]*users.ChatUser, 0),
		lastChatterTime:      misc.Clock.Now(),
		client:               client,
		chatCommandProcessor: chat.NewChatCommandProcessor(spineService),
		excludeNames:         excludeNames,
		roomId:               roomId,
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

func (c *ChibiActor) GetLastChatterTime() time.Time {
	return c.lastChatterTime
}

func (c *ChibiActor) GiveChibiToUser(ctx context.Context, userInfo misc.UserInfo) error {
	// Skip giving chibis to these Users
	if slices.Contains(c.excludeNames, userInfo.Username) {
		return nil
	}

	operatorInfo, err := c.spineService.GetRandomOperator()
	if err != nil {
		return err
	}
	log.Printf("Giving %s the chibi %s\n", userInfo.Username, operatorInfo.OperatorId)
	_, err = c.client.SetOperator(
		&spine.SetOperatorRequest{
			UserName:        userInfo.Username,
			UserNameDisplay: userInfo.UsernameDisplay,
			Operator:        *operatorInfo,
		})
	if err != nil {
		log.Printf("Failed to set chibi (%s)", err.Error())
		return nil
	}
	c.UpdateChatter(ctx, userInfo, operatorInfo)

	log.Println("User joined. Adding a chibi for them ", userInfo.Username)
	misc.Monitor.NumUsers += 1
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
	// TODO : Need to check that user exists in chatUsers before removing.
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
	userInfo misc.UserInfo,
	opName string,
	details misc.InitialOperatorDetails,
) {
	opInfo := c.spineService.OperatorFromDefault(opName, details)
	c.UpdateChatter(ctx, userInfo, opInfo)
}

func (c *ChibiActor) HandleMessage(msg chat.ChatMessage) (string, error) {
	ctx := context.Background()
	if !c.HasChibi(ctx, msg.Username) {
		c.GiveChibiToUser(ctx, misc.UserInfo{
			Username:        msg.Username,
			UsernameDisplay: msg.UserDisplayName,
			TwitchUserId:    msg.TwitchUserId,
		})
	}
	c.ChatUsers[msg.Username].SetLastChatTime(misc.Clock.Now())
	c.lastChatterTime = misc.Clock.Now()
	if msg.Message[0] != '!' {
		return "", nil
	}

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
func (c *ChibiActor) UpdateChibi(ctx context.Context, userinfo misc.UserInfo, opInfo *operator.OperatorInfo) error {
	c.spineService.ValidateUpdateSetDefaultOtherwise(opInfo)

	_, err := c.client.SetOperator(
		&spine.SetOperatorRequest{
			UserName:        userinfo.Username,
			UserNameDisplay: userinfo.UsernameDisplay,
			Operator:        *opInfo,
		})
	if err != nil {
		log.Printf("Failed to set chibi (%s)", err.Error())
		return nil
	}

	c.UpdateChatter(ctx, userinfo, opInfo)
	return nil
}

func (c *ChibiActor) CurrentInfo(ctx context.Context, userName string) (operator.OperatorInfo, error) {
	chatUser, ok := c.ChatUsers[userName]
	if !ok {
		return *operator.EmptyOperatorInfo(), spine.NewUserNotFound("User not found: " + userName)
	}

	return *chatUser.GetOperatorInfo(), nil
}

func (c *ChibiActor) UpdateChatter(
	ctx context.Context,
	userInfo misc.UserInfo,
	update *operator.OperatorInfo,
) error {
	_, ok := c.ChatUsers[userInfo.Username]
	if !ok {
		userDb, err := c.usersRepo.GetOrInsertUser(ctx, userInfo)
		if err != nil {
			return err
		}
		chatterDb, err := c.chattersRepo.GetOrInsertChatter(ctx, c.roomId, userDb, misc.Clock.Now(), update)
		if err != nil {
			return err
		}

		chatUser, err := users.NewChatUser(
			c.usersRepo,
			c.chattersRepo,
			userDb.UserId,
			chatterDb.ChatterId,
		)
		if err != nil {
			return err
		}
		c.ChatUsers[userInfo.Username] = chatUser
	}

	chatUser := c.ChatUsers[userInfo.Username]
	chatUser.UpdateWithLatestChat(update)
	// TODO: Make this more efficient. no need to save to DB if things haven't changed
	// chatUser.Save()
	return nil
}
