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
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"

	spine "github.com/Stymphalian/ak_chibi_bot/server/internal/spine_runtime"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
)

type RoomsManager struct {
	Rooms                     map[string]*Room
	rooms_mutex               sync.Mutex
	nextGarbageCollectionTime time.Time

	assetService *operator.AssetService
	spineService *operator.OperatorService
	roomRepo     RoomRepository
	usersRepo    users.UserRepository
	chattersRepo users.ChatterRepository

	botConfig      *misc.BotConfig
	twitchClient   twitch_api.TwitchApiClientInterface
	shutdownDoneCh chan struct{}
}

func NewRoomsManager(
	assets *operator.AssetService,
	roomRepo RoomRepository,
	usersRepo users.UserRepository,
	chattersRepo users.ChatterRepository,
	botConfig *misc.BotConfig,
) *RoomsManager {
	spineService := operator.NewOperatorService(assets, botConfig.SpineRuntimeConfig)
	return &RoomsManager{
		Rooms:        make(map[string]*Room, 0),
		assetService: assets,
		spineService: spineService,

		roomRepo:     roomRepo,
		usersRepo:    usersRepo,
		chattersRepo: chattersRepo,

		botConfig: botConfig,
		twitchClient: twitch_api.NewTwitchApiClient(
			botConfig.TwitchClientId,
			botConfig.TwitchAccessToken,
		),
		shutdownDoneCh: make(chan struct{}),
	}
}

func (r *RoomsManager) LoadExistingRooms(ctx context.Context) error {
	log.Println("Reloading Existing rooms")
	roomDbs, err := r.roomRepo.GetActiveRooms(ctx)
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

func (r *RoomsManager) checkChannelValid(channel string) (*misc.UserInfo, error) {
	resp, err := r.twitchClient.GetUsers(channel)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("channel does not exist")
	}
	return &misc.UserInfo{
		Username:        resp.Data[0].Login,
		UsernameDisplay: resp.Data[0].DisplayName,
		TwitchUserId:    resp.Data[0].Id,
	}, nil
}

func (r *RoomsManager) getRoomServices(roomDb *RoomDb) (*operator.OperatorService, *spine.SpineBridge, *chibi.ChibiActor, *chatbot.TwitchBot, error) {
	channelName := roomDb.ChannelName
	spineRuntimeConfig, err := r.roomRepo.GetSpineRuntimeConfigById(context.Background(), roomDb.RoomId)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	newSpineService := r.spineService.WithConfig(spineRuntimeConfig)

	spineBridge, err := spine.NewSpineBridge(newSpineService)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	chibiActor := chibi.NewChibiActor(
		roomDb.RoomId,
		newSpineService,
		r.usersRepo,
		r.chattersRepo,
		spineBridge,
		r.botConfig.ExcludeNames,
	)
	twitchBot, err := chatbot.NewTwitchBot(
		chibiActor,
		channelName,
		r.botConfig.TwitchBot,
		r.botConfig.TwitchAccessToken,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return newSpineService, spineBridge, chibiActor, twitchBot, nil
}

func (r *RoomsManager) InsertRoom(roomDb *RoomDb) error {
	spineService, spineBridge, chibiActor, twitchBot, err := r.getRoomServices(roomDb)
	if err != nil {
		return err
	}
	log.Println("Inserting room: ", roomDb.ChannelName)
	room := NewRoom(
		roomDb.RoomId,
		roomDb.ChannelName,
		r.roomRepo,
		r.usersRepo,
		r.chattersRepo,
		spineService,
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
	var userinfo *misc.UserInfo
	userinfo, err := r.checkChannelValid(channel)
	if err != nil {
		return err
	}

	if _, ok := r.Rooms[channel]; ok {
		// Refresh the room's configs, best effort
		r.Rooms[channel].RefreshConfigs(ctx)
		return nil
	}

	// Get the Room database object
	roomConfig := &RoomConfig{
		ChannelName:                 channel,
		DefaultOperatorName:         r.botConfig.InitialOperator,
		DefaultOperatorConfig:       r.botConfig.OperatorDetails,
		GarbageCollectionPeriodMins: r.botConfig.RemoveChibiAfterMinutes,
		SpineRuntimeConfig:          r.botConfig.SpineRuntimeConfig,
	}
	roomDb, isNew, err := r.roomRepo.GetOrInsertRoom(ctx, roomConfig)
	if err != nil {
		return err
	}
	roomWasInactive := !roomDb.IsActive
	roomDb.IsActive = true
	err = r.roomRepo.SetRoomActiveById(ctx, roomDb.RoomId, true)
	if err != nil {
		return err
	}

	// Create the services needed by the room
	spineService, spineBridge, chibiActor, twitchBot, err := r.getRoomServices(roomDb)
	if err != nil {
		return err
	}

	roomObj := NewRoom(
		roomDb.RoomId,
		roomDb.ChannelName,
		r.roomRepo,
		r.usersRepo,
		r.chattersRepo,
		spineService,
		spineBridge,
		chibiActor,
		twitchBot,
	)
	if isNew || roomWasInactive {
		log.Println("Adding default chibi for ", roomDb.ChannelName)
		roomObj.chibiActor.SetToDefault(
			ctx,
			*userinfo,
			roomDb.DefaultOperatorName,
			roomDb.DefaultOperatorConfig,
		)
	}

	r.rooms_mutex.Lock()
	r.Rooms[channel] = roomObj
	r.rooms_mutex.Unlock()

	misc.Monitor.NumRoomsCreated += 1
	go roomObj.Run()
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

func (r *RoomsManager) GetShutdownChan() chan struct{} {
	return r.shutdownDoneCh
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
