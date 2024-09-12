package chibi

import (
	"testing"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
	"github.com/stretchr/testify/assert"
)

func setupActorTest() *ChibiActor {
	assetManager := spine.NewTestAssetService()
	spineService := spine.NewSpineService(assetManager)
	fakeSpineClient := spine.NewFakeSpineClient()
	sut := NewChibiActor(
		spineService,
		fakeSpineClient,
		[]string{"exlude_user"},
	)
	return sut
}

// TODO: more tests for actor

func TestGiveChibiToUserToExclude(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	err := sut.GiveChibiToUser("exlude_user", "userDisplay")
	assert.Nil(err)
}

func TestGiveChibiToUserNormalFlow(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	err := sut.GiveChibiToUser("user", "userDisplay")

	assert.Nil(err)
	assert.Equal(len(sut.ChatUsers), 1)
}

func TestRemoveUserNormal(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	sut.GiveChibiToUser("user1", "userDisplay1")
	sut.GiveChibiToUser("user2", "userDisplay2")
	assert.Equal(len(sut.ChatUsers), 2)

	err := sut.RemoveUserChibi("user1")

	assert.Nil(err)
	assert.Equal(len(sut.ChatUsers), 1)
	assert.Contains(sut.ChatUsers, "user2")
}

func TestHasChibi(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	sut.GiveChibiToUser("user1", "userDisplay1")
	sut.GiveChibiToUser("user2", "userDisplay2")
	assert.Equal(len(sut.ChatUsers), 2)

	assert.True(sut.HasChibi("user1"))
	assert.True(sut.HasChibi("user2"))
	assert.False(sut.HasChibi("user3"))
}

func TestSetToDefault(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()

	sut.SetToDefault("user1", "Amiya", misc.InitialOperatorDetails{
		Skin:       spine.DEFAULT_SKIN_NAME,
		Stance:     string(spine.CHIBI_STANCE_ENUM_BASE),
		Animations: []string{spine.DEFAULT_ANIM_BASE},
		PositionX:  0.5,
	})
	assert.Contains(sut.ChatUsers, "user1")
	assert.Contains(sut.ChatUsers["user1"].CurrentOperator.Skin, "default")
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
	assert.Contains(sut.ChatUsers["user1"].CurrentOperator.Skin, "default")
	assert.True(misc.Clock.Since(sut.ChatUsers["user1"].LastChatTime) < time.Duration(1)*time.Second)
}

func TestChibiActorUpdateChibi(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()

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
	sut.UpdateChibi("user1", "userDisplay", &opInfo)
	assert.Contains(sut.ChatUsers, "user1")
	assert.Contains(sut.ChatUsers["user1"].CurrentOperator.Skin, "default")
	assert.Equal(sut.ChatUsers["user1"].CurrentOperator.AvailableAnimations, []string{"Move", "base_front1", "base_front2"})
}

func TestActorCurrentInfoEmpty(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	_, err := sut.CurrentInfo("user1")
	assert.Error(err)
}

func TestActorCurrentInfoExists(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	sut.GiveChibiToUser("user1", "userDisplay1")
	opInfo, err := sut.CurrentInfo("user1")
	assert.Nil(err)
	assert.Equal(opInfo.OperatorId, "char_002_amiya")
}

func TestUpdateChatter(t *testing.T) {
	assert := assert.New(t)
	sut := setupActorTest()
	opInfo := spine.EmptyOperatorInfo()
	sut.UpdateChatter("user1", "userDisplay1", opInfo)

	assert.Equal(len(sut.ChatUsers), 1)
	period := time.Duration(1) * time.Second
	assert.True(misc.Clock.Since(sut.ChatUsers["user1"].LastChatTime) < period)
}
