package room

import (
	"log"

	"github.com/Stymphalian/ak_chibi_bot/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/internal/twitchbot"
)

type Room struct {
	ChannelName string

	DefaultOperatorName   string
	DefaultOperatorConfig misc.InitialOperatorDetails
	SpineBridge           *spine.SpineBridge
	ChibiActor            *chibi.ChibiActor
	TwitchChat            *twitchbot.TwitchBot
}

func NewRoom(
	channel string,
	opName string,
	opConfig misc.InitialOperatorDetails,
	spineBridge *spine.SpineBridge,
	chibiActor *chibi.ChibiActor,
	twitchBot *twitchbot.TwitchBot) *Room {
	return &Room{
		ChannelName:           channel,
		DefaultOperatorName:   opName,
		DefaultOperatorConfig: opConfig,
		SpineBridge:           spineBridge,
		ChibiActor:            chibiActor,
		TwitchChat:            twitchBot,
	}
}

func (r *Room) Close() error {
	log.Println("Closing room ", r.ChannelName)
	// Disconnect the twitch chat
	err := r.TwitchChat.Close()
	if err != nil {
		return err
	}

	// Disconnect all websockets
	err = r.SpineBridge.Close()
	if err != nil {
		return err
	}
	log.Println("Closed room ", r.ChannelName, "successfully")
	return nil
}

func (r *Room) Run() {
	log.Printf("Room %s is running\n", r.ChannelName)
	r.ChibiActor.SetToDefault(r.ChannelName, r.DefaultOperatorName, r.DefaultOperatorConfig)
	r.TwitchChat.ReadPump()
	log.Printf("Room %s run is done\n", r.ChannelName)
}
