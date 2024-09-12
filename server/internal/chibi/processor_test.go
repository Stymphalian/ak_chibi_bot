package chibi

import (
	"testing"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
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
*/

func setupCommandTest() (*spine.OperatorInfo, *ChatCommandProcessor) {
	current := spine.NewOperatorInfo(
		"Amiya",
		spine.FACTION_ENUM_OPERATOR,
		"char_002_amiya",
		spine.DEFAULT_SKIN_NAME,
		spine.CHIBI_STANCE_ENUM_BASE,
		spine.CHIBI_FACING_ENUM_FRONT,
		[]string{spine.DEFAULT_SKIN_NAME, "skin1"},
		[]string{spine.DEFAULT_ANIM_BASE, spine.DEFAULT_ANIM_BATTLE, "anim1"},
		1.0,
		misc.EmptyOption[misc.Vector2](),
		spine.ACTION_PLAY_ANIMATION,
		spine.NewActionPlayAnimation([]string{spine.DEFAULT_ANIM_BASE}),
	)
	assetManager := spine.NewTestAssetService()
	spineService := spine.NewSpineService(assetManager)
	sut := &ChatCommandProcessor{
		spineService: spineService,
	}
	return &current, sut
}

func TestCmdProcessorHandleMessage_NotChibiCommand(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "hello world",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
}

func TestCmdProcessorHandleMessage_TooLongCommand(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi 01234567890123456789012345678901234567890123456789 01234567890123456789012345678901234567890123456789",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
}
func TestCmdProcessorHandleMessage_NotChibiCommandExplicit(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibiextended",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
}
func TestCmdProcessorHandleMessage_ChibiOnlyCommandShowHelp(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "!chibi to control your Arknights chibi.")
}

func TestCmdProcessorHandleMessage_ChibiHelp(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi help",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "!chibi to control your Arknights chibi.")
}
func TestCmdProcessorHandleMessage_ChibiSkins(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi skins",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Amiya skins: default")
}

func TestCmdProcessorHandleMessage_ChibiAnims(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi anims",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Amiya animations: Move,Idle")
}

func TestCmdProcessorHandleMessage_ChibiInfo(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi info",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Amiya: default, base, Front, (Move)")
}

func TestCmdProcessorHandleMessage_ChibiWhoFail(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi who notanoperator",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Could not find any operators/enemies with that name")
}

func TestCmdProcessorHandleMessage_ChibiWhoSuccess(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi who amiy",
	})

	assert.Nil(err)
	assert.Contains(cmd.Reply(actor), "Did you mean")
}

func TestCmdProcessorHandleMessage_ChibiSetSkin(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	assert.Equal(spine.ACTION_PLAY_ANIMATION, current.CurrentAction)
	assert.Equal([]string{"anim1"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiPlayMultpleAnimation(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	assert.Equal(spine.ACTION_PLAY_ANIMATION, current.CurrentAction)
	assert.Equal([]string{"anim1", "Idle"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiFace(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	current.ChibiStance = spine.CHIBI_STANCE_ENUM_BATTLE
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	assert.Equal(spine.CHIBI_STANCE_ENUM_BATTLE, current.ChibiStance)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiSetFaceOnlyForBattleStance(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	_, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi face back",
	})

	if assert.Error(err) {
		assert.ErrorContains(err, "base chibi's can't face backwards.")
	}
}

func TestCmdProcessorHandleMessage_ChibiSetFaceOnly(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	current.ChibiStance = spine.CHIBI_STANCE_ENUM_BATTLE
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	assert.Equal(spine.CHIBI_FACING_ENUM_BACK, current.Facing)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiEnemyNotEnoughArgs(t *testing.T) {
	// TODO: Need to add enemy to AssetService in order to test enemy happy path
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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

func TestCmdProcessorHandleMessage_ChibiWalk(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	assert.Equal(spine.ACTION_WANDER, current.CurrentAction)
	assert.Equal([]string{"Move"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiWalkTo(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	assert.Equal(spine.ACTION_WALK_TO, current.CurrentAction)
	assert.Equal([]string{"Move", "Relax"}, current.Action.GetAnimations(current.CurrentAction))
	assert.Equal(0.5, current.Action.TargetPos.Unwrap().X)
	assert.Equal(1.0, current.AnimationSpeed)
}

func TestCmdProcessorHandleMessage_ChibiAnimationSpeed(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	_, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi speed 6.0",
	})

	if assert.Error(err) {
		assert.ErrorContains(err, "try something like !chibi speed 2.0")
	}
}

func TestCmdProcessorHandleMessage_ChibiChooseChibi(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
		Message:         "!chibi amiya",
	})

	assert.Nil(err)
	assert.Empty(cmd.Reply(actor))
	// cmd.UpdateActor(actor)
	if updateActorCmd, ok := cmd.(*ChatCommandUpdateActor); ok {
		assert.Equal(updateActorCmd.username, "user1")
		assert.Equal(updateActorCmd.usernameDisplay, "user1DisplayName")
	} else {
		assert.Fail("Command is not of type: ChatCommandUpdateActor")
	}
	assert.Equal("char_002_amiya", current.OperatorId)
	assert.Equal(spine.FACTION_ENUM_OPERATOR, current.Faction)
	assert.Equal(1.0, current.AnimationSpeed)
	assert.True(current.SpriteScale.IsNone())
}

func TestCmdProcessorHandleMessage_ChibiScale(t *testing.T) {
	actor := NewFakeChibiActor()
	current, sut := setupCommandTest()
	actor.UpdateChatter("user1", "user1DisplayName", current)

	assert := assert.New(t)
	cmd, err := sut.HandleMessage(current, chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "user1DisplayName",
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
