package chat

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/stretchr/testify/assert"
)

/*
Manual Test Cases:
!chibi
!cibi
!chibi 01234567890123456789012345678901234567890123456789 01234567890123456789012345678901234567890123456789
!chibi help
!chibi skins
!chibi anims
!chibi info
!chibi who
!chibi who steam knight
!chibi who lakjsdlkjfsad
!chibi skin default
!chibi anim
!chibi anim asdjlsdf
!chibi anim relax
!chibi play sleep
!chibi play relax sleep
!chibi stance
!chibi stance base
!chibi stance battle
!chibi face
!chibi face back
!chibi face front
!chibi enemy
!chibi enemy The last steam knight
!chibi enemy Eblana
!chibi enemy B4
!chibi walk
!chibi walk 0
!chibi walk 1.0
!chibi walk 2
!chibi walk 0.7
!chibi speed
!chibi speed 0
!chibi speed 6
!chibi speed 3
!chibi speed 5.0
!chibi amiya
!chibi base
!chibi battle
!chibi amiya
!chibi base
!chibi walk
!chibi battle
!chibi size 1.5
!chibi scale 0.5
!chibi move_speed 160
!chibi pace 0.2 0.5

// Change from enemy to operator during a walk/wander
// and then transitioning to a battle stance would cause the operator
// to slide across the screen in their "idle" battle animations
!chibi enemy b2
!chibi walk
!chibi reed
!chibi battle

// Changing from enemy to operate during a walkto action was causing
// the operator to forever loop in their "move" animation once they
// reached their destination
!chibi enemy b2
!chibi walk 0.8
!chibi reed

// For Walkto actions.
// When walking to the same poisition as your startPosition this would
// cause the chibi to forever be in their "move" animation. Make sure
// they correctly transition into their "idle"
!chibi walk 0.5
*/

type FakeActorUpdater struct {
	opInfo operator.OperatorInfo
}

func (f *FakeActorUpdater) CurrentInfo(ctx context.Context, username string) (operator.OperatorInfo, error) {
	return f.opInfo, nil
}
func (f *FakeActorUpdater) UpdateChibi(ctx context.Context, userInfo misc.UserInfo, update *operator.OperatorInfo) error {
	return nil
}

func setupCommandTest() (*operator.OperatorInfo, ActorUpdater, *ChatCommandProcessor) {
	current := operator.NewOperatorInfo(
		"Amiya",
		operator.FACTION_ENUM_OPERATOR,
		"char_002_amiya",
		operator.DEFAULT_SKIN_NAME,
		operator.CHIBI_STANCE_ENUM_BASE,
		operator.CHIBI_FACING_ENUM_FRONT,
		[]string{operator.DEFAULT_SKIN_NAME, "skin1"},
		[]string{operator.DEFAULT_ANIM_BASE, operator.DEFAULT_ANIM_BATTLE, "anim1"},
		1.0,
		misc.EmptyOption[misc.Vector2](),
		operator.ACTION_PLAY_ANIMATION,
		operator.NewActionPlayAnimation([]string{operator.DEFAULT_ANIM_BASE}),
	)
	actor := &FakeActorUpdater{
		opInfo: current,
	}
	assetManager := operator.NewTestAssetService()
	spineService := operator.NewOperatorService(assetManager, misc.DefaultSpineRuntimeConfig())
	sut := &ChatCommandProcessor{
		spineService: spineService,
	}
	return &current, actor, sut
}

func TestCmdProcessorHandleMessage_NotChibiCommand(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "hello world",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
}

func TestCmdProcessorHandleMessage_TooLongCommand(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi 01234567890123456789012345678901234567890123456789 01234567890123456789012345678901234567890123456789",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
}
func TestCmdProcessorHandleMessage_NotChibiCommandExplicit(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibiextended",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
}
func TestCmdProcessorHandleMessage_ChibiOnlyCommandShowHelp(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "!chibi to control your Arknights chibi.")
}

func TestCmdProcessorHandleMessage_ChibiHelp(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi help",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "!chibi to control your Arknights chibi.")
}
func TestCmdProcessorHandleMessage_ChibiSkins(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi skins",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Amiya skins: default")
}

func TestCmdProcessorHandleMessage_ChibiAnims(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi anims",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Amiya animations: Move,Idle")
}

func TestCmdProcessorHandleMessage_ChibiInfo(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi info",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Amiya: default, base, Front, (Move)")
}

func TestCmdProcessorHandleMessage_ChibiWhoFail(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi who notanoperator",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Could not find any operators/enemies with that name")
}

func TestCmdProcessorHandleMessage_ChibiWhoSuccess(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi who amiy",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Did you mean")
}

func TestCmdProcessorHandleMessage_ChibiSetSkin(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi skin1",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal("skin1", current.Skin)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiPlayAnimation(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi anim1",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(operator.ACTION_PLAY_ANIMATION, current.CurrentAction)
	assert.Equal([]string{"anim1"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiPlayMultpleAnimation(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi play anim1 Idle",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(operator.ACTION_PLAY_ANIMATION, current.CurrentAction)
	assert.Equal([]string{"anim1", "Idle"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiFace(t *testing.T) {
	current, actor, sut := setupCommandTest()
	current.ChibiStance = operator.CHIBI_STANCE_ENUM_BATTLE

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi battle",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(operator.CHIBI_STANCE_ENUM_BATTLE, current.ChibiStance)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiSetFaceOnlyForBattleStance(t *testing.T) {
	current, _, sut := setupCommandTest()

	assert := assert.New(t)
	_, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi face back",
	})

	if assert.Error(err) {
		assert.ErrorContains(err, "base chibi's can't face backwards.")
	}
}

func TestCmdProcessorHandleMessage_ChibiSetFaceOnly(t *testing.T) {
	current, actor, sut := setupCommandTest()
	current.ChibiStance = operator.CHIBI_STANCE_ENUM_BATTLE

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi face back",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(operator.CHIBI_FACING_ENUM_BACK, current.Facing)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiEnemyNotEnoughArgs(t *testing.T) {
	// TODO: Need to add enemy to AssetService in order to test enemy happy path
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi enemy",
	})

	// assert.Nil(err)
	if assert.Error(err) {
		assert.ErrorContains(err, "try something like !chibi enemy <enemyname or ID>")
	}
	assert.Empty(cmd.Reply(actor))
	if _, ok := cmd.(*ChatCommandNoOp); !ok {
		assert.Fail("Command is not of type: ChatCommandNoOp")
	}
}

func TestCmdProcessorHandleMessage_ChibiEnemyHappyPath(t *testing.T) {
	current, actor, sut := setupCommandTest()
	current.MovementSpeed = misc.NewOption[misc.Vector2](
		misc.Vector2{X: 160, Y: 0},
	)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi enemy Slug",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal("enemy_1007_slime_2", current.OperatorId)
	assert.Equal(operator.FACTION_ENUM_ENEMY, current.Faction)
	assert.Equal(1.0, current.AnimationSpeed)
	assert.True(current.SpriteScale.IsNone())
	assert.Equal(160.0, current.MovementSpeed.Unwrap().X)
}

func TestCmdProcessorHandleMessage_ChibiWalk(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi walk",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(operator.ACTION_WANDER, current.CurrentAction)
	assert.Equal([]string{"Move"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiWalkTo(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi walk 0.5",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(operator.ACTION_WALK_TO, current.CurrentAction)
	assert.Equal([]string{"Move", "Relax"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(0.5, current.Action.TargetPos.Unwrap().X)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiAnimationSpeed(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi speed 2.0",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(2.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiAnimationSpeedMaxSpeed(t *testing.T) {
	current, _, sut := setupCommandTest()

	assert := assert.New(t)
	_, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi speed 6.0",
	})

	assert.Nil(err)
	assert.Equal(sut.spineService.GetMaxAnimationSpeed(), current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiChibiModel(t *testing.T) {
	current, actor, sut := setupCommandTest()
	current.MovementSpeed = misc.NewOption[misc.Vector2](
		misc.Vector2{X: 160, Y: 0},
	)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi amiya",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal("char_002_amiya", current.OperatorId)
	assert.Equal(operator.FACTION_ENUM_OPERATOR, current.Faction)
	assert.Equal(1.0, current.AnimationSpeed)
	assert.True(current.SpriteScale.IsNone())
	assert.Equal(160.0, current.MovementSpeed.Unwrap().X)
}

func TestCmdProcessorHandleMessage_Regression_EnemyToOperatorShouldMaintainWanderAction(t *testing.T) {
	current, actor, sut := setupCommandTest()
	current.OperatorDisplayName = "Slug"
	current.OperatorId = "enemy_1007_slime_2"
	current.Faction = operator.FACTION_ENUM_ENEMY
	current.ChibiStance = operator.CHIBI_STANCE_ENUM_BATTLE
	current.CurrentAction = operator.ACTION_WANDER
	current.Action = operator.NewActionWander("Move")

	log.Printf("%v\n", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi amiya",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal("char_002_amiya", current.OperatorId)
	assert.Equal(operator.FACTION_ENUM_OPERATOR, current.Faction)
	assert.Equal(1.0, current.AnimationSpeed)
	assert.True(current.SpriteScale.IsNone())
	assert.True(current.MovementSpeed.IsNone())
	assert.Equal(operator.CHIBI_STANCE_ENUM_BASE, current.ChibiStance)
}

func TestCmdProcessorHandleMessage_ChibiScale(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi size 0.5",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(0.5, current.SpriteScale.Unwrap().X)
	assert.Equal(0.5, current.SpriteScale.Unwrap().Y)
}

func TestCmdProcessorHandleMessage_ChibiMoveSpeed(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi move_speed 2",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal(2.0, current.MovementSpeed.Unwrap().X)
	assert.Equal(0.0, current.MovementSpeed.Unwrap().Y)
}

func TestCmdProcessorHandleMessage_ChibiMoveSpeedOutOfRange(t *testing.T) {
	current, _, sut := setupCommandTest()

	assert := assert.New(t)
	_, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         fmt.Sprintf("!chibi move_speed %f", sut.spineService.GetMaxMovementSpeed()+1),
	})

	assert.Nil(err)
	assert.Equal(float64(sut.spineService.GetMaxMovementSpeed()), current.MovementSpeed.Unwrap().X)
}

func TestCmdProcessorHandleMessage_ChibiPaceAround(t *testing.T) {
	current, actor, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		TwitchUserId:    "100",
		Message:         "!chibi pace 0.1 0.5",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	assert.Equal(operator.ACTION_PACE_AROUND, current.CurrentAction)
	assert.Equal([]string{"Move"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(0.1, current.Action.PaceStartPos.Unwrap().X)
	assert.Equal(0.5, current.Action.PaceEndPos.Unwrap().X)
	assert.Equal(1.0, current.AnimationSpeed)
}
