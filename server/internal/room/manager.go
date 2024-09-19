package room

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chatbot"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
)

const (
	forcefulShutdownDuration = time.Duration(30) * time.Second
)

type RoomsManager struct {
	Rooms                     map[string]*Room
	rooms_mutex               sync.Mutex
	nextGarbageCollectionTime time.Time

	assetService *spine.AssetService
	spineService *spine.SpineService

	botConfig      *misc.BotConfig
	twitchClient   twitch_api.TwitchApiClientInterface
	shutdownDoneCh chan struct{}
}

func NewRoomsManager(assets *spine.AssetService, botConfig *misc.BotConfig) *RoomsManager {
	spineService := spine.NewSpineService(assets, botConfig.SpineRuntimeConfig)
	return &RoomsManager{
		Rooms:        make(map[string]*Room, 0),
		assetService: assets,
		spineService: spineService,
		botConfig:    botConfig,
		twitchClient: twitch_api.NewTwitchApiClient(
			botConfig.TwitchClientId,
			botConfig.TwitchAccessToken,
		),
		shutdownDoneCh: make(chan struct{}),
	}
}

func (r *RoomsManager) LoadExistingRooms(ctx context.Context) error {
	log.Println("Reloading Existing rooms")
	roomDbs, err := GetActiveRooms(ctx)
	if err != nil {
		return err
	}

	for _, roomDb := range roomDbs {
		log.Println("Reloading room", roomDb.ChannelName)
		r.InsertRoom(roomDb)
		room := r.Rooms[roomDb.ChannelName]
		room.LoadExistingChatters(ctx)
	}
	return nil
}

func (r *RoomsManager) garbageCollectRooms() {
	log.Println("Garbage collecting unused chat rooms")
	period := time.Duration(r.botConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
	r.rooms_mutex.Lock()
	for channel, room := range r.Rooms {
		if !room.HasActiveChatters(period) {
			log.Println("Removing unused room", channel)
			room.SetActive(false)
			room.Close()
			delete(r.Rooms, channel)
		}
	}
	r.rooms_mutex.Unlock()

	r.nextGarbageCollectionTime = time.Now().Add(period)
}

func (r *RoomsManager) GetNextGarbageCollectionTime() time.Time {
	return r.nextGarbageCollectionTime
}

func (r *RoomsManager) RunLoop() {
	// misc.GoRunCounter.Add(1)
	// defer misc.GoRunCounter.Add(-1)

	if r.botConfig.RemoveUnusedRoomsAfterMinutes > 0 {
		period := time.Duration(r.botConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
		r.nextGarbageCollectionTime = time.Now().Add(period)

		stopTimer := misc.StartTimer(
			"garbageCollectRooms",
			period,
			r.garbageCollectRooms,
		)
		defer stopTimer()
	}
	<-r.shutdownDoneCh
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

func (r *RoomsManager) getRoomServices(channelName string) (*spine.SpineBridge, *chibi.ChibiActor, *chatbot.TwitchBot, error) {
	spineBridge, err := spine.NewSpineBridge(r.spineService)
	if err != nil {
		return nil, nil, nil, err
	}
	chibiActor := chibi.NewChibiActor(r.spineService, spineBridge, r.botConfig.ExcludeNames)
	twitchBot, err := chatbot.NewTwitchBot(
		chibiActor,
		channelName,
		r.botConfig.TwitchBot,
		r.botConfig.TwitchAccessToken,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return spineBridge, chibiActor, twitchBot, nil
}

func (r *RoomsManager) InsertRoom(roomDb *RoomDb) error {
	spineBridge, chibiActor, twitchBot, err := r.getRoomServices(roomDb.ChannelName)
	if err != nil {
		return err
	}
	log.Println("Inserting room: ", roomDb.ChannelName)
	room := NewRoom(
		roomDb,
		r.spineService,
		spineBridge,
		chibiActor,
		twitchBot,
	)

	r.rooms_mutex.Lock()
	r.Rooms[roomDb.ChannelName] = room
	r.rooms_mutex.Unlock()

	misc.Monitor.NumRoomsCreated += 1
	go room.Run()
	return nil
}

func (r *RoomsManager) CreateRoomOrNoOp(ctx context.Context, channel string) error {
	// Check to see if channel is valid
	if valid, err := r.checkChannelValid(channel); err != nil || !valid {
		return err
	}
	if _, ok := r.Rooms[channel]; ok {
		return nil
	}
	spineBridge, chibiActor, twitchBot, err := r.getRoomServices(channel)
	if err != nil {
		return err
	}

	roomConfig := &RoomConfig{
		ChannelName:                 channel,
		DefaultOperatorName:         r.botConfig.InitialOperator,
		DefaultOperatorConfig:       r.botConfig.OperatorDetails,
		GarbageCollectionPeriodMins: r.botConfig.RemoveChibiAfterMinutes,
		SpineRuntimeConfig:          r.botConfig.SpineRuntimeConfig,
	}
	room, err := GetOrNewRoom(
		ctx,
		roomConfig,
		r.spineService,
		spineBridge,
		chibiActor,
		twitchBot,
	)
	if err != nil {
		return err
	}

	r.rooms_mutex.Lock()
	r.Rooms[channel] = room
	r.rooms_mutex.Unlock()

	misc.Monitor.NumRoomsCreated += 1
	go room.Run()
	return nil
}

func (m *RoomsManager) HandleSpineWebSocket(channelName string, w http.ResponseWriter, r *http.Request) error {
	room, ok := m.Rooms[channelName]
	if !ok {
		return errors.New("channel room does not exist")
	}
	return room.AddWebsocketConnection(w, r)
}

func (r *RoomsManager) Shutdown() {
	log.Println("RoomsManager calling Shutdown")

	go func() {
		// misc.GoRunCounter.Add(1)
		// defer misc.GoRunCounter.Add(-1)

		defer close(r.shutdownDoneCh)
		r.rooms_mutex.Lock()
		defer r.rooms_mutex.Unlock()
		for _, room := range r.Rooms {
			// close the room, but don't set it as inactive
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

	// log.Println("Number of running gorountines:", misc.GoRunCounter.Load())
}

func (r *RoomsManager) RemoveRoom(channel string) error {
	if _, ok := r.Rooms[channel]; !ok {
		return nil
	}
	r.rooms_mutex.Lock()
	defer r.rooms_mutex.Unlock()
	r.Rooms[channel].SetActive(false)
	r.Rooms[channel].Close()
	delete(r.Rooms, channel)
	return nil
}
