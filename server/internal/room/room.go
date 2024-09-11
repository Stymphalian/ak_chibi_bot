package room

import (
	"log"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chatbot"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/chibi"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

type Room struct {
	ChannelName string

	DefaultOperatorName   string
	DefaultOperatorConfig misc.InitialOperatorDetails
	SpineService          *spine.SpineService

	// Displays (View)
	// Communcations
	// Actors (chatters) (Model)
	SpineBridge *spine.SpineBridge
	ChibiActor  *chibi.ChibiActor
	TwitchChat  *chatbot.TwitchBot
}

func NewRoom(
	channel string,
	opName string,
	opConfig misc.InitialOperatorDetails,
	spineService *spine.SpineService,
	spineBridge *spine.SpineBridge,
	chibiActor *chibi.ChibiActor,
	twitchBot *chatbot.TwitchBot,
) *Room {
	r := &Room{
		ChannelName:           channel,
		DefaultOperatorName:   opName,
		DefaultOperatorConfig: opConfig,
		SpineService:          spineService,
		SpineBridge:           spineBridge,
		ChibiActor:            chibiActor,
		TwitchChat:            twitchBot,
	}
	r.ChibiActor.SetToDefault(r.ChannelName, r.DefaultOperatorName, r.DefaultOperatorConfig)
	return r
}

func (r *Room) Close() error {
	log.Println("Closing room ", r.ChannelName)
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
	log.Println("Closed room ", r.ChannelName, "successfully")
	return nil
}

func (r *Room) Run() {
	log.Printf("Room %s is running\n", r.ChannelName)
	err := r.TwitchChat.ReadPump()
	if err != nil {
		log.Printf("Room %s run error=", err)
	}
	log.Printf("Room %s run is done\n", r.ChannelName)
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
