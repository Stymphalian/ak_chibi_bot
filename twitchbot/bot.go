package twitchbot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/misc"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchBot struct {
	chibiClient  ChibiClient
	twitchConfig *misc.TwitchConfig

	tc           *twitch.Client
	lastUserChat map[string]time.Time
}

func NewTwitchBot(chibiClient ChibiClient, config *misc.TwitchConfig) (*TwitchBot, error) {

	accessToken := config.TwitchAccessToken
	if len(accessToken) == 0 {
		accessToken = os.Getenv("TWITCH_ACCESS_TOKEN")
	}
	if len(accessToken) == 0 {
		return nil, fmt.Errorf("no twitch access token is set")
	}
	tc := twitch.NewClient(
		config.TwitchBot,
		"oauth:"+accessToken,
	)
	self := &TwitchBot{
		tc:           tc,
		chibiClient:  chibiClient,
		twitchConfig: config,
		lastUserChat: make(map[string]time.Time),
	}
	return self, nil
}

func (t *TwitchBot) Close() {
	log.Println("AKChibiBot::Close() called")
	t.tc.Disconnect()
	log.Println("AKChibiBot::Close() finished")
}

func (t *TwitchBot) HandlePrivateMessage(m twitch.PrivateMessage) {
	var trimmed = strings.TrimSpace(m.Message)
	if len(trimmed) == 0 {
		return
	}
	t.lastUserChat[m.User.Name] = m.Time
	if !t.chibiClient.HasChibi(m.User.Name) {
		t.chibiClient.GiveChibiToUser(m.User.Name, m.User.DisplayName)
	}
	if trimmed[0] != '!' {
		return
	}

	outputMsg, err := t.chibiClient.HandleCommand(m.User.Name, m.User.DisplayName, trimmed)
	if len(outputMsg) > 0 {
		t.tc.Say(t.twitchConfig.ChannelName, outputMsg)
	}
	if err != nil {
		t.tc.Say(t.twitchConfig.ChannelName, err.Error())
	}
}

func (t *TwitchBot) garbageCollectOldChibis(timer *time.Ticker, period time.Duration) {
	for range timer.C {
		log.Println("Garbage collecting old chibis")
		for user, lastChat := range t.lastUserChat {
			if time.Since(lastChat) > period {
				if user == t.twitchConfig.Broadcaster {
					continue
				}
				log.Println("Removing chibi for", user)
				t.chibiClient.RemoveUserChibi(user)
			}
		}
	}
}

func (t *TwitchBot) ReadPump() {
	// TODO: Make this timer for timing out chibis configurable
	// TODO: Move this logic into chibi
	if t.twitchConfig.RemoveChibiAfterMinutes > 0 {
		cleanupInterval := time.Duration(t.twitchConfig.RemoveChibiAfterMinutes) * time.Minute
		timer := time.NewTicker(cleanupInterval)
		defer timer.Stop()
		go t.garbageCollectOldChibis(timer, 20*time.Minute)
	}

	t.tc.OnNoticeMessage(func(m twitch.NoticeMessage) {
		log.Printf("NOTICE message %v\n", m)
	})
	t.tc.OnUserJoinMessage(func(m twitch.UserJoinMessage) {
		log.Printf("JOIN message %v\n", m)
	})
	t.tc.OnUserPartMessage(func(m twitch.UserPartMessage) {
		log.Printf("PART message %v\n", m)
		t.chibiClient.RemoveUserChibi(m.User)
	})
	t.tc.OnPrivateMessage(func(m twitch.PrivateMessage) {
		log.Printf("PRIVMSG message %v\n", m)
		t.HandlePrivateMessage(m)
	})
	t.tc.Join(t.twitchConfig.ChannelName)

	log.Println("Joined channel name", t.twitchConfig.ChannelName)
	if err := t.tc.Connect(); err != nil {
		log.Println("Failed to connect to twitch", err)
	}
	log.Println("Read pump done")
}
