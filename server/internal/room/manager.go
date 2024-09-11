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

	AssetManager *spine.AssetManager
	SpineService *spine.SpineService

	BotConfig      *misc.BotConfig
	twitchClient   *twitch_api.Client
	shutdownDoneCh chan struct{}
}

func NewRoomsManager(assets *spine.AssetManager, botConfig *misc.BotConfig) *RoomsManager {
	spineService := spine.NewSpineService(assets)
	return &RoomsManager{
		Rooms:        make(map[string]*Room, 0),
		AssetManager: assets,
		SpineService: spineService,
		BotConfig:    botConfig,
		twitchClient: twitch_api.NewClient(
			botConfig.TwitchClientId,
			botConfig.TwitchAccessToken,
		),
		shutdownDoneCh: make(chan struct{}),
	}
}

func (r *RoomsManager) garbageCollectRooms() {
	log.Println("Garbage collecting unused chat rooms")
	period := time.Duration(r.BotConfig.RemoveUnusedRoomsAfterMinutes) * time.Minute
	r.rooms_mutex.Lock()
	for channel, room := range r.Rooms {
		lastChat := room.TwitchChat.LastChatterTime()
		if time.Since(lastChat) > period {
			log.Println("Removing unused room", channel)
			room.Close()
			delete(r.Rooms, channel)
		}
	}
	r.rooms_mutex.Unlock()
}

func (r *RoomsManager) RunLoop() {
	if r.BotConfig.RemoveUnusedRoomsAfterMinutes > 0 {
		stopTimer := misc.StartTimer(
			"garbageCollectRooms",
			time.Duration(r.BotConfig.RemoveUnusedRoomsAfterMinutes)*time.Minute,
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

	spineBridge, err := spine.NewSpineBridge(r.SpineService)
	if err != nil {
		return err
	}
	chibiActor := chibi.NewChibiActor(r.SpineService, spineBridge, r.BotConfig.ExcludeNames)
	twitchBot, err := chatbot.NewTwitchBot(
		chibiActor,
		channel,
		r.BotConfig.TwitchBot,
		r.BotConfig.TwitchAccessToken,
		r.BotConfig.RemoveChibiAfterMinutes,
	)
	if err != nil {
		return err
	}

	r.rooms_mutex.Lock()
	room := NewRoom(
		channel,
		r.BotConfig.InitialOperator,
		r.BotConfig.OperatorDetails,
		r.SpineService,
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

	chatters := make([]*spine.ChatUser, 0)
	for _, chatUser := range room.ChibiActor.ChatUsers {
		chatters = append(chatters, chatUser)
	}

	return room.SpineBridge.AddWebsocketConnection(w, r, chatters)
}

func (r *RoomsManager) Restore() error {
	ctx := context.Background()
	client, err := CreateFirestoreClient(
		ctx,
		r.BotConfig.GoogleCloudProjectId,
		r.BotConfig.GoogleCloudProjectCredentialsFilePath)
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
		r.BotConfig.GoogleCloudProjectId,
		r.BotConfig.GoogleCloudProjectCredentialsFilePath)
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
