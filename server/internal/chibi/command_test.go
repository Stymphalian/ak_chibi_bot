package chibi

import (
	"testing"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
	"github.com/stretchr/testify/assert"
)

func setupCommandTest() (*FakeChibiActor, ChatCommandProcessor) {
	fakeChibiActor := NewFakeChibiActor()
	assetManager := spine.NewTestAssetService()
	spineService := spine.NewSpineService(assetManager)
	fakeSpineClient := spine.NewFakeSpineClient()
	sut := ChatCommandProcessor{
		chibiActor:   fakeChibiActor,
		spineService: spineService,
		client:       fakeSpineClient,
	}
	return fakeChibiActor, sut
}

func TestHandleMesssageNotChibiCommand(t *testing.T) {
	_, sut := setupCommandTest()
	sut.chibiActor.GiveChibiToUser("user", "userDisplay")
	assert := assert.New(t)
	gotMsg, gotErr := sut.HandleMessage("user", "userDisplay", "hello world")

	assert.Nil(gotErr)
	assert.Equal("", gotMsg)
}

func TestChibiHelp(t *testing.T) {
	_, sut := setupCommandTest()
	sut.chibiActor.GiveChibiToUser("user", "userDisplay")

	assert := assert.New(t)
	gotMsg, gotErr := sut.HandleMessage("user", "userDisplay", "!chibi help")

	assert.Nil(gotErr)
	assert.Contains(gotMsg, "!chibi to control your Arknights chibi.")
}

func TestChibiSkins(t *testing.T) {
	_, sut := setupCommandTest()
	sut.chibiActor.GiveChibiToUser("user", "userDisplay")

	assert := assert.New(t)
	gotMsg, gotErr := sut.HandleMessage("user", "userDisplay", "!chibi skins")

	assert.Nil(gotErr)
	assert.Contains(gotMsg, "user skins")
}

func TestChibiFuzzyChooseOperator(t *testing.T) {
	fakeChibiActor, sut := setupCommandTest()
	sut.chibiActor.GiveChibiToUser("user", "userDisplay")

	assert := assert.New(t)
	gotMsg, gotErr := sut.HandleMessage("user", "userDisplay", "!chibi amiya")

	assert.Nil(gotErr)
	assert.Contains(gotMsg, "")
	assert.Equal(fakeChibiActor.Users["user"].OperatorId, "char_002_amiya")
}
