package chatbot

import (
	"testing"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chibi"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/stretchr/testify/assert"
)

func setupTest() (*TwitchBot, *chibi.FakeChibiActor) {
	fakeChibiActor := chibi.NewFakeChibiActor()
	twitchBot, _ := NewTwitchBot(
		fakeChibiActor,
		"stymphalian__",
		"stymtwitchbot",
		"access_token",
		1,
	)
	return twitchBot, fakeChibiActor
}

// Very weird dependency on ChibiActors

func TestHandlePrivateMessageNormalMessage(t *testing.T) {
	assert := assert.New(t)
	twitchBot, fakeChibiActor := setupTest()

	testMsg := twitch.PrivateMessage{
		User:    twitch.User{Name: "user"},
		Message: "hello world",
	}
	twitchBot.HandlePrivateMessage(testMsg)

	assert.Equal("Chibi", fakeChibiActor.Users["user"])
}

func TestHandlePrivateMessageChibiCommand(t *testing.T) {
	assert := assert.New(t)
	twitchBot, fakeChibiActor := setupTest()

	testMsg := twitch.PrivateMessage{
		User:    twitch.User{Name: "user"},
		Message: "!chibi help",
	}
	twitchBot.HandlePrivateMessage(testMsg)

	assert.Equal("!chibi help", fakeChibiActor.Users["user"])
}

func TestReadPumpShouldShouldFailToConnectDueToInvalidAccessToken(t *testing.T) {
	twitchBot, _ := setupTest()
	err := twitchBot.ReadPump()
	if !assert.Error(t, err) {
		assert.Fail(t, "Read pump should have failed to connect")
	}
}
