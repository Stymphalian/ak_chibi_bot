package chatbot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchBot struct {
	chatMessageHandler chat.ChatMessageHandler
	channelName        string
	tc                 *twitch.Client
}

func NewTwitchBot(
	chatMessageHandler chat.ChatMessageHandler,
	twitchChannelName string,
	twitchBotName string,
	twitchAccessToken string) (*TwitchBot, error) {

	accessToken := twitchAccessToken
	if len(accessToken) == 0 {
		accessToken = os.Getenv("TWITCH_ACCESS_TOKEN")
	}
	if len(accessToken) == 0 {
		return nil, fmt.Errorf("no twitch access token is set")
	}
	tc := twitch.NewClient(
		twitchBotName,
		"oauth:"+accessToken,
	)
	self := &TwitchBot{
		chatMessageHandler: chatMessageHandler,
		channelName:        twitchChannelName,
		tc:                 tc,
	}
	return self, nil
}

func (t *TwitchBot) Close() error {
	log.Println("TwitchBot::Close() called")
	err := t.tc.Disconnect()
	if err != nil {
		log.Println(err)
	}
	log.Println("TwitchBot::Close() finished")
	return err
}

func (t *TwitchBot) HandlePrivateMessage(m twitch.PrivateMessage) {
	var trimmed = strings.TrimSpace(m.Message)
	if len(trimmed) == 0 {
		return
	}
	if trimmed[0] == '!' {
		log.Printf("PRIVMSG message %v\n", m)
	}

	chatMessage := chat.ChatMessage{
		Username:        m.User.Name,
		UserDisplayName: m.User.DisplayName,
		TwitchUserId:    m.User.ID,
		Message:         trimmed,
	}
	outputMsg, err := t.chatMessageHandler.HandleMessage(chatMessage)
	if err == nil && len(outputMsg) > 0 {
		t.tc.Say(m.Channel, outputMsg)
	}
}

func (t *TwitchBot) ReadLoop() error {
	t.tc.OnNoticeMessage(func(m twitch.NoticeMessage) {
		log.Printf("NOTICE message %v\n", m)
	})
	t.tc.OnUserJoinMessage(func(m twitch.UserJoinMessage) {
		log.Printf("JOIN message %v\n", m)
	})
	t.tc.OnUserPartMessage(func(m twitch.UserPartMessage) {
		log.Printf("PART message %v\n", m)
		// t.chatMessageHandler.RemoveUserChibi(m.User)
	})
	t.tc.OnPrivateMessage(func(m twitch.PrivateMessage) {
		// log.Printf("PRIVMSG message %v\n", m)
		t.HandlePrivateMessage(m)
	})
	t.tc.Join(t.channelName)

	log.Println("Joined channel name", t.channelName)
	if err := t.tc.Connect(); err != nil {
		if !errors.Is(err, twitch.ErrClientDisconnected) {
			log.Println("Failed to connect to twitch", err)
			return err
		}
	}
	log.Println("Read pump done for ", t.channelName)
	return nil
}
