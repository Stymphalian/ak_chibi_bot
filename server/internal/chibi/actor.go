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
	client               spine.SpineClient
	chatCommandProcessor ChatCommandProcessor
	excludeNames         []string
}

func NewChibiActor(client spine.SpineClient, excludeNames []string) *ChibiActor {
	return &ChibiActor{
		client:               client,
		chatCommandProcessor: ChatCommandProcessor{client},
		excludeNames:         excludeNames,
	}
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
	_, err := c.client.CurrentInfo(userName)
	if err != nil {
		_, ok := err.(*spine.UserNotFound)
		return !ok
	}
	return true
}

func (c *ChibiActor) SetToDefault(
	broadcasterName string,
	opName string,
	details misc.InitialOperatorDetails,
) {
	c.client.SetToDefault(broadcasterName, opName, details)
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
