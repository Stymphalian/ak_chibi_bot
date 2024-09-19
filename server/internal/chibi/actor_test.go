package chibi

import (
	"context"
	"testing"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
	"github.com/stretchr/testify/assert"
)

func setupActorTest() *ChibiActor {
	assetManager := spine.NewTestAssetService()
	spineService := spine.NewSpineService(assetManager, misc.DefaultSpineRuntimeConfig())
	fakeSpineClient := spine.NewFakeSpineClient()
	sut := NewChibiActor(
		spineService,
		fakeSpineClient,
		[]string{"exlude_user"},
	)
	sut.SetRoomId(5000)
	return sut
}

// TODO: more tests for actor

func TestGiveChibiToUserToExclude(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	err := sut.GiveChibiToUser(ctx, "exlude_user", "userDisplay")
	assert.Nil(err)
}

func TestGiveChibiToUserNormalFlow(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	err := sut.GiveChibiToUser(ctx, "user", "userDisplay")

	assert.Nil(err)
	assert.Equal(len(sut.ChatUsers), 1)
}

func TestRemoveUserNormal(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	sut.GiveChibiToUser(ctx, "user1", "userDisplay1")
	sut.GiveChibiToUser(ctx, "user2", "userDisplay2")
	assert.Equal(len(sut.ChatUsers), 2)

	err := sut.RemoveUserChibi(ctx, "user1")

	assert.Nil(err)
	assert.Equal(len(sut.ChatUsers), 1)
	assert.Contains(sut.ChatUsers, "user2")
}

func TestHasChibi(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	sut.GiveChibiToUser(ctx, "user1", "userDisplay1")
	sut.GiveChibiToUser(ctx, "user2", "userDisplay2")
	assert.Equal(len(sut.ChatUsers), 2)

	assert.True(sut.HasChibi(ctx, "user1"))
	assert.True(sut.HasChibi(ctx, "user2"))
	assert.False(sut.HasChibi(ctx, "user3"))
}

func TestSetToDefault(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()

	sut.SetToDefault(ctx, "user1", "Amiya", misc.InitialOperatorDetails{
		Skin:       spine.DEFAULT_SKIN_NAME,
		Stance:     string(spine.CHIBI_STANCE_ENUM_BASE),
		Animations: []string{spine.DEFAULT_ANIM_BASE},
		PositionX:  0.5,
	})
	assert.Contains(sut.ChatUsers, "user1")
	assert.Contains(sut.ChatUsers["user1"].GetOperatorInfo().Skin, "default")
}

func TestChibiActorHandleMessage(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	sut.HandleMessage(chat.ChatMessage{
		Username:        "user1",
		UserDisplayName: "userDisplay",
		Message:         "!chibi Amiya",
	})

	assert.Contains(sut.ChatUsers, "user1")
	assert.Contains(sut.ChatUsers["user1"].GetOperatorInfo().Skin, "default")
	assert.True(misc.Clock.Since(sut.ChatUsers["user1"].GetLastChatTime()) < time.Duration(1)*time.Second)
}

func TestChibiActorUpdateChibi(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()

	opInfo := spine.NewOperatorInfo(
		"Amiya",
		spine.FACTION_ENUM_OPERATOR,
		"char_002_amiya",
		"default",
		spine.CHIBI_STANCE_ENUM_BASE,
		spine.CHIBI_FACING_ENUM_FRONT,
		[]string{"default"},
		[]string{"Idle", "Relax"},
		1.0,
		misc.EmptyOption[misc.Vector2](),
		spine.ACTION_PLAY_ANIMATION,
		spine.NewActionPlayAnimation([]string{"default"}),
	)
	sut.UpdateChibi(ctx, "user1", "userDisplay", &opInfo)
	assert.Contains(sut.ChatUsers, "user1")
	assert.Contains(sut.ChatUsers["user1"].GetOperatorInfo().Skin, "default")
	assert.Equal(sut.ChatUsers["user1"].GetOperatorInfo().AvailableAnimations, []string{"Move", "base_front1", "base_front2"})
}

func TestActorCurrentInfoEmpty(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	_, err := sut.CurrentInfo(ctx, "user1")
	assert.Error(err)
}

func TestActorCurrentInfoExists(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	sut.GiveChibiToUser(ctx, "user1", "userDisplay1")
	opInfo, err := sut.CurrentInfo(ctx, "user1")
	assert.Nil(err)
	assert.Equal(opInfo.OperatorId, "char_002_amiya")
}

func TestUpdateChatter(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	opInfo := spine.EmptyOperatorInfo()
	err := sut.UpdateChatter(ctx, "user1", "userDisplay1", opInfo)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(len(sut.ChatUsers), 1)
	period := time.Duration(1) * time.Second
	assert.True(misc.Clock.Since(sut.ChatUsers["user1"].GetLastChatTime()) < period)
}
