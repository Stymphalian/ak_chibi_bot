package room

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/internal/twitchbot"
)

const (
	forcefulShutdownDuration = time.Duration(30) * time.Second
)

type RoomsManager struct {
	Rooms       map[string]*Room
	rooms_mutex sync.Mutex

	AssetManager *spine.AssetManager
	TwitchConfig *misc.TwitchConfig

	runCh          chan string
	shutdownDoneCh chan struct{}
}

func NewRoomsManager(assets *spine.AssetManager, twitchConfig *misc.TwitchConfig) *RoomsManager {
	return &RoomsManager{
		Rooms:          make(map[string]*Room, 0),
		AssetManager:   assets,
		TwitchConfig:   twitchConfig,
		runCh:          make(chan string),
		shutdownDoneCh: make(chan struct{}),
	}
}

func (r *RoomsManager) GarbageCollectRooms(timer *time.Ticker, period time.Duration) {
	for range timer.C {
		log.Println("Garbage collecting unused chat rooms")

		r.rooms_mutex.Lock()
		for channel, room := range r.Rooms {
			lastChat := room.TwitchChat.LastChatterTime()
			if time.Since(lastChat) > period {
				room.Close()
				delete(r.Rooms, channel)
			}
		}
		r.rooms_mutex.Unlock()
	}
}

func (r *RoomsManager) RunLoop() {
	if r.TwitchConfig.RemoveUnusedRoomsAfterMinutes > 0 {
		cleanupInterval := time.Duration(r.TwitchConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
		timer := time.NewTicker(cleanupInterval)
		defer timer.Stop()
		go r.GarbageCollectRooms(timer, cleanupInterval)
	}

	for channel := range r.runCh {
		go r.Rooms[channel].Run()
	}
}

func (r *RoomsManager) CreateRoomOrNoOp(channel string, ctx context.Context) error {
	if _, ok := r.Rooms[channel]; ok {
		return nil
	}

	log.Println("CreateRoomOrNoOp before")
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
	log.Println("CreateRoomOrNoOp after")

	r.rooms_mutex.Lock()
	defer r.rooms_mutex.Unlock()
	r.Rooms[channel] = NewRoom(
		channel,
		r.TwitchConfig.InitialOperator,
		r.TwitchConfig.OperatorDetails,
		spineBridge,
		chibiActor,
		twitchBot,
	)
	r.runCh <- channel
	return nil
}

func (m *RoomsManager) HandleSpineWebSocket(channelName string, w http.ResponseWriter, r *http.Request) error {
	room, ok := m.Rooms[channelName]
	if !ok {
		return errors.New("channel room does not exist")
	}
	return room.SpineBridge.AddWebsocketConnection(w, r)
}

func (r *RoomsManager) Shutdown() {
	log.Println("RoomsManager calling Shutdown")
	close(r.runCh)

	go func() {
		defer close(r.shutdownDoneCh)
		r.rooms_mutex.Lock()
		defer r.rooms_mutex.Unlock()
		for _, room := range r.Rooms {
			err := room.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

func (r *RoomsManager) WaitForShutdownWithTimeout() {
	log.Println("Waiting for shutdown")
	select {
	case <-time.After(forcefulShutdownDuration):
		log.Printf("Shutting down forcefully after %v", forcefulShutdownDuration)
	case <-r.shutdownDoneCh:
		log.Println("Shutting down gracefully")
	}
}
