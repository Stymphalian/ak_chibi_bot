package twitchbot

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/misc"
	"github.com/Stymphalian/ak_chibi_bot/spine"
)

type ChibiClient interface {
	HandleCommand(userName string, userNameDisplay string, msg string) (string, error)
	GiveChibiToUser(userName string, userNameDisplay string) error
	RemoveUserChibi(userName string) error
	HasChibi(userName string) bool
}

type ChibiActor struct {
	client       spine.SpineClient
	twitchConfig *misc.TwitchConfig
}

func NewChibiActor(client spine.SpineClient, config *misc.TwitchConfig) *ChibiActor {
	return &ChibiActor{client: client, twitchConfig: config}
}

func (c *ChibiActor) GiveChibiToUser(userName string, userNameDisplay string) error {
	// Skip giving chibis to these Users
	if userName == c.twitchConfig.Broadcaster ||
		slices.Contains(c.twitchConfig.ExcludeNames, userName) {
		return nil
	}

	_, err := c.client.CurrentInfo(userName)
	if err == nil {
		return nil
	}

	_, err = c.addRandomChibi(userName, userNameDisplay)
	if err == nil {
		log.Println("User joined. Adding a chibi for them ", userName)
	}
	return err
}

func (c *ChibiActor) RemoveUserChibi(userName string) error {
	if userName == c.twitchConfig.Broadcaster ||
		slices.Contains(c.twitchConfig.ExcludeNames, userName) {
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

func (c *ChibiActor) HandleCommand(userName string, userNameDisplay string, trimmed string) (string, error) {
	if !strings.HasPrefix(trimmed, "!chibi") {
		return "", nil
	}
	if len(trimmed) >= 100 {
		return "", nil
	}

	args := strings.Split(trimmed, " ")
	args = slices.DeleteFunc(args, func(s string) bool { return s == "" })
	if len(args) == 1 && args[0] == "!chibi" {
		return c.ChibiHelp(trimmed)
	}
	if args[0] != "!chibi" {
		return "", nil
	}

	// !chibi skins
	// !chibi anims
	// !chibi help
	// !chibi <operator names multi-space>
	// !chibi skin epoque
	// !chibi play Move
	// !chibi anim Special
	// !chibi stance base|battle
	// !chibi face front|back
	// !chibi walk <number>
	// !chibi enemy The last steam knight

	current, err := c.client.CurrentInfo(userName)
	if err != nil {
		switch err.(type) {
		case *spine.UserNotFound:
			log.Println("Chibi not found for user ", userName)
		}
		return "", errors.New("something went wrong please try again")
	}

	var msg string
	arg2 := strings.TrimSpace(args[1])
	switch arg2 {
	case "admin":
		if userName != c.twitchConfig.Broadcaster {
			log.Printf("Only broadcaster can use !chibi admin: %s\n", userName)
			return "", nil
		}
		// !chibi admin <username> "!chibi command"
		if len(args) < 3 {
			log.Printf("Only broadcaster can use !chibi admin not enough args %v\n", args)
			return "", nil
		}
		splitStr := strings.SplitN(trimmed, " ", 4)
		targetUser := splitStr[2]
		if len(splitStr) == 4 {
			restCommand := splitStr[3]
			return c.HandleCommand(targetUser, targetUser, restCommand)
		}
		return "", nil
	case "help":
		return c.ChibiHelp(trimmed)
	case "skins":
		return c.GetChibiInfo(userName, "skins")
	case "anims":
		return c.GetChibiInfo(userName, "anims")
	case "info":
		return c.GetChibiInfo(userName, "info")
	case "skin":
		msg, err = c.SetSkin(args, &current)
	case "anim":
		msg, err = c.SetAnimation(args, &current)
	case "play":
		msg, err = c.SetAnimation(args, &current)
	case "stance":
		msg, err = c.SetStance(args, &current)
	case "face":
		msg, err = c.SetFacing(args, &current)
	case "enemy":
		msg, err = c.SetEnemy(args, &current)
	case "walk":
		msg, err = c.SetWalk(args, &current)
	default:
		if _, ok := misc.MatchesKeywords(arg2, current.Animations); ok {
			msg, err = c.SetAnimation([]string{"!chibi", "play", arg2}, &current)
		} else if _, ok := misc.MatchesKeywords(arg2, current.Skins); ok {
			msg, err = c.SetSkin([]string{"!chibi", "skin", arg2}, &current)
		} else if _, ok := misc.MatchesKeywords(arg2, []string{"base", "battle"}); ok {
			msg, err = c.SetStance([]string{"!chibi", "stance", arg2}, &current)
		} else {
			// matches against operator names
			msg, err = c.SetChibiModel(trimmed, &current)
		}
	}

	if err == nil {
		c.UpdateChibi(userName, userNameDisplay, &current)
	}

	return msg, err
}

func (c *ChibiActor) validateUpdateSetDefaultOtherwise(update *spine.OperatorInfo) error {
	if len(update.Faction) == 0 {
		update.Faction = spine.FACTION_ENUM_OPERATOR
	}

	currentOp, err := c.client.GetOperator(&spine.GetOperatorRequest{
		OperatorId: update.OperatorId,
		Faction:    update.Faction,
	})
	if err != nil {
		return errors.New("something went wrong please try again")
	}

	if _, ok := currentOp.Skins[update.Skin]; !ok {
		update.Skin = "default"
	}
	facings := currentOp.Skins[update.Skin].Stances[update.ChibiType]
	if len(facings.Facings) == 0 {
		switch update.ChibiType {
		case spine.CHIBI_TYPE_ENUM_BASE:
			update.ChibiType = spine.CHIBI_TYPE_ENUM_BATTLE
		case spine.CHIBI_TYPE_ENUM_BATTLE:
			update.ChibiType = spine.CHIBI_TYPE_ENUM_BASE
		default:
			update.ChibiType = spine.CHIBI_TYPE_ENUM_BASE
		}
	}
	if _, ok := currentOp.Skins[update.Skin].Stances[update.ChibiType].Facings[update.Facing]; !ok {
		update.Facing = "Front"
	}

	animations := currentOp.Skins[update.Skin].Stances[update.ChibiType].Facings[update.Facing]
	if !slices.Contains(animations, update.Animation) {
		update.Animation = spine.GetDefaultAnimForChibiType(update.ChibiType)
	}
	if !slices.Contains(animations, update.Animation) {
		// If it still doesn't exist then just choose one randomly
		update.Animation = currentOp.Skins[update.Skin].Stances[update.ChibiType].Facings[update.Facing][0]
	}
	return nil
}

func (c *ChibiActor) UpdateChibi(username string, usernameDisplay string, update *spine.OperatorInfo) error {
	c.validateUpdateSetDefaultOtherwise(update)

	_, err := c.client.SetOperator(&spine.SetOperatorRequest{
		UserName:        username,
		UserNameDisplay: usernameDisplay,
		OperatorId:      update.OperatorId,
		Faction:         update.Faction,
		Skin:            update.Skin,
		ChibiType:       update.ChibiType,
		Facing:          update.Facing,
		Animation:       update.Animation,
		PositionX:       update.PositionX,
	})
	if err != nil {
		log.Printf("Failed to set chibi (%s)", err.Error())
		return nil
	}
	return nil
}

func (c *ChibiActor) ChibiHelp(trimmed string) (string, error) {
	log.Printf("!chibi_help command triggered with %v\n", trimmed)
	msg := `!chibi to control your Arknights chibi. ` +
		`"!chibi Rockrock" to change your operator. ` +
		`"!chibi play Move", "!chibi skin epoque#2" to change the animation and skin. ` +
		`"!chibi skins" and "!chibi anims" lists available skins and animations. ` +
		`"!chibi stance battle" to change from base or battle chibis. ` +
		`"!chibi enemy mandragora" to change into an enemy mob instead of an operator. ` +
		`"!chibi walk" to have your chibi walk around the screen. ` +
		`Look at the About section of the stream for more info.`
	return msg, nil
}

func (c *ChibiActor) SetSkin(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("incorrect usage: !chibi skin <skinName> (ie. !chibi skin default, !chibi skin epoque#2)")
	}

	skinName, hasSkin := misc.MatchesKeywords(args[2], current.Skins)
	if !hasSkin {
		return "", errors.New("unsupported skin")
	}
	current.Skin = skinName
	return "", nil
}

func (c *ChibiActor) SetAnimation(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("incorrect usage: !chibi play <animationName> (ie. !chibi play move, !chibi play special)")
	}

	animation, ok := misc.MatchesKeywords(args[2], current.Animations)
	if !ok {
		return "", errors.New("unsupported animation")
	}
	current.Animation = animation
	return "", nil
}

func (c *ChibiActor) SetStance(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("incorrect usage: !chibi stance <base|battle> (ie. !chibi stance base, !chibi stance battle)")
	}
	stance, err := spine.ChibiTypeEnum_Parse(args[2])
	if err != nil {
		return "", errors.New("incorrect usage: !chibi stance <base|battle> (ie. !chibi stance base, !chibi stance battle)")
	}
	current.ChibiType = stance
	return "", nil
}

func (c *ChibiActor) SetFacing(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("incorrect usage: !chibi face <front|back> (ie. !chibi face front, !chibi face back)")
	}
	facing, err := spine.ChibiFacingEnum_Parse(args[2])
	if err != nil {
		return "", errors.New("incorrect usage: !chibi face <front|back> (ie. !chibi face front, !chibi face back)")
	}
	if current.ChibiType == spine.CHIBI_TYPE_ENUM_BASE && facing == spine.CHIBI_FACING_ENUM_BACK {
		return "", errors.New("base chibi's can't face backwards. Try setting to battle stance first")
	}
	current.Facing = facing
	return "", nil
}

func (c *ChibiActor) SetEnemy(args []string, current *spine.OperatorInfo) (string, error) {
	errMsg := errors.New("incorrect usage: !chibi enemy <enemyname or ID> (ie. !chibi enemy Avenger, !chibi enemy SM8")
	if len(args) < 3 {
		return "", errMsg
	}
	trimmed := strings.Join(args[2:], " ")

	humanOperatorName := strings.TrimSpace(trimmed)
	operatorId, matches := c.client.GetOperatorIdFromName(humanOperatorName, spine.FACTION_ENUM_ENEMY)
	if matches != nil {
		if len(matches) == 0 {
			return "", errors.New("unknown enemy name")
		} else {
			return "", errors.New("Did you mean: " + strings.Join(matches, ", "))
		}
	}
	current.OperatorId = operatorId
	current.Faction = spine.FACTION_ENUM_ENEMY

	return "", nil
}

// Short cut to switch to base stance, and then invoke "Move" animation
func (c *ChibiActor) SetWalk(args []string, current *spine.OperatorInfo) (string, error) {
	current.ChibiType = spine.CHIBI_TYPE_ENUM_BASE

	// Set the animatino to "Move". If "Move" doesn't exist in the list of
	// animations then try to find an animation with "Move" in its name
	moveAnimation := "Move"
	if !slices.Contains(current.Animations, moveAnimation) {
		for _, animation := range current.Animations {
			if strings.Contains(animation, "Move") {
				moveAnimation = animation
				break
			}
		}
	}
	current.Animation = moveAnimation

	if len(args) == 3 {
		errMsg := errors.New("incorrect usage: !chibi walk <number> (ie. !chibi walk 1, !chibi walk 0.2)")
		desiredPosition, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return "", errMsg
		}
		if (desiredPosition < 0.0) || (desiredPosition > 1.0) {
			return "", errMsg
		}
		current.PositionX = &desiredPosition
	}
	return "", nil
}

func (c *ChibiActor) SetChibiModel(trimmed string, current *spine.OperatorInfo) (string, error) {
	log.Printf("!chibi command triggered with %v\n", trimmed)
	args := strings.Split(trimmed, " ")
	errMsg := errors.New("incorrect usage: !chibi <name> (ie. !chibi Amiya, !chibi Lava Alter)")
	if len(args) < 2 {
		return "", errMsg
	}

	splitStrs := strings.SplitN(trimmed, " ", 2)
	if len(splitStrs) != 2 {
		return "", errMsg
	}
	humanOperatorName := strings.TrimSpace(splitStrs[1])
	operatorId, matches := c.client.GetOperatorIdFromName(humanOperatorName, spine.FACTION_ENUM_OPERATOR)
	if matches != nil {
		if len(matches) == 0 {
			return "", errors.New("unknown operator name")
		} else {
			return "", errors.New("Did you mean: " + strings.Join(matches, ", "))
		}
	}

	current.OperatorId = operatorId
	current.Skin = "default"
	current.Faction = spine.FACTION_ENUM_OPERATOR
	return "", nil
}

func (c *ChibiActor) addRandomChibi(userName string, userNameDisplay string) (string, error) {
	operatorIds, err := c.client.GetOperatorIds(spine.FACTION_ENUM_OPERATOR)
	if err != nil {
		return "", err
	}

	index := rand.Intn(len(operatorIds))
	operatorId := operatorIds[index]

	operatorData, err := c.client.GetOperator(&spine.GetOperatorRequest{
		OperatorId: operatorId,
		Faction:    spine.FACTION_ENUM_OPERATOR,
	})
	if err != nil {
		return "", err
	}
	chibiType := spine.CHIBI_TYPE_ENUM_BASE
	stanceMap, ok := operatorData.Skins["default"].Stances[spine.CHIBI_TYPE_ENUM_BASE]
	if !ok {
		chibiType = spine.CHIBI_TYPE_ENUM_BATTLE
	}
	if len(stanceMap.Facings) == 0 {
		chibiType = spine.CHIBI_TYPE_ENUM_BATTLE
	}

	log.Printf("Giving %s the chibi %s\n", userName, operatorId)
	_, err = c.client.SetOperator(&spine.SetOperatorRequest{
		UserName:        userName,
		UserNameDisplay: userNameDisplay,
		OperatorId:      operatorId,
		Faction:         spine.FACTION_ENUM_OPERATOR,
		Skin:            "default",
		ChibiType:       chibiType,
		Facing:          spine.CHIBI_FACING_ENUM_FRONT,
		Animation:       spine.GetDefaultAnimForChibiType(chibiType),
		PositionX:       nil,
	})
	return "", err
}

func (c *ChibiActor) GetChibiInfo(userName string, subInfoName string) (string, error) {
	current, err := c.client.CurrentInfo(userName)
	if err != nil {
		return "", nil
	}

	var msg string
	switch subInfoName {
	case "skins":
		msg = fmt.Sprintf("Available skins for %s: %s", current.Name, strings.Join(current.Skins, ", "))
	case "anims":
		msg = fmt.Sprintf("Available animations for %s: %s", current.Name, strings.Join(current.Animations, ","))
	case "info":
		msg = fmt.Sprintf("Current Chibi: %s, %s, %s, %s, %s", current.Name, current.Skin, current.ChibiType, current.Facing, current.Animation)
	default:
		return "", errors.New("incorrect usage: !chibi <skins|anims|info>")
	}
	return msg, nil
}
