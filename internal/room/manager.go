package room

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/internal/twitch_api"
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
	twitchClient *twitch_api.Client

	runCh          chan string
	shutdownDoneCh chan struct{}
}

func NewRoomsManager(assets *spine.AssetManager, twitchConfig *misc.TwitchConfig) *RoomsManager {
	return &RoomsManager{
		Rooms:        make(map[string]*Room, 0),
		AssetManager: assets,
		TwitchConfig: twitchConfig,
		twitchClient: twitch_api.NewClient(
			twitchConfig.TwitchClientId,
			twitchConfig.TwitchAccessToken,
		),
		runCh:          make(chan string),
		shutdownDoneCh: make(chan struct{}),
	}
}

func (r *RoomsManager) garbageCollectRooms() {
	log.Println("Garbage collecting unused chat rooms")
	period := time.Duration(r.TwitchConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
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

func (r *RoomsManager) RunLoop() {
	if r.TwitchConfig.RemoveUnusedRoomsAfterMinutes > 0 {
		stopTimer := misc.StartTimer(
			"garbageCollectRooms",
			time.Duration(r.TwitchConfig.RemoveUnusedRoomsAfterMinutes)*time.Minute,
			r.garbageCollectRooms,
		)
		defer stopTimer()
	}

	for channel := range r.runCh {
		go r.Rooms[channel].Run()
	}
}

func (r *RoomsManager) checkChannelValid(channel string) (bool, error) {
	resp, err := r.twitchClient.GetUsers(channel)
	if err != nil {
		return false, err
	}
	if len(resp.Data) == 0 {
		return false, fmt.Errorf("channel does not exist")
	}
	return true, nil
}

func (r *RoomsManager) CreateRoomOrNoOp(channel string, ctx context.Context) error {
	// Check to see if channel is valid
	if valid, err := r.checkChannelValid(channel); err != nil || !valid {
		return err
	}
	if _, ok := r.Rooms[channel]; ok {
		return nil
	}

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
	log.Println("RoomsManager: Waiting for shutdown")
	select {
	case <-time.After(forcefulShutdownDuration):
		log.Printf("RoomsManager: Shutting down forcefully after %v", forcefulShutdownDuration)
	case <-r.shutdownDoneCh:
		log.Println("RoomsManager: Shutting down gracefully")
	}
}

func (r *RoomsManager) RemoveRoom(channel string) error {
	if _, ok := r.Rooms[channel]; !ok {
		return nil
	}
	r.rooms_mutex.Lock()
	defer r.rooms_mutex.Unlock()
	r.Rooms[channel].Close()
	delete(r.Rooms, channel)
	return nil
}
