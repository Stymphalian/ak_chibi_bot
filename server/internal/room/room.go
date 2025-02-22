package room

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chatbot"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	spine "github.com/Stymphalian/ak_chibi_bot/server/internal/spine_runtime"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

type RoomConfig struct {
	ChannelName                 string
	DefaultOperatorName         string
	DefaultOperatorConfig       misc.InitialOperatorDetails
	GarbageCollectionPeriodMins int
	InactiveRoomPeriodMins      int
	SpineRuntimeConfig          *misc.SpineRuntimeConfig
}

// View - spineRuntime
// Model - chibiActor
// View-Model/Controller - twitchChat
type Room struct {
	operatorService           *operator.OperatorService
	roomId                    uint
	channelName               string
	roomRepo                  RoomRepository
	usersRepo                 users.UserRepository
	chatterRepo               users.ChatterRepository
	spineRuntime              spine.SpineRuntime
	chibiActor                *chibi.ChibiActor
	chatBots                  []chatbot.ChatBotter
	createdAt                 time.Time
	nextGarbageCollectionTime time.Time
	isClosed                  bool
	removeRoomCh              chan string
	removalFns                []func()
}

func NewRoom(
	roomId uint,
	chanelName string,
	roomRepo RoomRepository,
	usersRepo users.UserRepository,
	chattersRepo users.ChatterRepository,
	operatorService *operator.OperatorService,
	spineRuntime spine.SpineRuntime,
	chibiActor *chibi.ChibiActor,
	chatBots []chatbot.ChatBotter,
	removeRoomCh chan string) (*Room, error) {
	r := &Room{
		roomId:          roomId,
		channelName:     chanelName,
		roomRepo:        roomRepo,
		usersRepo:       usersRepo,
		chatterRepo:     chattersRepo,
		operatorService: operatorService,
		spineRuntime:    spineRuntime,
		chibiActor:      chibiActor,
		chatBots:        chatBots,
		createdAt:       misc.Clock.Now(),
		isClosed:        false,
		removeRoomCh:    removeRoomCh,
	}

	removeFn, err := spineRuntime.AddListenerToClientRequests(r.handleClientWebsocketRequests)
	if err != nil {
		return nil, err
	}
	r.removalFns = append(r.removalFns, removeFn)

	return r, nil
}

func (r *Room) GetChannelName() string {
	return r.channelName
}

func (r *Room) GetLastChatterTime() time.Time {
	return r.chibiActor.GetLastChatterTime()
}

func (r *Room) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Room) GetNextGarbageCollectionTime() time.Time {
	return r.nextGarbageCollectionTime
}

func (r *Room) SetActive(isActive bool) {
	r.roomRepo.SetRoomActiveById(context.Background(), r.roomId, isActive)
}

func (r *Room) Close() error {
	log.Println("Closing room ", r.GetChannelName())
	r.isClosed = true

	if !r.roomRepo.IsRoomActiveById(context.Background(), r.roomId) {
		// If the room is inactive, we can clear out all the chibis/chatters
		err := r.chibiActor.Close()
		if err != nil {
			log.Println("Failed to close ChibiActor", err)
		}
	}

	// Disconnect the twitch chat and other chats bots
	for _, chatBot := range r.chatBots {
		err := chatBot.Close()
		if err != nil {
			log.Println("Failed to close ChatBot", err)
		}
	}

	// Disconnect all websockets
	err := r.spineRuntime.Close()
	if err != nil {
		log.Println("Failed to close SpineRuntime", err)
	}

	if r.removeRoomCh != nil {
		r.removeRoomCh <- r.GetChannelName()
	}
	log.Println("Closed room ", r.GetChannelName(), "successfully")
	return nil
}

func (r *Room) garbageCollectOldChibis(interval time.Duration) {
	log.Printf("Garbage collecting old chibis from room %s", r.GetChannelName())

	ctx := context.Background()
	numRemoved := 0
	numRemovedErr := 0
	for username, chatUser := range r.chibiActor.ChatUsers {
		if username == r.GetChannelName() {
			// Skip removing the broadcaster's chibi
			continue
		}
		if !chatUser.IsActiveChatter(interval) {
			log.Println("Removing chibi for", username)
			err := r.chibiActor.RemoveUserChibi(ctx, username)
			if err != nil {
				numRemovedErr += 1
			} else {
				numRemoved += 1
			}
		}
	}

	r.nextGarbageCollectionTime = time.Now().Add(interval)
	log.Printf(
		"Finished garbage collecting old chibis from room %s, %d removed, %d errors\n",
		r.GetChannelName(), numRemoved, numRemovedErr,
	)
}

func (r *Room) Run() {
	// misc.GoRunCounter.Add(1)
	// defer misc.GoRunCounter.Add(-1)
	log.Printf("Room %s is running\n", r.GetChannelName())

	GCPeriodMins := r.roomRepo.GetRoomGarbageCollectionPeriodMins(context.Background(), r.roomId)
	if GCPeriodMins > 0 {
		period := time.Duration(GCPeriodMins) * time.Minute
		r.nextGarbageCollectionTime = time.Now().Add(period)
		stopTimer := misc.StartTimer(
			fmt.Sprintf("GarbageCollectOldChibis %s", r.GetChannelName()),
			period,
			func() { r.garbageCollectOldChibis(period) },
		)
		defer stopTimer()
	}

	wg := sync.WaitGroup{}
	for _, chatBot := range r.chatBots {
		wg.Add(1)
		go func() {
			err := chatBot.ReadLoop()
			if err != nil {
				log.Printf("Room %s run error=", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if !r.isClosed {
		log.Printf("Closing room %s due to Readloop finishing early\n", r.channelName)
		r.SetActive(false)
		r.Close()
	}
	log.Printf("Room %s run is done\n", r.GetChannelName())
}

func (r *Room) GetChatters() []users.ChatUser {
	chatters := make([]users.ChatUser, 0)
	for _, chatter := range r.chibiActor.ChatUsers {
		chatters = append(chatters, *chatter)
	}
	return chatters
}

func (r *Room) AddOperatorToRoom(
	ctx context.Context,
	userinfo misc.UserInfo,
	operatorId string,
	faction operator.FactionEnum,
	skin string,
	stance operator.ChibiStanceEnum,
	startPos misc.Vector2,
) error {
	// TODO: Leaky interface. Need to move this into a Service or ChibiActor
	opInfo, err := r.operatorService.GetRandomOperator()
	if err != nil {
		return err
	}
	opInfo.OperatorId = operatorId
	opInfo.Faction = faction
	opInfo.Skin = skin
	opInfo.CurrentAction = operator.ACTION_PLAY_ANIMATION
	opInfo.Action = operator.NewActionPlayAnimation([]string{"Default"})
	opInfo.StartPos = misc.NewOption(startPos)
	opInfo.ChibiStance = stance

	err = r.chibiActor.UpdateChibi(ctx, userinfo, opInfo)
	if err != nil {
		return err
	}
	return nil
}

func (r *Room) GiveChibiToUser(ctx context.Context, userInfo misc.UserInfo) error {
	return r.chibiActor.GiveChibiToUser(ctx, userInfo)
}

func (s *Room) AddWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	// TODO: Maybe just pass in the map directly
	chatters := make([]*spine.ChatterInfo, 0)
	for _, chatUser := range s.chibiActor.ChatUsers {
		chatters = append(chatters, &spine.ChatterInfo{
			Username:        chatUser.GetUsername(),
			UsernameDisplay: chatUser.GetUsernameDisplay(),
			OperatorInfo:    *chatUser.GetOperatorInfo(),
		})
	}
	return s.spineRuntime.AddConnection(w, r, chatters)
}

func (r *Room) HasActiveChatters(period time.Duration) bool {
	return misc.Clock.Since(r.chibiActor.GetLastChatterTime()) <= period
}

func (r *Room) NumConnectedClients() int {
	return r.spineRuntime.NumConnections()
}

func (r *Room) ForEachChatter(callback func(chatUser *users.ChatUser)) {
	for _, chatUser := range r.chibiActor.ChatUsers {
		callback(chatUser)
	}
}

func (r *Room) LoadExistingChatters(ctx context.Context) error {
	chatters, err := r.chatterRepo.GetActiveChatters(ctx, r.GetRoomId())
	if err != nil {
		return err
	}
	for _, chatter := range chatters {
		user, err := r.usersRepo.GetById(ctx, chatter.UserId)
		if err != nil {
			continue
		}
		if r.chibiActor.ShouldExcludeUser(user.Username) {
			log.Println("Excluding user ", user.Username)
			r.chibiActor.RemoveUserChibi(ctx, user.Username)
			continue
		}

		log.Printf("Reloading chatter %s in room %s with operator %s", user.Username, r.GetChannelName(), chatter.OperatorInfo.OperatorDisplayName)
		err = r.chibiActor.UpdateChibi(
			ctx,
			misc.UserInfo{
				Username:        user.Username,
				UsernameDisplay: user.UserDisplayName,
				TwitchUserId:    user.TwitchUserId,
			},
			&chatter.OperatorInfo,
		)
		if err != nil {
			log.Printf("Failed to reload chatter %s in room %s: %s\n", user.Username, r.GetChannelName(), err)
		}
	}
	return nil
}

// TODO: Leaky interface. Exposing all the ChibiActor methods through the Room
func (r *Room) RemoveUserChibi(ctx context.Context, username string) error {
	return r.chibiActor.RemoveUserChibi(ctx, username)
}

func (r *Room) GetSpineRuntimeConfig(ctx context.Context) (*misc.SpineRuntimeConfig, error) {
	return r.roomRepo.GetSpineRuntimeConfigById(ctx, r.roomId)
}

func (r *Room) GetRoomId() uint {
	return r.roomId
}

func (r *Room) RefreshConfigs(ctx context.Context, botConfig *misc.BotConfig) error {
	newConfig, err := r.GetSpineRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	r.operatorService.SetConfig(newConfig)
	r.chibiActor.UpdateExcludeNames(
		append(botConfig.ExcludeNames, newConfig.UsernamesBlacklist...),
	)
	return nil
}

func (r *Room) Refresh(ctx context.Context, botConfig *misc.BotConfig) error {
	err := r.RefreshConfigs(ctx, botConfig)
	if err != nil {
		return err
	}
	return r.LoadExistingChatters(ctx)
}

func (r *Room) handleClientWebsocketRequests(connectionName string, typeName string, message []byte) {
}
