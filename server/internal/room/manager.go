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
	Rooms       map[string]*Room
	rooms_mutex sync.Mutex

	assetService *spine.AssetService
	spineService *spine.SpineService

	botConfig      *misc.BotConfig
	twitchClient   *twitch_api.Client
	shutdownDoneCh chan struct{}
}

func NewRoomsManager(assets *spine.AssetService, botConfig *misc.BotConfig) *RoomsManager {
	spineService := spine.NewSpineService(assets, botConfig.SpineRuntimeConfig)
	return &RoomsManager{
		Rooms:        make(map[string]*Room, 0),
		assetService: assets,
		spineService: spineService,
		botConfig:    botConfig,
		twitchClient: twitch_api.NewClient(
			botConfig.TwitchClientId,
			botConfig.TwitchAccessToken,
		),
		shutdownDoneCh: make(chan struct{}),
	}
}

func (r *RoomsManager) garbageCollectRooms() {
	log.Println("Garbage collecting unused chat rooms")
	period := time.Duration(r.botConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
	r.rooms_mutex.Lock()
	for channel, room := range r.Rooms {
		if !room.IsActive(period) {
			log.Println("Removing unused room", channel)
			room.Close()
			delete(r.Rooms, channel)
		}
	}
	r.rooms_mutex.Unlock()
}

func (r *RoomsManager) RunLoop() {
	// misc.GoRunCounter.Add(1)
	// defer misc.GoRunCounter.Add(-1)

	if r.botConfig.RemoveUnusedRoomsAfterMinutes > 0 {
		stopTimer := misc.StartTimer(
			"garbageCollectRooms",
			time.Duration(r.botConfig.RemoveUnusedRoomsAfterMinutes)*time.Minute,
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

func (r *RoomsManager) CreateRoomOrNoOp(channel string, ctx context.Context) error {
	// Check to see if channel is valid
	if valid, err := r.checkChannelValid(channel); err != nil || !valid {
		return err
	}
	if _, ok := r.Rooms[channel]; ok {
		return nil
	}

	spineBridge, err := spine.NewSpineBridge(r.spineService)
	if err != nil {
		return err
	}
	chibiActor := chibi.NewChibiActor(r.spineService, spineBridge, r.botConfig.ExcludeNames)
	twitchBot, err := chatbot.NewTwitchBot(
		chibiActor,
		channel,
		r.botConfig.TwitchBot,
		r.botConfig.TwitchAccessToken,
	)
	if err != nil {
		return err
	}

	r.rooms_mutex.Lock()
	roomConfig := &RoomConfig{
		ChannelName:                 channel,
		DefaultOperatorName:         r.botConfig.InitialOperator,
		DefaultOperatorConfig:       r.botConfig.OperatorDetails,
		GarbageCollectionPeriodMins: r.botConfig.RemoveChibiAfterMinutes,
		SpineRuntimeConfig:          r.botConfig.SpineRuntimeConfig,
	}
	room := NewRoom(
		roomConfig,
		r.spineService,
		spineBridge,
		chibiActor,
		twitchBot,
	)
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

func (r *RoomsManager) Restore() error {
	ctx := context.Background()
	client, err := CreateFirestoreClient(
		ctx,
		r.botConfig.GoogleCloudProjectId,
		r.botConfig.GoogleCloudProjectCredentialsFilePath)
	if err != nil {
		return err
	}
	defer client.Close()

	ref := client.Collection("backup").Doc("rooms")
	dsnap, err := ref.Get(ctx)
	if err != nil {
		return err
	}
	if !dsnap.Exists() {
		return nil
	}

	var saveData SaveData
	dsnap.DataTo(&saveData)
	// fmt.Printf("Saved Data %#v\n", saveData)
	RestoreSaveData(r, &saveData)

	// Remove the restore point from firestore
	_, err = ref.Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoomsManager) Save() error {
	ctx := context.Background()

	client, err := CreateFirestoreClient(
		ctx,
		r.botConfig.GoogleCloudProjectId,
		r.botConfig.GoogleCloudProjectCredentialsFilePath)
	if err != nil {
		return err
	}
	defer client.Close()
	saveData := CreateSaveData(r)

	_, err = client.Collection("backup").Doc("rooms").Set(ctx, saveData)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoomsManager) Shutdown() {
	log.Println("RoomsManager calling Shutdown")
	r.Save()

	go func() {
		// misc.GoRunCounter.Add(1)
		// defer misc.GoRunCounter.Add(-1)

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

	// log.Println("Number of running gorountines:", misc.GoRunCounter.Load())
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
