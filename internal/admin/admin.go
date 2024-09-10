package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/room"
)

type AdminServer struct {
	roomsManager *room.RoomsManager
	botConfig    *misc.BotConfig
	assetDir     string
}

type Chatter struct {
	Username     string
	Operator     string
	LastChatTime string
}

type Room struct {
	ChannelName             string
	LastTimeUsed            string
	Chatters                []Chatter
	NumWebsocketConnections int
}
type AdminInfo struct {
	Rooms []Room
}

type RemoveRoomRequest struct {
	ChannelName string `json:"channel_name"`
}

type RemoveUserRequest struct {
	ChannelName string `json:"channel_name"`
	Username    string `json:"username"`
}

func NewAdminServer(roomManager *room.RoomsManager, botConfig *misc.BotConfig, adminAssetDir string) *AdminServer {
	return &AdminServer{
		roomsManager: roomManager,
		botConfig:    botConfig,
		assetDir:     adminAssetDir,
	}
}

func (a *AdminServer) adminAuth(h misc.HandlerWithErr) misc.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		if !r.URL.Query().Has("secret") {
			http.Error(w, "", http.StatusNotFound)
			return nil
		}
		if len(a.botConfig.AdminSecret) != 256 {
			http.Error(w, "", http.StatusNotFound)
			return nil
		}
		secret := r.URL.Query().Get("secret")
		if len(secret) != 256 {
			http.Error(w, "", http.StatusNotFound)
			return nil
		}
		if a.botConfig.AdminSecret != secret {
			http.Error(w, "", http.StatusNotFound)
			return nil
		}
		return h(w, r)
	}
}

func (s *AdminServer) middleware(h misc.HandlerWithErr) http.Handler {
	return misc.Middleware(s.adminAuth(h))
}

func (s *AdminServer) RegisterAdmin() {
	http.Handle("/admin", s.middleware(s.HandleAdmin))
	http.Handle("/admin/list", s.middleware(s.HandleList))
	http.Handle("/admin/room/remove", s.middleware(s.HandleRemoveRoom))
	http.Handle("/admin/user/remove", s.middleware(s.HandleRemoveUser))
}

func (s *AdminServer) HandleAdmin(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles(fmt.Sprintf("%s/index.html", s.assetDir))
	if err != nil {
		return err
	}
	t.Execute(w, nil)
	return nil
}

func (s *AdminServer) HandleList(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")

	var adminInfo AdminInfo
	adminInfo.Rooms = make([]Room, 0)
	for _, roomVal := range s.roomsManager.Rooms {

		newRoom := Room{
			ChannelName:             roomVal.ChannelName,
			LastTimeUsed:            roomVal.TwitchChat.LastChatterTime().Format(time.DateTime),
			Chatters:                make([]Chatter, 0),
			NumWebsocketConnections: len(roomVal.SpineBridge.WebSocketConnections),
		}

		for _, chatUser := range roomVal.SpineBridge.ChatUsers {
			newChatter := Chatter{
				Username:     chatUser.UserName,
				Operator:     chatUser.CurrentOperator.OperatorDisplayName,
				LastChatTime: "",
			}

			lastChatTime, ok := roomVal.TwitchChat.LastChatTime(chatUser.UserName)
			if ok {
				newChatter.LastChatTime = lastChatTime.Format(time.DateTime)
			}
			newRoom.Chatters = append(newRoom.Chatters, newChatter)
		}
		adminInfo.Rooms = append(adminInfo.Rooms, newRoom)
	}
	json.NewEncoder(w).Encode(adminInfo)
	return nil
}

func (s *AdminServer) HandleRemoveRoom(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	var reqBody RemoveRoomRequest
	if err := decoder.Decode(&reqBody); err != nil {
		return err
	}

	channelName := reqBody.ChannelName
	if len(channelName) == 0 {
		return nil
	}
	if _, ok := s.roomsManager.Rooms[channelName]; !ok {
		return nil
	}

	return s.roomsManager.RemoveRoom(channelName)
}

func (s *AdminServer) HandleRemoveUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	var reqBody RemoveUserRequest
	if err := decoder.Decode(&reqBody); err != nil {
		return err
	}

	channelName := reqBody.ChannelName
	if len(channelName) == 0 {
		return nil
	}
	if _, ok := s.roomsManager.Rooms[channelName]; !ok {
		return nil
	}
	room := s.roomsManager.Rooms[channelName]

	userName := reqBody.Username
	if len(userName) == 0 {
		return nil
	}
	return room.ChibiActor.RemoveUserChibi(userName)
}
