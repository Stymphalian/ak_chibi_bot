package room

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	twitchChat                chatbot.ChatBotter
	createdAt                 time.Time
	nextGarbageCollectionTime time.Time
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
	twitchBot chatbot.ChatBotter) *Room {
	r := &Room{
		roomId:          roomId,
		channelName:     chanelName,
		roomRepo:        roomRepo,
		usersRepo:       usersRepo,
		chatterRepo:     chattersRepo,
		operatorService: operatorService,
		spineRuntime:    spineRuntime,
		chibiActor:      chibiActor,
		twitchChat:      twitchBot,
		createdAt:       misc.Clock.Now(),
	}
	return r
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

	if !r.roomRepo.IsRoomActiveById(context.Background(), r.roomId) {
		// If the room is inactive, we can clear out all the chibis/chatters
		err := r.chibiActor.Close()
		if err != nil {
			return err
		}
	}

	// Disconnect the twitch chat
	err := r.twitchChat.Close()
	if err != nil {
		return err
	}

	// Disconnect all websockets
	err = r.spineRuntime.Close()
	if err != nil {
		return err
	}
	log.Println("Closed room ", r.GetChannelName(), "successfully")
	return nil
}

func (r *Room) garbageCollectOldChibis(interval time.Duration) {
	log.Printf("Garbage collecting old chibis from room %s", r.GetChannelName())

	ctx := context.Background()
	// interval := time.Duration(r.roomDb.GetGarbageCollectionPeriodMins()) * time.Minute
	for username, chatUser := range r.chibiActor.ChatUsers {
		if username == r.GetChannelName() {
			// Skip removing the broadcaster's chibi
			continue
		}
		if !chatUser.IsActiveChatter(interval) {
			log.Println("Removing chibi for", username)
			r.chibiActor.RemoveUserChibi(ctx, username)
		}
	}

	r.nextGarbageCollectionTime = time.Now().Add(interval)
}

func (r *Room) Run() {
	// misc.GoRunCounter.Add(1)
	// defer misc.GoRunCounter.Add(-1)
	var err error
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

	err = r.twitchChat.ReadLoop()
	if err != nil {
		log.Printf("Room %s run error=", err)
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
) error {
	// TODO: Leaky interface. Need to move this into a Service or ChibiActor
	opInfo, err := r.operatorService.GetRandomOperator()
	if err != nil {
		return err
	}
	opInfo.OperatorId = operatorId
	opInfo.Faction = faction

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
		log.Printf("Reloading chatter %s in room %s with operator %s", user.Username, r.GetChannelName(), chatter.OperatorInfo.OperatorDisplayName)
		r.chibiActor.UpdateChibi(
			ctx,
			misc.UserInfo{
				Username:        user.Username,
				UsernameDisplay: user.UserDisplayName,
				TwitchUserId:    user.TwitchUserId,
			},
			&chatter.OperatorInfo,
		)
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

func (r *Room) UpdateSpineRuntimeConfig(ctx context.Context, newConfig *misc.SpineRuntimeConfig) error {
	if err := misc.ValidateSpineRuntimeConfig(newConfig); err != nil {
		return err
	}
	runtimeConfig, err := r.roomRepo.GetSpineRuntimeConfigById(ctx, r.roomId)
	if err != nil {
		return err
	}
	if newConfig.MinAnimationSpeed > 0 {
		runtimeConfig.MinAnimationSpeed = newConfig.MinAnimationSpeed
	}
	if newConfig.MaxAnimationSpeed > 0 {
		runtimeConfig.MaxAnimationSpeed = newConfig.MaxAnimationSpeed
	}
	if newConfig.MinMovementSpeed > 0 {
		runtimeConfig.MinMovementSpeed = newConfig.MinMovementSpeed
	}
	if newConfig.MaxMovementSpeed > 0 {
		runtimeConfig.MaxMovementSpeed = newConfig.MaxMovementSpeed
	}
	if newConfig.MinScaleSize > 0 {
		runtimeConfig.MinScaleSize = newConfig.MinScaleSize
	}
	if newConfig.MaxScaleSize > 0 {
		runtimeConfig.MaxScaleSize = newConfig.MaxScaleSize
	}
	if newConfig.MaxSpritePixelSize > 0 {
		runtimeConfig.MaxSpritePixelSize = newConfig.MaxSpritePixelSize
	}
	if newConfig.ReferenceMovementSpeedPx > 0 {
		runtimeConfig.ReferenceMovementSpeedPx = newConfig.ReferenceMovementSpeedPx
	}
	if newConfig.DefaultAnimationSpeed > 0 {
		runtimeConfig.DefaultAnimationSpeed = newConfig.DefaultAnimationSpeed
	}
	if newConfig.DefaultMovementSpeed > 0 {
		runtimeConfig.DefaultMovementSpeed = newConfig.DefaultMovementSpeed
	}
	r.roomRepo.UpdateSpineRuntimeConfigForId(ctx, r.roomId, runtimeConfig)
	return nil
}

func (r *Room) GetRoomId() uint {
	return r.roomId
}

func (r *Room) RefreshConfigs(ctx context.Context) error {
	newConfig, err := r.GetSpineRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	r.operatorService.SetConfig(newConfig)
	return nil
}
