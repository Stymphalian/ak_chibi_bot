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
	// !chibi admin <username> "!chibi command"

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
	case "who":
		return c.GetWhoInfo(args, &current)
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
		update.Skin = spine.DEFAULT_SKIN_NAME
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
	for _, anim := range update.CurrentAnimations {
		if !slices.Contains(animations, anim) {
			update.CurrentAnimations = []string{spine.GetDefaultAnimForChibiType(update.ChibiType)}
			break
		}
	}
	// If it still doesn't exist then just choose one randomly
	for _, anim := range update.CurrentAnimations {
		if !slices.Contains(animations, anim) {
			update.CurrentAnimations = []string{
				currentOp.Skins[update.Skin].Stances[update.ChibiType].Facings[update.Facing][0],
			}
			break
		}
	}
	return nil
}

func (c *ChibiActor) UpdateChibi(username string, usernameDisplay string, update *spine.OperatorInfo) error {
	c.validateUpdateSetDefaultOtherwise(update)

	_, err := c.client.SetOperator(&spine.SetOperatorRequest{
		UserName:        username,
		UserNameDisplay: usernameDisplay,
		Operator: spine.OperatorInfo{
			OperatorId:        update.OperatorId,
			Faction:           update.Faction,
			Skin:              update.Skin,
			ChibiType:         update.ChibiType,
			Facing:            update.Facing,
			CurrentAnimations: update.CurrentAnimations,
			TargetPos:         update.TargetPos,
		},
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
		return "", errors.New("")
	}

	skinName, hasSkin := misc.MatchesKeywords(args[2], current.Skins)
	if !hasSkin {
		return "", errors.New("")
	}
	current.Skin = skinName
	return "", nil
}

func (c *ChibiActor) SetAnimation(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("")
	}

	animation, ok := misc.MatchesKeywords(args[2], current.Animations)
	if !ok {
		return "", errors.New("")
	}
	current.CurrentAnimations = []string{animation}

	if len(args) >= 4 {
		animations := make([]string, 0)
		skipAdding := false
		for i := 2; i < len(args); i++ {
			anim, ok := misc.MatchesKeywords(args[i], current.Animations)
			if !ok {
				skipAdding = true
				break
			}
			animations = append(animations, anim)
		}
		if !skipAdding {
			current.CurrentAnimations = animations
		}
	}
	return "", nil
}

func (c *ChibiActor) SetStance(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("try something like !chibi stance battle")
	}
	stance, err := spine.ChibiTypeEnum_Parse(args[2])
	if err != nil {
		return "", errors.New("try something like !chibi stance battle")
	}
	current.ChibiType = stance
	return "", nil
}

func (c *ChibiActor) SetFacing(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("try something like !chibi face back")
	}
	facing, err := spine.ChibiFacingEnum_Parse(args[2])
	if err != nil {
		return "", errors.New("try something like !chibi face back or !chibi face front")
	}
	if current.ChibiType == spine.CHIBI_TYPE_ENUM_BASE && facing == spine.CHIBI_FACING_ENUM_BACK {
		return "", errors.New("base chibi's can't face backwards. Try setting to battle stance first")
	}
	current.Facing = facing
	return "", nil
}

func (c *ChibiActor) SetEnemy(args []string, current *spine.OperatorInfo) (string, error) {
	errMsg := errors.New("try something like !chibi enemy <enemyname or ID> (ie. !chibi enemy Avenger, !chibi enemy SM8")
	if len(args) < 3 {
		return "", errMsg
	}
	trimmed := strings.Join(args[2:], " ")

	humanOperatorName := strings.TrimSpace(trimmed)
	operatorId, matches := c.client.GetOperatorIdFromName(humanOperatorName, spine.FACTION_ENUM_ENEMY)
	if matches != nil {
		return "", errors.New("")
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
	moveAnimation := spine.DEFAULT_MOVE_ANIM_NAME
	if !slices.Contains(current.Animations, moveAnimation) {
		for _, animation := range current.Animations {
			if strings.Contains(animation, spine.DEFAULT_MOVE_ANIM_NAME) {
				moveAnimation = animation
				break
			}
		}
	}
	current.CurrentAnimations = []string{moveAnimation}

	if len(args) == 3 {
		errMsg := errors.New("try something like !chibi walk 0.45")
		desiredPosition, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return "", errMsg
		}
		if (desiredPosition < 0.0) || (desiredPosition > 1.0) {
			return "", errMsg
		}
		current.TargetPos = misc.NewOption(misc.Vector2{
			X: desiredPosition, Y: 0.0,
		})
	}
	return "", nil
}

func (c *ChibiActor) GetWhoInfo(args []string, current *spine.OperatorInfo) (string, error) {
	// !chibi who <name>
	if len(args) < 3 {
		return "", errors.New("try something like !chibi who steam knight")
	}
	chibiName := strings.Join(args[2:], " ")
	log.Printf("Searching for %s\n", chibiName)

	operatorId, operatorMatches := c.client.GetOperatorIdFromName(chibiName, spine.FACTION_ENUM_OPERATOR)
	enemyId, enemyMatches := c.client.GetOperatorIdFromName(chibiName, spine.FACTION_ENUM_ENEMY)

	opMat := make([]string, 0)
	if operatorMatches != nil {
		opMat = append(opMat, operatorMatches...)
	} else {
		resp, err := c.client.GetOperator(&spine.GetOperatorRequest{
			OperatorId: operatorId,
			Faction:    spine.FACTION_ENUM_OPERATOR,
		})
		if err != nil {
			return "", nil
		}
		opMat = append(opMat, resp.OperatorName)
	}

	enemyMat := make([]string, 0)
	if enemyMatches != nil {
		enemyMat = append(enemyMat, enemyMatches...)
	} else {
		resp, err := c.client.GetOperator(&spine.GetOperatorRequest{
			OperatorId: enemyId,
			Faction:    spine.FACTION_ENUM_ENEMY,
		})
		if err != nil {
			return "", nil
		}
		enemyMat = append(enemyMat, resp.OperatorName)
	}

	return fmt.Sprintf("Did you mean: %s or enemies %s", strings.Join(opMat, ", "), strings.Join(enemyMat, ", ")), nil
}

func (c *ChibiActor) SetChibiModel(trimmed string, current *spine.OperatorInfo) (string, error) {
	log.Printf("!chibi command triggered with %v\n", trimmed)
	args := strings.Split(trimmed, " ")
	errMsg := errors.New("try something like !chibi <name> (ie. !chibi Amiya, !chibi Lava Alter)")
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
		return "", errors.New("")
	}

	current.OperatorId = operatorId
	current.Skin = spine.DEFAULT_SKIN_NAME
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
	stanceMap, ok := operatorData.Skins[spine.DEFAULT_SKIN_NAME].Stances[spine.CHIBI_TYPE_ENUM_BASE]
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
		Operator: spine.OperatorInfo{
			OperatorId:        operatorId,
			Faction:           spine.FACTION_ENUM_OPERATOR,
			Skin:              spine.DEFAULT_SKIN_NAME,
			ChibiType:         chibiType,
			Facing:            spine.CHIBI_FACING_ENUM_FRONT,
			CurrentAnimations: []string{spine.GetDefaultAnimForChibiType(chibiType)},
		},
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
		msg = fmt.Sprintf("%s skins: %s", current.DisplayName, strings.Join(current.Skins, ", "))
	case "anims":
		msg = fmt.Sprintf("%s animations: %s", current.DisplayName, strings.Join(current.Animations, ","))
	case "info":
		msg = fmt.Sprintf(
			"%s: %s, %s, %s, (%s)",
			current.DisplayName,
			current.Skin,
			current.ChibiType,
			current.Facing,
			strings.Join(current.CurrentAnimations, ","),
		)
	default:
		return "", errors.New("incorrect usage: !chibi <skins|anims|info>")
	}
	return msg, nil
}
