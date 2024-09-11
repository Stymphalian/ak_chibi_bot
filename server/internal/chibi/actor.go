package chibi

import (
	"log"
	"slices"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

const (
	MIN_ANIMATION_SPEED     = 0.1
	DEFAULT_ANIMATION_SPEED = 1.0
	MAX_ANIMATION_SPEED     = 5.0
)

type ChibiActor struct {
	ChatUsers map[string]*spine.ChatUser

	client               spine.SpineClient
	chatCommandProcessor ChatCommandProcessor
	excludeNames         []string
}

func NewChibiActor(client spine.SpineClient, excludeNames []string) *ChibiActor {
	a := &ChibiActor{
		ChatUsers:            make(map[string]*spine.ChatUser, 0),
		client:               client,
		chatCommandProcessor: ChatCommandProcessor{nil, client},
		excludeNames:         excludeNames,
	}
	a.chatCommandProcessor.chibiActor = a
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
	return nil
}

func (c *ChibiActor) HasChibi(userName string) bool {
	// _, err := c.client.CurrentInfo(userName)
	_, err := c.CurrentInfo(userName)
	if err != nil {
		_, ok := err.(*spine.UserNotFound)
		return !ok
	}
	return true
}

// TODO: Move this to command processor?
func (c *ChibiActor) SetToDefault(
	broadcasterName string,
	opName string,
	details misc.InitialOperatorDetails,
) {
	if len(opName) == 0 {
		opName = "Amiya"
	}

	faction := spine.FACTION_ENUM_OPERATOR
	opId, matches := c.client.GetOperatorIdFromName(opName, spine.FACTION_ENUM_OPERATOR)
	if matches != nil {
		faction = spine.FACTION_ENUM_ENEMY
		opId, matches = c.client.GetOperatorIdFromName(opName, spine.FACTION_ENUM_ENEMY)
	}
	if matches != nil {
		log.Panic("Failed to get operator id", matches)
	}
	stance, err2 := spine.ChibiStanceEnum_Parse(details.Stance)
	if err2 != nil {
		log.Panic("Failed to parse stance", err2)
	}

	opResp, err := c.client.GetOperator(&spine.GetOperatorRequest{opId, faction})
	if err != nil {
		log.Panic("Failed to fetch operator info")
	}
	availableAnims := opResp.Skins[details.Skin].Stances[stance].Facings[spine.CHIBI_FACING_ENUM_FRONT]
	availableAnims = spine.FilterAnimations(availableAnims)
	availableSkins := opResp.GetSkinNames()

	opInfo := spine.NewOperatorInfo(
		opResp.OperatorName,
		faction,
		opId,
		details.Skin,
		stance,
		spine.CHIBI_FACING_ENUM_FRONT,
		availableSkins,
		availableAnims,
		1.0,
		misc.NewOption(misc.Vector2{X: details.PositionX, Y: 0.0}),
		spine.ACTION_PLAY_ANIMATION,
		spine.NewActionPlayAnimation(details.Animations),
	)
	c.UpdateChatter(
		broadcasterName,
		broadcasterName,
		&opInfo,
	)
	// c.client.SetToDefault(broadcasterName, opName, details)
}

func (c *ChibiActor) HandleCommand(msg ChatMessage) (string, error) {
	return c.chatCommandProcessor.HandleCommand(
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
		c.ChatUsers[username] = &spine.ChatUser{
			UserName:        username,
			UserNameDisplay: usernameDisplay,
		}
		chatUser = c.ChatUsers[username]
	}
	chatUser.UserNameDisplay = usernameDisplay
	chatUser.CurrentOperator = *update
}
