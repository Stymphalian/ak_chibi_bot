package room

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chatbot"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

type RoomConfig struct {
	ChannelName                 string
	DefaultOperatorName         string
	DefaultOperatorConfig       misc.InitialOperatorDetails
	GarbageCollectionPeriodMins int
	InactiveRoomPeriodMins      int
}

// Displays (View)
// Communcations
// Actors (chatters) (Model)
type Room struct {
	SpineService *spine.SpineService

	config      RoomConfig
	SpineBridge *spine.SpineBridge
	ChibiActor  *chibi.ChibiActor
	TwitchChat  *chatbot.TwitchBot
}

func NewRoom(
	roomConfig *RoomConfig,
	spineService *spine.SpineService,
	spineBridge *spine.SpineBridge,
	chibiActor *chibi.ChibiActor,
	twitchBot *chatbot.TwitchBot,
) *Room {
	r := &Room{
		config:       *roomConfig,
		SpineService: spineService,
		SpineBridge:  spineBridge,
		ChibiActor:   chibiActor,
		TwitchChat:   twitchBot,
	}
	r.ChibiActor.SetToDefault(
		r.config.ChannelName,
		r.config.DefaultOperatorName,
		r.config.DefaultOperatorConfig)
	return r
}

func (r *Room) GetChannelName() string {
	return r.config.ChannelName
}

func (r *Room) GetLastChatterTime() time.Time {
	return r.ChibiActor.LastChatterTime
}

func (r *Room) Close() error {
	log.Println("Closing room ", r.config.ChannelName)
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
	log.Println("Closed room ", r.config.ChannelName, "successfully")
	return nil
}

func (r *Room) garbageCollectOldChibis() {
	log.Printf("Garbage collecting old chibis from room %s", r.config.ChannelName)
	interval := time.Duration(r.config.GarbageCollectionPeriodMins) * time.Minute
	for username, chatUser := range r.ChibiActor.ChatUsers {
		if username == r.config.ChannelName {
			// Skip removing the broadcaster's chibi
			continue
		}
		if !chatUser.IsActive(interval) {
			log.Println("Removing chibi for", username)
			r.ChibiActor.RemoveUserChibi(username)
		}
	}
}

func (r *Room) Run() {
	log.Printf("Room %s is running\n", r.config.ChannelName)

	if r.config.GarbageCollectionPeriodMins > 0 {
		stopTimer := misc.StartTimer(
			fmt.Sprintf("GarbageCollectOldChibis %s", r.config.ChannelName),
			time.Duration(r.config.GarbageCollectionPeriodMins)*time.Minute,
			r.garbageCollectOldChibis,
		)
		defer stopTimer()
	}

	err := r.TwitchChat.ReadPump()
	if err != nil {
		log.Printf("Room %s run error=", err)
	}
	log.Printf("Room %s run is done\n", r.config.ChannelName)
}

func (r *Room) GetChatters() []spine.ChatUser {
	chatters := make([]spine.ChatUser, 0)
	for _, chatter := range r.ChibiActor.ChatUsers {
		chatters = append(chatters, *chatter)
	}
	return chatters
}

func (r *Room) AddOperatorToRoom(
	username string,
	usernameDisplay string,
	operatorId string,
	faction spine.FactionEnum,
) error {
	// TODO: Leaky interface. Need to move this into a Service or ChibiActor
	opInfo, err := r.SpineService.GetRandomOperator()
	if err != nil {
		return err
	}
	opInfo.OperatorId = operatorId
	opInfo.Faction = faction

	err = r.ChibiActor.UpdateChibi(username, usernameDisplay, opInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *Room) AddWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
	// TODO: Maybe just pass in the map directly
	chatters := make([]*spine.ChatUser, 0)
	for _, chatUser := range s.ChibiActor.ChatUsers {
		chatters = append(chatters, chatUser)
	}
	return s.SpineBridge.AddWebsocketConnection(w, r, chatters)
}

func (r *Room) IsActive(period time.Duration) bool {
	return time.Since(r.ChibiActor.LastChatterTime) <= period
}

func (r *Room) NumConnectedClients() int {
	return len(r.SpineBridge.WebSocketConnections)
}

func (r *Room) ForEachChatter(callback func(chatUser *spine.ChatUser)) {
	for _, chatUser := range r.ChibiActor.ChatUsers {
		callback(chatUser)
	}
}

// TODO: Leaky interface. Exposing all the ChibiActor methods through the Room
func (r *Room) RemoveUserChibi(username string) error {
	return r.ChibiActor.RemoveUserChibi(username)
}
