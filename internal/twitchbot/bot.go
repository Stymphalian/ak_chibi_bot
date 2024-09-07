package twitchbot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/chibi"
	"github.com/gempir/go-twitch-irc/v4"
)

type ChannelUser struct {
	User string
}

type TwitchBot struct {
	chibiActor                chibi.ChibiActorInterface
	channelName               string
	garbageCollectionRateMins int
	tc                        *twitch.Client
	lastUserChat              map[ChannelUser]time.Time
	latestChatterTime         time.Time
}

func NewTwitchBot(
	chibiActor chibi.ChibiActorInterface,
	twitchChannelName string,
	twitchBotName string,
	twitchAccessToken string,
	garbageCollectionRateMins int) (*TwitchBot, error) {

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
		chibiActor:                chibiActor,
		channelName:               twitchChannelName,
		garbageCollectionRateMins: garbageCollectionRateMins,
		tc:                        tc,
		lastUserChat:              make(map[ChannelUser]time.Time),
		latestChatterTime:         time.Now(),
	}
	return self, nil
}

func (t *TwitchBot) Close() {
	log.Println("TwitchBot::Close() called")
	t.chibiActor.Close()
	t.tc.Disconnect()
	log.Println("TwitchBot::Close() finished")
}

func (t *TwitchBot) HandlePrivateMessage(m twitch.PrivateMessage) {
	var trimmed = strings.TrimSpace(m.Message)
	if len(trimmed) == 0 {
		return
	}
	t.lastUserChat[ChannelUser{m.User.Name}] = m.Time
	t.latestChatterTime = m.Time
	if !t.chibiActor.HasChibi(m.User.Name) {
		t.chibiActor.GiveChibiToUser(m.User.Name, m.User.DisplayName)
	}
	if trimmed[0] != '!' {
		return
	}

	outputMsg, err := t.chibiActor.HandleCommand(m.User.Name, m.User.DisplayName, trimmed)
	if len(outputMsg) > 0 {
		t.tc.Say(m.Channel, outputMsg)
	}
	if err != nil {
		t.tc.Say(m.Channel, err.Error())
	}
}

func (t *TwitchBot) garbageCollectOldChibis(timer *time.Ticker, period time.Duration) {
	for range timer.C {
		log.Println("Garbage collecting old chibis")
		for channelUser, lastChat := range t.lastUserChat {
			user := channelUser.User
			if time.Since(lastChat) > period {
				if user == t.channelName {
					continue
				}
				log.Println("Removing chibi for", user)
				t.chibiActor.RemoveUserChibi(user)
			}
		}
	}
}

func (t *TwitchBot) ReadPump() {
	if t.garbageCollectionRateMins > 0 {
		cleanupInterval := time.Duration(t.garbageCollectionRateMins) * time.Minute
		timer := time.NewTicker(cleanupInterval)
		defer timer.Stop()
		go t.garbageCollectOldChibis(timer, cleanupInterval)
	}

	t.tc.OnNoticeMessage(func(m twitch.NoticeMessage) {
		log.Printf("NOTICE message %v\n", m)
	})
	t.tc.OnUserJoinMessage(func(m twitch.UserJoinMessage) {
		log.Printf("JOIN message %v\n", m)
	})
	t.tc.OnUserPartMessage(func(m twitch.UserPartMessage) {
		log.Printf("PART message %v\n", m)
		t.chibiActor.RemoveUserChibi(m.User)
	})
	t.tc.OnPrivateMessage(func(m twitch.PrivateMessage) {
		log.Printf("PRIVMSG message %v\n", m)
		t.HandlePrivateMessage(m)
	})
	t.tc.Join(t.channelName)

	log.Println("Joined channel name", t.channelName)
	if err := t.tc.Connect(); err != nil {
		log.Println("Failed to connect to twitch", err)
	}
	log.Println("Read pump done")
}

func (t *TwitchBot) LastChatterTime() time.Time {
	return t.latestChatterTime
}
