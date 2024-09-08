package twitchbot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/gempir/go-twitch-irc/v4"
)

type ChannelUser struct {
	User string
}

type TwitchBot struct {
	chibiActor                  chibi.ChibiActorInterface
	channelName                 string
	garbageCollectionPeriodMins int
	tc                          *twitch.Client
	lastUserChat                map[ChannelUser]time.Time
	latestChatterTime           time.Time
}

func NewTwitchBot(
	chibiActor chibi.ChibiActorInterface,
	twitchChannelName string,
	twitchBotName string,
	twitchAccessToken string,
	garbageCollectionPeriodMins int) (*TwitchBot, error) {

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
		chibiActor:                  chibiActor,
		channelName:                 twitchChannelName,
		garbageCollectionPeriodMins: garbageCollectionPeriodMins,
		tc:                          tc,
		lastUserChat:                make(map[ChannelUser]time.Time),
		latestChatterTime:           time.Now(),
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

func (t *TwitchBot) garbageCollectOldChibis() {
	log.Println("Garbage collecting old chibis")
	interval := time.Duration(t.garbageCollectionPeriodMins) * time.Minute
	for channelUser, lastChat := range t.lastUserChat {
		user := channelUser.User
		if time.Since(lastChat) > interval {
			if user == t.channelName {
				// Skip removing the broadcaster's chibi
				continue
			}
			log.Println("Removing chibi for", user)
			t.chibiActor.RemoveUserChibi(user)
		}
	}
}

func (t *TwitchBot) ReadPump() {
	if t.garbageCollectionPeriodMins > 0 {
		stopTimer := misc.StartTimer(
			fmt.Sprintf("GarbageCollectOldChibis %s", t.channelName),
			time.Duration(t.garbageCollectionPeriodMins)*time.Minute,
			t.garbageCollectOldChibis,
		)
		defer stopTimer()
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
		if !errors.Is(err, twitch.ErrClientDisconnected) {
			log.Println("Failed to connect to twitch", err)
		}
	}
	log.Println("Read pump done for ", t.channelName)
}

func (t *TwitchBot) LastChatterTime() time.Time {
	return t.latestChatterTime
}

func (t *TwitchBot) LastChatTime(username string) (time.Time, bool) {
	if lastTime, ok := t.lastUserChat[ChannelUser{
		User: username,
	}]; ok {
		return lastTime, true
	}
	return time.Now(), false
}
