package chibi

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	spine "github.com/Stymphalian/ak_chibi_bot/server/internal/spine_runtime"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

type ChibiActor struct {
	spineService  *operator.OperatorService
	usersRepo     users.UserRepository
	chattersRepo  users.ChatterRepository
	userPrefsRepo users.UserPreferencesRepository

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
	userPrefsRepo users.UserPreferencesRepository,
	chattersRepo users.ChatterRepository,
	client spine.SpineClient,
	excludeNames []string,
) *ChibiActor {
	a := &ChibiActor{
		spineService:  spineService,
		usersRepo:     usersRepo,
		userPrefsRepo: userPrefsRepo,
		chattersRepo:  chattersRepo,

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
	for key, chatUser := range c.ChatUsers {
		chatUser.Close()
		delete(c.ChatUsers, key)
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

	userPrefs, _ := c.userPrefsRepo.GetByTwitchIdOrNil(ctx, userInfo.TwitchUserId)
	var operatorInfo *operator.OperatorInfo
	if userPrefs != nil {
		operatorInfo = &userPrefs.OperatorInfo
	} else {
		var err error
		operatorInfo, err = c.spineService.GetRandomOperator()
		if err != nil {
			return err
		}
	}

	log.Printf("Giving %s the chibi %s\n", userInfo.Username, operatorInfo.OperatorId)
	err := c.UpdateChibi(ctx, userInfo, operatorInfo)
	if err != nil {
		log.Printf("Failed to update chatter: %s\n", err.Error())
		return err
	}

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
	if _, ok := c.ChatUsers[userName]; !ok {
		log.Printf("Error removing chibi for %s. User not found\n", userName)
		return nil
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
	userInfo misc.UserInfo,
	opName string,
	details misc.InitialOperatorDetails,
) {
	opInfo := c.spineService.OperatorFromDefault(opName, details)
	err := c.UpdateChatter(ctx, userInfo, opInfo)
	if err != nil {
		log.Printf("Failed to SetToDefault for %s: %s\n", userInfo.Username, err)
	}
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
	if cu, ok := c.ChatUsers[msg.Username]; ok {
		cu.SetLastChatTime(misc.Clock.Now())
	}
	c.lastChatterTime = misc.Clock.Now()
	if len(msg.Message) == 0 {
		return "", nil
	}
	// if msg.Message[0] != '!' {
	// 	return "", nil
	// }

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
		log.Printf("Failed to set chibi (%s)\n", err.Error())
		return nil
	}

	return c.UpdateChatter(ctx, userinfo, opInfo)
}

func (c *ChibiActor) FollowChibi(ctx context.Context, userinfo misc.UserInfo, opInfo *operator.OperatorInfo) error {
	if opInfo.CurrentAction != operator.ACTION_FOLLOW {
		return nil
	}
	// Make sure the target user exists, otherwise we can't follow.
	targetUsername := strings.ToLower(opInfo.Action.ActionFollowTarget)
	_, ok := c.ChatUsers[targetUsername]
	if !ok {
		return nil
	}

	return c.UpdateChibi(ctx, userinfo, opInfo)
}

func (c *ChibiActor) CurrentInfo(ctx context.Context, userName string) (operator.OperatorInfo, error) {
	chatUser, ok := c.ChatUsers[userName]
	if !ok {
		return *operator.EmptyOperatorInfo(), spine.NewUserNotFound("User not found: " + userName)
	}

	return *chatUser.GetOperatorInfo(), nil
}

func (c *ChibiActor) SaveUserPreferences(ctx context.Context, userInfo misc.UserInfo, update *operator.OperatorInfo) error {
	userDb, err := c.usersRepo.GetByTwitchId(ctx, userInfo.TwitchUserId)
	if err != nil {
		return err
	}
	return c.userPrefsRepo.SetByUserId(ctx, userDb.UserId, update)
}

func (c *ChibiActor) ClearUserPreferences(ctx context.Context, userInfo misc.UserInfo) error {
	userDb, err := c.usersRepo.GetByTwitchId(ctx, userInfo.TwitchUserId)
	if err != nil {
		return err
	}
	return c.userPrefsRepo.DeleteByUserId(ctx, userDb.UserId)
}

func (c *ChibiActor) GetUserPreferences(ctx context.Context, userInfo misc.UserInfo) (*operator.OperatorInfo, error) {
	userDb, err := c.usersRepo.GetByTwitchId(ctx, userInfo.TwitchUserId)
	if err != nil {
		return nil, err
	}
	val, err := c.userPrefsRepo.GetByUserIdOrNil(ctx, userDb.UserId)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	return &val.OperatorInfo, nil
}

func (c *ChibiActor) ShowMessage(ctx context.Context, userInfo misc.UserInfo, msg string) error {
	_, ok := c.ChatUsers[userInfo.Username]
	if !ok {
		return nil
	}
	_, err := c.client.ShowChatMessage(&spine.ShowChatMessageRequest{
		UserName: userInfo.Username,
		Message:  msg,
	})
	return err
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
	err := chatUser.UpdateWithLatestChat(update)
	if err != nil {
		return err
	}
	// TODO: Make this more efficient. no need to save to DB if things haven't changed
	// chatUser.Save()
	return nil
}
