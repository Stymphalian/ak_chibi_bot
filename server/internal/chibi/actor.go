package chibi

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

const (
	MIN_ANIMATION_SPEED     = 0.1
	DEFAULT_ANIMATION_SPEED = 1.0
	MAX_ANIMATION_SPEED     = 5.0
)

type ChibiActor struct {
	client       spine.SpineClient
	excludeNames []string
}

func NewChibiActor(client spine.SpineClient, excludeNames []string) *ChibiActor {
	return &ChibiActor{client: client, excludeNames: excludeNames}
}

func (c *ChibiActor) GiveChibiToUser(userName string, userNameDisplay string) error {
	// Skip giving chibis to these Users
	if slices.Contains(c.excludeNames, userName) {
		return nil
	}

	_, err := c.addRandomChibi(userName, userNameDisplay)
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

func (c *ChibiActor) HandleCommand(userName string, userNameDisplay string, trimmed string) (string, error) {
	if !strings.HasPrefix(trimmed, "!chibi") {
		return "", nil
	}
	if len(trimmed) >= 100 {
		return "", nil
	}
	misc.Monitor.NumCommands += 1

	args := strings.Split(trimmed, " ")
	args = slices.DeleteFunc(args, func(s string) bool { return s == "" })
	if len(args) == 1 && args[0] == "!chibi" {
		return c.chibiHelp(trimmed)
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
	// !chibi speed 0.1

	current, err := c.client.CurrentInfo(userName)
	if err != nil {
		switch err.(type) {
		case *spine.UserNotFound:
			log.Println("Chibi not found for user ", userName)
		}
		return "", errors.New("")
	}

	var msg string
	arg2 := strings.TrimSpace(args[1])
	switch arg2 {
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
		return c.chibiHelp(trimmed)
	case "skins":
		return c.getChibiInfo(userName, "skins")
	case "anims":
		return c.getChibiInfo(userName, "anims")
	case "info":
		return c.getChibiInfo(userName, "info")
	case "who":
		return c.getWhoInfo(args, &current)
	case "skin":
		msg, err = c.setSkin(args, &current)
	case "anim":
		msg, err = c.setAnimation(args, &current)
	case "play":
		msg, err = c.setAnimation(args, &current)
	case "stance":
		msg, err = c.setStance(args, &current)
	case "face":
		msg, err = c.setFacing(args, &current)
	case "enemy":
		msg, err = c.setEnemy(args, &current)
	case "walk":
		msg, err = c.setWalk(args, &current)
	case "speed":
		msg, err = c.setAnimationSpeed(args, &current)
	default:
		if _, ok := misc.MatchesKeywords(arg2, current.AvailableAnimations); ok {
			msg, err = c.setAnimation([]string{"!chibi", "play", arg2}, &current)
		} else if _, ok := misc.MatchesKeywords(arg2, current.Skins); ok {
			msg, err = c.setSkin([]string{"!chibi", "skin", arg2}, &current)
		} else if _, ok := misc.MatchesKeywords(arg2, []string{"base", "battle"}); ok {
			msg, err = c.setStance([]string{"!chibi", "stance", arg2}, &current)
		} else {
			// matches against operator names
			msg, err = c.setChibiModel(trimmed, &current)
		}
	}

	if err == nil {
		c.UpdateChibi(userName, userNameDisplay, &current)
	}

	return msg, err
}

func getValidAnimations(availableAnimations []string, actionAnimations []string, defaultAnim string) []string {
	for _, anim := range actionAnimations {
		if !slices.Contains(availableAnimations, anim) {
			actionAnimations = []string{defaultAnim}
			break
		}
	}
	// If it still doesn't exist then just choose one randomly
	for _, anim := range actionAnimations {
		if !slices.Contains(availableAnimations, anim) {
			actionAnimations = []string{availableAnimations[0]}
			break
		}
	}
	return actionAnimations
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
	update.OperatorDisplayName = currentOp.OperatorName

	if _, ok := currentOp.Skins[update.Skin]; !ok {
		update.Skin = spine.DEFAULT_SKIN_NAME
	}
	facings := currentOp.Skins[update.Skin].Stances[update.ChibiStance]
	if len(facings.Facings) == 0 {
		switch update.ChibiStance {
		case spine.CHIBI_STANCE_ENUM_BASE:
			update.ChibiStance = spine.CHIBI_STANCE_ENUM_BATTLE
		case spine.CHIBI_STANCE_ENUM_BATTLE:
			update.ChibiStance = spine.CHIBI_STANCE_ENUM_BASE
		default:
			update.ChibiStance = spine.CHIBI_STANCE_ENUM_BASE
		}
	}
	if _, ok := currentOp.Skins[update.Skin].Stances[update.ChibiStance].Facings[update.Facing]; !ok {
		update.Facing = spine.CHIBI_FACING_ENUM_FRONT
	}

	update.Skins = currentOp.GetSkinNames()

	// Validate animations
	update.AvailableAnimations = currentOp.Skins[update.Skin].Stances[update.ChibiStance].Facings[update.Facing]
	update.AvailableAnimations = spine.FilterAnimations(update.AvailableAnimations)

	// Validate animationSpeed
	if update.AnimationSpeed == 0 {
		update.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	}
	update.AnimationSpeed = misc.ClampF64(
		update.AnimationSpeed,
		MIN_ANIMATION_SPEED,
		MAX_ANIMATION_SPEED,
	)

	// Validate startPos
	if update.StartPos.IsSome() {
		vec := update.StartPos.Unwrap()
		if vec.X < 0 || vec.X > 1.0 || vec.Y < 0 || vec.Y > 1.0 {
			update.StartPos = misc.NewOption(
				misc.Vector2{
					X: misc.ClampF64(vec.X, 0, 1.0),
					Y: misc.ClampF64(vec.Y, 0, 1.0),
				},
			)
		}
	}

	// Validate actions
	switch update.CurrentAction {
	case spine.ACTION_PLAY_ANIMATION:
		update.Action.Animations = getValidAnimations(
			update.AvailableAnimations,
			update.Action.Animations,
			spine.GetDefaultAnimForChibiStance(update.ChibiStance),
		)
	case spine.ACTION_WANDER:
		update.Action.WanderAnimation = getValidAnimations(
			update.AvailableAnimations,
			[]string{update.Action.WanderAnimation},
			spine.GetDefaultAnimForChibiStance(update.ChibiStance),
		)[0]
	case spine.ACTION_WALK_TO:
		if update.Action.TargetPos.IsNone() {
			update.Action.TargetPos = misc.NewOption(misc.Vector2{X: 0.5, Y: 0.5})
		} else {
			vec := update.Action.TargetPos.Unwrap()
			update.Action.TargetPos = misc.NewOption(
				misc.Vector2{
					X: misc.ClampF64(vec.X, 0, 1.0),
					Y: misc.ClampF64(vec.Y, 0, 1.0),
				},
			)
			availableAnimations := update.AvailableAnimations
			defaultAnimation := spine.GetDefaultAnimForChibiStance(update.ChibiStance)
			update.Action.WalkToAnimation = getValidAnimations(
				availableAnimations,
				[]string{update.Action.WalkToAnimation},
				defaultAnimation,
			)[0]
			update.Action.WalkToFinalAnimation = getValidAnimations(
				availableAnimations,
				[]string{update.Action.WalkToFinalAnimation},
				defaultAnimation,
			)[0]
		}
	}

	return nil
}

func (c *ChibiActor) UpdateChibi(username string, usernameDisplay string, update *spine.OperatorInfo) error {
	c.validateUpdateSetDefaultOtherwise(update)

	_, err := c.client.SetOperator(
		&spine.SetOperatorRequest{
			UserName:        username,
			UserNameDisplay: usernameDisplay,
			Operator:        *update,
		})
	if err != nil {
		log.Printf("Failed to set chibi (%s)", err.Error())
		return nil
	}
	return nil
}

func (c *ChibiActor) chibiHelp(trimmed string) (string, error) {
	log.Printf("!chibi_help command triggered with %v\n", trimmed)
	msg := `!chibi to control your Arknights chibi. ` +
		`"!chibi Rockrock" to change your operator. ` +
		`"!chibi play Move", "!chibi skin epoque#2" to change the animation and skin. ` +
		`"!chibi skins" and "!chibi anims" lists available skins and animations. ` +
		`"!chibi stance battle" to change from base or battle chibis. ` +
		`"!chibi enemy mandragora" to change into an enemy mob instead of an operator. ` +
		`"!chibi walk" to have your chibi walk around the screen. ` +
		`Source Code from github: search Stymphalian/ak_chibi_bot`
	return msg, nil
}

func (c *ChibiActor) setSkin(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("")
	}

	skinName, hasSkin := misc.MatchesKeywords(args[2], current.Skins)
	if !hasSkin {
		return "", errors.New("")
	}
	current.Skin = skinName
	current.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	return "", nil
}

func (c *ChibiActor) setAnimation(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("")
	}

	animation, ok := misc.MatchesKeywords(args[2], current.AvailableAnimations)
	if !ok {
		return "", errors.New("")
	}
	current.CurrentAction = spine.ACTION_PLAY_ANIMATION
	current.Action = spine.NewActionPlayAnimation([]string{animation})

	if len(args) >= 4 {
		animations := make([]string, 0)
		skipAdding := false
		for i := 2; i < len(args); i++ {
			anim, ok := misc.MatchesKeywords(args[i], current.AvailableAnimations)
			if !ok {
				skipAdding = true
				break
			}
			animations = append(animations, anim)
		}
		if !skipAdding {
			current.Action = spine.NewActionPlayAnimation(animations)
		}
	}
	return "", nil
}

func (c *ChibiActor) setStance(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("try something like !chibi stance battle")
	}
	stance, err := spine.ChibiStanceEnum_Parse(args[2])
	if err != nil {
		return "", errors.New("try something like !chibi stance battle")
	}
	current.ChibiStance = stance
	current.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	return "", nil
}

func (c *ChibiActor) setFacing(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("try something like !chibi face back")
	}
	facing, err := spine.ChibiFacingEnum_Parse(args[2])
	if err != nil {
		return "", errors.New("try something like !chibi face back or !chibi face front")
	}
	if current.ChibiStance == spine.CHIBI_STANCE_ENUM_BASE && facing == spine.CHIBI_FACING_ENUM_BACK {
		return "", errors.New("base chibi's can't face backwards. Try setting to battle stance first")
	}
	current.Facing = facing
	current.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	return "", nil
}

func (c *ChibiActor) setEnemy(args []string, current *spine.OperatorInfo) (string, error) {
	errMsg := errors.New("try something like !chibi enemy <enemyname or ID> (ie. !chibi enemy Avenger, !chibi enemy SM8")
	if len(args) < 3 {
		return "", errMsg
	}
	trimmed := strings.Join(args[2:], " ")

	mobName := strings.TrimSpace(trimmed)
	operatorId, matches := c.client.GetOperatorIdFromName(mobName, spine.FACTION_ENUM_ENEMY)
	if matches != nil {
		return "", errors.New("")
	}
	current.OperatorId = operatorId
	current.Faction = spine.FACTION_ENUM_ENEMY
	current.AnimationSpeed = DEFAULT_ANIMATION_SPEED

	return "", nil
}

func (c *ChibiActor) setWalk(args []string, current *spine.OperatorInfo) (string, error) {
	current.ChibiStance = spine.CHIBI_STANCE_ENUM_BASE

	// Set the animation to "Move". If "Move" doesn't exist in the list of
	// animations then try to find an animation with "Move" in its name
	// Try to keep the current animation if it is already a "move" like animation
	currentAnimations := current.Action.GetAnimations(current.CurrentAction)
	moveAnimation := spine.DEFAULT_MOVE_ANIM_NAME
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
	current.CurrentAction = spine.ACTION_WANDER
	current.Action = spine.NewActionWander(moveAnimation)
	current.AnimationSpeed = DEFAULT_ANIMATION_SPEED

	if len(args) == 3 {
		errMsg := errors.New("try something like !chibi walk 0.45")
		desiredPosition, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return "", errMsg
		}
		if (desiredPosition < 0.0) || (desiredPosition > 1.0) {
			return "", errMsg
		}

		current.CurrentAction = spine.ACTION_WALK_TO
		animationAfterStance := ""
		if current.ChibiStance == spine.CHIBI_STANCE_ENUM_BASE {
			animationAfterStance = spine.DEFAULT_ANIM_BASE_RELAX
		} else {
			animationAfterStance = spine.DEFAULT_ANIM_BATTLE
		}

		current.Action = spine.NewActionWalkTo(
			misc.Vector2{X: desiredPosition, Y: 0.0},
			moveAnimation,
			animationAfterStance,
		)
		current.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	}
	return "", nil
}

func (c *ChibiActor) setAnimationSpeed(args []string, current *spine.OperatorInfo) (string, error) {
	if len(args) < 3 {
		return "", errors.New("try something like !chibi speed 0.5")
	}
	animationSpeed, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return "", errors.New("try something like !chibi speed 1.5")
	}
	if animationSpeed <= 0 || animationSpeed > MAX_ANIMATION_SPEED {
		return "", errors.New("try something like !chibi speed 2.0")
	}
	current.AnimationSpeed = animationSpeed
	return "", nil
}

func (c *ChibiActor) getWhoInfo(args []string, current *spine.OperatorInfo) (string, error) {
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
		resp, err := c.client.GetOperator(
			&spine.GetOperatorRequest{
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

func (c *ChibiActor) setChibiModel(trimmed string, current *spine.OperatorInfo) (string, error) {
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
	current.Faction = spine.FACTION_ENUM_OPERATOR
	current.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	return "", nil
}

func (c *ChibiActor) addRandomChibi(userName string, userNameDisplay string) (string, error) {
	operatorInfo, err := c.client.GetRandomOperator()
	if err != nil {
		return "", err
	}

	log.Printf("Giving %s the chibi %s\n", userName, operatorInfo.OperatorId)
	_, err = c.client.SetOperator(
		&spine.SetOperatorRequest{
			UserName:        userName,
			UserNameDisplay: userNameDisplay,
			Operator:        *operatorInfo,
		})
	return "", err
}

func (c *ChibiActor) getChibiInfo(userName string, subInfoName string) (string, error) {
	current, err := c.client.CurrentInfo(userName)
	if err != nil {
		return "", nil
	}

	var msg string
	switch subInfoName {
	case "skins":
		msg = fmt.Sprintf("%s skins: %s", current.OperatorDisplayName, strings.Join(current.Skins, ", "))
	case "anims":
		msg = fmt.Sprintf("%s animations: %s", current.OperatorDisplayName, strings.Join(current.AvailableAnimations, ","))
	case "info":
		currentAnimations := current.Action.GetAnimations(current.CurrentAction)
		msg = fmt.Sprintf(
			"%s: %s, %s, %s, (%s)",
			current.OperatorDisplayName,
			current.Skin,
			current.ChibiStance,
			current.Facing,
			strings.Join(currentAnimations, ","),
		)
	default:
		return "", errors.New("incorrect usage: !chibi <skins|anims|info>")
	}
	return msg, nil
}
