package chibi

import (
	"context"
	"testing"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	spine "github.com/Stymphalian/ak_chibi_bot/server/internal/spine_runtime"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
	"github.com/stretchr/testify/assert"
)

func setupActorTest() *ChibiActor {
	usersRepo := users.NewUserRepositoryPsql()
	chattersRepo := users.NewChatterRepositoryPsql()
	assetManager := operator.NewTestAssetService()
	spineService := operator.NewOperatorService(assetManager, misc.DefaultSpineRuntimeConfig())
	fakeSpineClient := spine.NewFakeSpineClient()
	sut := NewChibiActor(
		5000,
		spineService,
		usersRepo,
		chattersRepo,
		fakeSpineClient,
		[]string{"exlude_user"},
	)
	return sut
}

// TODO: more tests for actor

func TestGiveChibiToUserToExclude(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	userinfo := misc.UserInfo{
		Username:        "exlude_user",
		UsernameDisplay: "userDisplay",
		TwitchUserId:    "100",
	}
	err := sut.GiveChibiToUser(ctx, userinfo)
	assert.Nil(err)
}

func TestGiveChibiToUserNormalFlow(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	userinfo := misc.UserInfo{
		Username:        "user",
		UsernameDisplay: "userDisplay",
		TwitchUserId:    "100",
	}
	err := sut.GiveChibiToUser(ctx, userinfo)

	assert.Nil(err)
	assert.Equal(len(sut.ChatUsers), 1)
}

func TestRemoveUserNormal(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	userinfo := misc.UserInfo{
		Username:        "user1",
		UsernameDisplay: "userDisplay1",
		TwitchUserId:    "100",
	}
	userinfo2 := misc.UserInfo{
		Username:        "user2",
		UsernameDisplay: "userDisplay2",
		TwitchUserId:    "100",
	}
	sut.GiveChibiToUser(ctx, userinfo)
	sut.GiveChibiToUser(ctx, userinfo2)
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
	userinfo := misc.UserInfo{
		Username:        "user1",
		UsernameDisplay: "userDisplay1",
		TwitchUserId:    "100",
	}
	userinfo2 := misc.UserInfo{
		Username:        "user2",
		UsernameDisplay: "userDisplay2",
		TwitchUserId:    "100",
	}
	sut.GiveChibiToUser(ctx, userinfo)
	sut.GiveChibiToUser(ctx, userinfo2)
	assert.Equal(len(sut.ChatUsers), 2)

	assert.True(sut.HasChibi(ctx, "user1"))
	assert.True(sut.HasChibi(ctx, "user2"))
	assert.False(sut.HasChibi(ctx, "user3"))
}

func TestSetToDefault(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	userinfo := misc.UserInfo{
		Username:        "user1",
		UsernameDisplay: "userDisplay1",
		TwitchUserId:    "100",
	}

	sut.SetToDefault(ctx, userinfo, "Amiya", misc.InitialOperatorDetails{
		Skin:       operator.DEFAULT_SKIN_NAME,
		Stance:     string(operator.CHIBI_STANCE_ENUM_BASE),
		Animations: []string{operator.DEFAULT_ANIM_BASE},
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
		TwitchUserId:    "100",
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

	opInfo := operator.NewOperatorInfo(
		"Amiya",
		operator.FACTION_ENUM_OPERATOR,
		"char_002_amiya",
		"default",
		operator.CHIBI_STANCE_ENUM_BASE,
		operator.CHIBI_FACING_ENUM_FRONT,
		[]string{"default"},
		[]string{"Idle", "Relax"},
		1.0,
		misc.EmptyOption[misc.Vector2](),
		operator.ACTION_PLAY_ANIMATION,
		operator.NewActionPlayAnimation([]string{"default"}),
	)
	userinfo := misc.UserInfo{
		Username:        "user1",
		UsernameDisplay: "userDisplay1",
		TwitchUserId:    "100",
	}
	sut.UpdateChibi(ctx, userinfo, &opInfo)
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
	userinfo := misc.UserInfo{
		Username:        "user1",
		UsernameDisplay: "userDisplay1",
		TwitchUserId:    "100",
	}
	sut.GiveChibiToUser(ctx, userinfo)
	opInfo, err := sut.CurrentInfo(ctx, "user1")
	assert.Nil(err)
	assert.Equal(opInfo.OperatorId, "char_002_amiya")
}

func TestUpdateChatter(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	ctx := context.TODO()
	opInfo := operator.EmptyOperatorInfo()
	userinfo := misc.UserInfo{
		Username:        "user1",
		UsernameDisplay: "userDisplay1",
		TwitchUserId:    "100",
	}
	err := sut.UpdateChatter(ctx, userinfo, opInfo)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(len(sut.ChatUsers), 1)
	period := time.Duration(1) * time.Second
	assert.True(misc.Clock.Since(sut.ChatUsers["user1"].GetLastChatTime()) < period)
}
