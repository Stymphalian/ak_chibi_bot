package chat

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

type ChatCommandProcessor struct {
	spineService        *operator.OperatorService
	processChatMessages bool
}

func NewChatCommandProcessor(spineService *operator.OperatorService) *ChatCommandProcessor {
	return &ChatCommandProcessor{
		spineService:        spineService,
		processChatMessages: true,
	}
}

type ChatArgs struct {
	chatMsg *ChatMessage
	args    []string
}

func (c *ChatCommandProcessor) SetProcessChatMessagesFlag(val bool) {
	c.processChatMessages = val
}

func (c *ChatCommandProcessor) HandleMessage(current *operator.OperatorInfo, chatMsg ChatMessage) (ChatCommand, error) {
	if !strings.HasPrefix(chatMsg.Message, "!chibi") {
		if c.processChatMessages {
			return c.ShowChatMessage(&chatMsg)
		} else {
			return &ChatCommandNoOp{}, nil
		}
	}
	if len(chatMsg.Message) >= 100 {
		return &ChatCommandNoOp{}, nil
	}
	misc.Monitor.NumCommands += 1

	args := strings.Split(chatMsg.Message, " ")
	args = slices.DeleteFunc(args, func(s string) bool { return s == "" })
	if len(args) == 1 && args[0] == "!chibi" {
		chatArgs := &ChatArgs{
			chatMsg: &chatMsg,
			args:    nil,
		}
		return c.chibiHelp(chatArgs)
	}
	if args[0] != "!chibi" {
		return &ChatCommandNoOp{}, nil
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
	// !chibi speed 0.1
	// !chibi size 0.5 [0.1 1.5]
	// !chibi scale 0.5 [0.1 1.5]
	// !chibi velocity 2.0 [0.1 1.5]
	// !chibi pace 0.1 0.5
	// !chibi move_speed default
	// !chibi save
	// !chibi unsave
	// !chibi follow stymtwitchbot

	// var msg string
	subCommand := strings.TrimSpace(args[1])
	chatArgs := &ChatArgs{
		chatMsg: &chatMsg,
		args:    args,
	}
	switch subCommand {
	// case "admin":
	// 	if userName != c.botConfig.Broadcaster {
	// 		log.Printf("Only broadcaster can use !chibi admin: %s\n", userName)
	// 		return "", nil
	// 	}
	// 	// !chibi admin <username> "!chibi command"
	// 	if len(args) < 3 {
	// 		log.Printf("Only broadcaster can use !chibi admin not enough args %v\n", args)
	// 		return "", nil
	// 	}
	// 	splitStr := strings.SplitN(trimmed, " ", 4)
	// 	targetUser := splitStr[2]
	// 	if len(splitStr) == 4 {
	// 		restCommand := splitStr[3]
	// 		return c.HandleCommand(targetUser, targetUser, restCommand)
	// 	}
	// 	return "", nil
	case "help":
		return c.chibiHelp(chatArgs)
	case "skins":
		return c.getChibiInfo(chatArgs, "skins")
	case "anims":
		return c.getChibiInfo(chatArgs, "anims")
	case "info":
		return c.getChibiInfo(chatArgs, "info")
	case "who":
		return c.getWhoInfo(chatArgs)
	case "skin":
		return c.setSkin(chatArgs, current)
	case "anim":
		return c.setAnimation(chatArgs, current)
	case "play":
		return c.setAnimation(chatArgs, current)
	case "stance":
		return c.setStance(chatArgs, current)
	case "face":
		return c.setFacing(chatArgs, current)
	case "enemy":
		return c.setEnemy(chatArgs, current)
	case "walk":
		return c.setWalk(chatArgs, current)
	case "pace":
		return c.setPace(chatArgs, current)
	case "follow":
		return c.setFollow(chatArgs, current)
	case "speed":
		return c.setAnimationSpeed(chatArgs, current)
	case "size":
		return c.setScale(chatArgs, current)
	case "scale":
		return c.setScale(chatArgs, current)
	case "move_speed":
		return c.setMoveSpeed(chatArgs, current)
	case "velocity":
		return c.setMoveSpeed(chatArgs, current)
	case "save":
		return c.setSaveUserPrefs(chatArgs, current)
	case "unsave":
		return c.setClearUserPrefs(chatArgs, current)
	default:
		if _, ok := misc.MatchesKeywords(subCommand, current.AvailableAnimations); ok {
			chatArgs.args = []string{"!chibi", "play", subCommand}
			return c.setAnimation(chatArgs, current)
		} else if _, ok := misc.MatchesKeywords(subCommand, current.Skins); ok {
			chatArgs.args = []string{"!chibi", "skin", subCommand}
			return c.setSkin(chatArgs, current)
		} else if _, ok := misc.MatchesKeywords(subCommand, []string{"base", "battle"}); ok {
			// []string{"!chibi", "skin", subCommand}
			chatArgs.args = []string{"!chibi", "stance", subCommand}
			return c.setStance(chatArgs, current)
		} else {
			// matches against operator names
			return c.setChibiModel(chatArgs, current)
		}
	}
}

func (c *ChatCommandProcessor) chibiHelp(args *ChatArgs) (ChatCommand, error) {
	log.Printf("!chibi_help command triggered with %v\n", args.chatMsg.Message)
	msg := `!chibi to control your Arknights chibi. ` +
		`"!chibi Rockrock" to change your operator. ` +
		`"!chibi play Move", "!chibi skin epoque#2" to change the animation and skin. ` +
		`"!chibi skins" and "!chibi anims" lists available skins and animations. ` +
		`"!chibi stance battle" to change from base or battle chibis. ` +
		`"!chibi enemy mandragora" to change into an enemy mob instead of an operator. ` +
		`"!chibi walk" to have your chibi walk around the screen. ` +
		`Source Code from github: search Stymphalian/ak_chibi_bot`
	return &ChatCommandSimpleMessage{replyMessage: msg}, nil
}

func (c *ChatCommandProcessor) setSkin(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, nil
	}

	skinName, hasSkin := misc.MatchesKeywords(args.args[2], current.Skins)
	if !hasSkin {
		return &ChatCommandNoOp{}, nil
	}
	current.Skin = skinName
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()
	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setAnimation(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, nil
	}

	animation, ok := misc.MatchesKeywords(args.args[2], current.AvailableAnimations)
	if !ok {
		return &ChatCommandNoOp{}, nil
	}
	current.CurrentAction = operator.ACTION_PLAY_ANIMATION
	current.Action = operator.NewActionPlayAnimation([]string{animation})

	if len(args.args) >= 4 {
		animations := make([]string, 0)
		skipAdding := false
		for i := 2; i < len(args.args); i++ {
			anim, ok := misc.MatchesKeywords(args.args[i], current.AvailableAnimations)
			if !ok {
				skipAdding = true
				break
			}
			animations = append(animations, anim)
		}
		if !skipAdding {
			current.Action = operator.NewActionPlayAnimation(animations)
		}
	}

	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setStance(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi stance battle")
	}
	stance, err := operator.ChibiStanceEnum_Parse(args.args[2])
	if err != nil {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi stance battle")
	}
	current.ChibiStance = stance
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()

	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setFacing(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi face back")
	}
	facing, err := operator.ChibiFacingEnum_Parse(args.args[2])
	if err != nil {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi face back or !chibi face front")
	}
	if current.ChibiStance == operator.CHIBI_STANCE_ENUM_BASE && facing == operator.CHIBI_FACING_ENUM_BACK {
		return &ChatCommandNoOp{}, errors.New("base chibi's can't face backwards. Try setting to battle stance first")
	}
	current.Facing = facing
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()
	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setEnemy(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	errMsg := errors.New("try something like !chibi enemy <enemyname or ID> (ie. !chibi enemy Avenger, !chibi enemy SM8")
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, errMsg
	}
	trimmed := strings.Join(args.args[2:], " ")

	mobName := strings.TrimSpace(trimmed)
	operatorId, matches := c.spineService.GetOperatorIdFromName(mobName, operator.FACTION_ENUM_ENEMY)
	if matches != nil {
		return &ChatCommandNoOp{}, nil
	}
	current.OperatorId = operatorId
	current.Faction = operator.FACTION_ENUM_ENEMY
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()
	current.SpriteScale = misc.EmptyOption[misc.Vector2]()
	// current.MovementSpeed = misc.EmptyOption[misc.Vector2]()

	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) getMoveAnimFromCurrent(current *operator.OperatorInfo) string {
	currentAnimations := current.Action.GetAnimations(current.CurrentAction)
	moveAnimation := operator.DEFAULT_MOVE_ANIM_NAME
	for _, animation := range currentAnimations {
		if strings.Contains(animation, "Move") {
			moveAnimation = animation
			break
		}
	}
	if !slices.Contains(current.AvailableAnimations, moveAnimation) {
		for _, animation := range current.AvailableAnimations {
			if strings.Contains(animation, "Move") {
				moveAnimation = animation
				break
			}
		}
	}
	return moveAnimation
}

func (c *ChatCommandProcessor) setWalk(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if current.Faction == operator.FACTION_ENUM_OPERATOR {
		current.ChibiStance = operator.CHIBI_STANCE_ENUM_BASE
	}

	// Set the animation to "Move". If "Move" doesn't exist in the list of
	// animations then try to find an animation with "Move" in its name
	// Try to keep the current animation if it is already a "move" like animation
	moveAnimation := c.getMoveAnimFromCurrent(current)
	current.CurrentAction = operator.ACTION_WANDER
	current.Action = operator.NewActionWander(moveAnimation)
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()

	if len(args.args) == 3 {
		errMsg := errors.New("try something like !chibi walk 0.45")
		desiredPosition, err := strconv.ParseFloat(args.args[2], 64)
		if err != nil {
			return &ChatCommandNoOp{}, errMsg
		}
		if desiredPosition < 0.0 {
			desiredPosition = 0
		} else if desiredPosition > 1.0 {
			desiredPosition = 1.0
		}

		current.CurrentAction = operator.ACTION_WALK_TO
		animationAfterStance := ""
		if current.ChibiStance == operator.CHIBI_STANCE_ENUM_BASE {
			animationAfterStance = operator.DEFAULT_ANIM_BASE_RELAX
		} else {
			animationAfterStance = operator.DEFAULT_ANIM_BATTLE
		}

		current.Action = operator.NewActionWalkTo(
			misc.Vector2{X: desiredPosition, Y: 0.0},
			moveAnimation,
			animationAfterStance,
		)
		current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()
	} else if len(args.args) == 4 {
		startPos, err := strconv.ParseFloat(args.args[2], 64)
		if err != nil {
			return &ChatCommandNoOp{}, errors.New("try something like !chibi walk 0.1 0.5")
		}
		endPos, err := strconv.ParseFloat(args.args[3], 64)
		if err != nil {
			return &ChatCommandNoOp{}, errors.New("try something like !chibi walk 0.1 0.5")
		}
		if startPos < 0 || startPos > 1 || endPos < 0 || endPos > 1 {
			return &ChatCommandNoOp{}, errors.New("try something like !chibi walk 0.1 0.5")
		}

		moveAnimation := c.getMoveAnimFromCurrent(current)
		current.CurrentAction = operator.ACTION_PACE_AROUND
		current.Action = operator.NewActionPaceAround(
			misc.Vector2{X: startPos, Y: 0},
			misc.Vector2{X: endPos, Y: 0},
			moveAnimation,
		)
		current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()
	}
	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setAnimationSpeed(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi speed 0.5")
	}
	animationSpeed, err := strconv.ParseFloat(args.args[2], 64)
	if err != nil {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi speed 1.5")
	}
	if animationSpeed <= c.spineService.GetMinAnimationSpeed() {
		animationSpeed = c.spineService.GetMinAnimationSpeed()
	} else if animationSpeed > c.spineService.GetMaxAnimationSpeed() {
		animationSpeed = c.spineService.GetMaxAnimationSpeed()
	}
	current.AnimationSpeed = animationSpeed
	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) getWhoInfo(args *ChatArgs) (ChatCommand, error) {
	// !chibi who <name>
	if len(args.args) < 3 {
		return &ChatCommandSimpleMessage{}, errors.New("try something like !chibi who steam knight")
	}
	chibiName := strings.Join(args.args[2:], " ")
	log.Printf("Searching for %s\n", chibiName)

	operatorId, operatorMatches := c.spineService.GetOperatorIdFromName(chibiName, operator.FACTION_ENUM_OPERATOR)
	enemyId, enemyMatches := c.spineService.GetOperatorIdFromName(chibiName, operator.FACTION_ENUM_ENEMY)

	opMat := make([]string, 0)
	if operatorMatches != nil {
		opMat = append(opMat, operatorMatches...)
	} else {
		resp, err := c.spineService.GetOperator(operatorId, operator.FACTION_ENUM_OPERATOR)
		if err != nil {
			return &ChatCommandNoOp{}, nil
		}
		opMat = append(opMat, resp.OperatorName)
	}

	enemyMat := make([]string, 0)
	if enemyMatches != nil {
		enemyMat = append(enemyMat, enemyMatches...)
	} else {
		resp, err := c.spineService.GetOperator(enemyId, operator.FACTION_ENUM_ENEMY)
		if err != nil {
			return &ChatCommandNoOp{}, nil
		}
		enemyMat = append(enemyMat, resp.OperatorName)
	}

	var msg string
	if len(opMat) == 0 && len(enemyMat) == 0 {
		msg = "Could not find any operators/enemies with that name"
	} else if len(opMat) == 0 {
		msg = fmt.Sprintf("Did you mean enemies: %s", strings.Join(enemyMat, ", "))
	} else if len(enemyMat) == 0 {
		msg = fmt.Sprintf("Did you mean opeators: %s", strings.Join(opMat, ", "))
	} else {
		msg = fmt.Sprintf("Did you mean: %s or enemies %s", strings.Join(opMat, ", "), strings.Join(enemyMat, ", "))
	}
	return &ChatCommandSimpleMessage{replyMessage: msg}, nil
}

func (c *ChatCommandProcessor) setChibiModel(chatArgs *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	trimmed := chatArgs.chatMsg.Message
	log.Printf("!chibi command triggered with %v\n", trimmed)
	args := strings.Split(trimmed, " ")
	errMsg := errors.New("try something like !chibi <name> (ie. !chibi Amiya, !chibi Lava Alter)")
	if len(args) < 2 {
		return &ChatCommandNoOp{}, errMsg
	}

	splitStrs := strings.SplitN(trimmed, " ", 2)
	if len(splitStrs) != 2 {
		return &ChatCommandNoOp{}, errMsg
	}
	humanOperatorName := strings.TrimSpace(splitStrs[1])
	operatorId, matches := c.spineService.GetOperatorIdFromName(humanOperatorName, operator.FACTION_ENUM_OPERATOR)
	if matches != nil {
		return &ChatCommandNoOp{}, nil
	}

	prevFaction := current.Faction
	if prevFaction == operator.FACTION_ENUM_ENEMY {
		// If changing from an enemy to an operator and the current action is
		// "walking/wander" we need to set the stance to base in order to
		// make the chibi continue to walk
		if operator.IsWalkingAction(current.CurrentAction) {
			current.ChibiStance = operator.CHIBI_STANCE_ENUM_BASE
		}
	}

	current.OperatorId = operatorId
	current.Faction = operator.FACTION_ENUM_OPERATOR
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()
	current.SpriteScale = misc.EmptyOption[misc.Vector2]()
	// current.MovementSpeed = misc.EmptyOption[misc.Vector2]()
	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        chatArgs.chatMsg.Username,
		usernameDisplay: chatArgs.chatMsg.UserDisplayName,
		twitchUserId:    chatArgs.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) getChibiInfo(args *ChatArgs, subInfoName string) (ChatCommand, error) {
	return &ChatCommandInfo{
		info:     subInfoName,
		username: args.chatMsg.Username,
	}, nil
}

func (c *ChatCommandProcessor) setScale(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi size 0.5")
	}
	spriteScale, err := strconv.ParseFloat(args.args[2], 64)
	if err != nil {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi size 1.5")
	}
	if spriteScale < c.spineService.GetMinScaleSize() {
		spriteScale = c.spineService.GetMinScaleSize()
	} else if spriteScale > c.spineService.GetMaxScaleSize() {
		spriteScale = c.spineService.GetMaxScaleSize()
	}
	current.SpriteScale = misc.NewOption(
		misc.Vector2{X: spriteScale, Y: spriteScale},
	)
	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setMoveSpeed(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 3 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi move_speed 120")
	}

	if args.args[2] == "default" {
		current.MovementSpeed = misc.EmptyOption[misc.Vector2]()
	} else {
		moveSpeed, err := strconv.ParseFloat(args.args[2], 64)
		if err != nil {
			return &ChatCommandNoOp{}, errors.New("try something like !chibi move_speed 360")
		}
		if moveSpeed < c.spineService.GetMinMovementSpeed() {
			moveSpeed = c.spineService.GetMinMovementSpeed()
		} else if moveSpeed > c.spineService.GetMaxMovementSpeed() {
			moveSpeed = c.spineService.GetMaxMovementSpeed()
		}
		current.MovementSpeed = misc.NewOption(
			misc.Vector2{X: moveSpeed, Y: moveSpeed},
		)
	}

	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setPace(args *ChatArgs, current *operator.OperatorInfo) (ChatCommand, error) {
	if len(args.args) < 4 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi pace 0.1 0.5")
	}
	if current.Faction == operator.FACTION_ENUM_OPERATOR {
		current.ChibiStance = operator.CHIBI_STANCE_ENUM_BASE
	}

	startPos, err := strconv.ParseFloat(args.args[2], 64)
	if err != nil {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi pace 0.1 0.5")
	}
	endPos, err := strconv.ParseFloat(args.args[3], 64)
	if err != nil {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi pace 0.1 0.5")
	}
	if startPos < 0 || startPos > 1 || endPos < 0 || endPos > 1 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi pace 0.1 0.5")
	}

	moveAnimation := c.getMoveAnimFromCurrent(current)
	current.CurrentAction = operator.ACTION_PACE_AROUND
	current.Action = operator.NewActionPaceAround(
		misc.Vector2{X: startPos, Y: 0},
		misc.Vector2{X: endPos, Y: 0},
		moveAnimation,
	)
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()

	return &ChatCommandUpdateActor{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setSaveUserPrefs(
	args *ChatArgs,
	current *operator.OperatorInfo,
) (ChatCommand, error) {
	if len(args.args) != 2 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi save")
	}
	return &ChatCommandSavePrefs{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		action:          ChatCommandSaveChibi_Save,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) setClearUserPrefs(
	args *ChatArgs,
	_ *operator.OperatorInfo,
) (ChatCommand, error) {
	if len(args.args) != 2 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi unsave")
	}
	return &ChatCommandSavePrefs{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		action:          ChatCommandSaveChibi_Remove,
		update:          nil,
	}, nil
}

func (c *ChatCommandProcessor) setFollow(
	args *ChatArgs,
	current *operator.OperatorInfo,
) (ChatCommand, error) {
	if len(args.args) != 3 {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi follow <username>")
	}
	usernameTarget := strings.ToLower(args.args[2])
	if misc.ValidateChannelName(usernameTarget) != nil ||
		usernameTarget == strings.ToLower(args.chatMsg.Username) {
		return &ChatCommandNoOp{}, errors.New("try something like !chibi follow <username>")
	}

	if current.Faction == operator.FACTION_ENUM_OPERATOR {
		current.ChibiStance = operator.CHIBI_STANCE_ENUM_BASE
	}
	animationAfterStance := ""
	if current.ChibiStance == operator.CHIBI_STANCE_ENUM_BASE {
		animationAfterStance = operator.DEFAULT_ANIM_BASE_RELAX
	} else {
		animationAfterStance = operator.DEFAULT_ANIM_BATTLE
	}

	moveAnimation := c.getMoveAnimFromCurrent(current)
	current.CurrentAction = operator.ACTION_FOLLOW
	current.Action = operator.NewActionFollow(usernameTarget, moveAnimation, animationAfterStance)
	current.AnimationSpeed = c.spineService.GetDefaultAnimationSpeed()

	return &ChatCommandFollow{
		replyMessage:    "",
		username:        args.chatMsg.Username,
		usernameDisplay: args.chatMsg.UserDisplayName,
		twitchUserId:    args.chatMsg.TwitchUserId,
		update:          current,
	}, nil
}

func (c *ChatCommandProcessor) ShowChatMessage(chatMsg *ChatMessage) (ChatCommand, error) {
	return &ChatCommandShowMessage{
		replyMessage:    "",
		username:        chatMsg.Username,
		usernameDisplay: chatMsg.UserDisplayName,
		twitchUserId:    chatMsg.TwitchUserId,
		message:         chatMsg.Message,
	}, nil
}
