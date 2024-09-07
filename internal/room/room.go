package room

import (
	"context"
	"log"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/internal/twitchbot"
)

type Room struct {
	ChannelName string
	SpineBridge *spine.SpineBridge
	ChibiActor  *chibi.ChibiActor
	TwitchChat  *twitchbot.TwitchBot
}

func (r *Room) Close() {
	log.Println("Closing room: ", r.ChannelName)
	r.TwitchChat.Close()
}

func (r *Room) Run() {
	defer r.TwitchChat.Close()
	go r.TwitchChat.ReadPump()
}

type RoomsManager struct {
	Rooms        map[string]*Room
	AssetManager *spine.AssetManager
	TwitchConfig *misc.TwitchConfig
	RunCh        chan string
}

func NewRoomsManager(assets *spine.AssetManager, twitchConfig *misc.TwitchConfig) *RoomsManager {
	return &RoomsManager{
		Rooms:        make(map[string]*Room, 0),
		AssetManager: assets,
		TwitchConfig: twitchConfig,
		RunCh:        make(chan string),
	}
}

func (r *RoomsManager) GarbageCollectRooms(timer *time.Ticker, period time.Duration) {
	for range timer.C {
		log.Println("Garbage collecting unused chat rooms")
		for channel, room := range r.Rooms {
			lastChat := room.TwitchChat.LastChatterTime()
			if time.Since(lastChat) > period {
				room.Close()
				delete(r.Rooms, channel)
			}
		}
	}
}

func (r *RoomsManager) RunLoop() {
	if r.TwitchConfig.RemoveUnusedRoomsAfterMinutes > 0 {
		cleanupInterval := time.Duration(r.TwitchConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
		timer := time.NewTicker(cleanupInterval)
		defer timer.Stop()
		go r.GarbageCollectRooms(timer, cleanupInterval)
	}

	for {
		select {
		case channel := <-r.RunCh:
			go r.Rooms[channel].Run()
		}
	}
}

func (r *RoomsManager) CreateRoomOrNoOp(channel string, ctx context.Context) error {
	spineBridge, err := spine.NewSpineBridge(r.AssetManager)
	if err != nil {
		return err
	}
	chibiActor := chibi.NewChibiActor(spineBridge, r.TwitchConfig.ExcludeNames)
	twitchBot, err := twitchbot.NewTwitchBot(
		chibiActor,
		channel,
		r.TwitchConfig.TwitchBot,
		r.TwitchConfig.TwitchAccessToken,
		r.TwitchConfig.RemoveChibiAfterMinutes,
	)
	if err != nil {
		return err
	}

	if _, ok := r.Rooms[channel]; !ok {
		r.Rooms[channel] = &Room{
			ChannelName: channel,
			SpineBridge: spineBridge,
			ChibiActor:  chibiActor,
			TwitchChat:  twitchBot,
		}
		r.RunCh <- channel
	}
	return nil
}

func (r *RoomsManager) Shutdown() {
	for _, room := range r.Rooms {
		room.Close()
	}
}
