package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

type AdminServer struct {
	roomsManager *room.RoomsManager
	botConfig    *misc.BotConfig
	assetDir     string
}

type chatter struct {
	Username     string
	Operator     string
	LastChatTime string
}

type roomInfo struct {
	ChannelName             string
	CreatedAt               string
	LastTimeUsed            string
	Chatters                []*chatter
	NextGCTime              string
	NumWebsocketConnections int
	ConnectionAverageFps    map[string]float64
}

type AdminInfo struct {
	Rooms      []*roomInfo
	NextGCTime string
	Metrics    map[string]interface{}
}

type removeRoomRequest struct {
	ChannelName string `json:"channel_name"`
}

type removeUserRequest struct {
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
	return misc.MiddlewareWithTimeout(s.adminAuth(h), 5*time.Second)
}

func (s *AdminServer) RegisterHandlers() {
	http.Handle("GET /admin", s.middleware(s.HandleAdmin))
	http.Handle("GET /admin/list", s.middleware(s.HandleList))
	http.Handle("POST /admin/room/remove", s.middleware(s.HandleRemoveRoom))
	http.Handle("POST /admin/user/remove", s.middleware(s.HandleRemoveUser))
}

func (s *AdminServer) HandleAdmin(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles(fmt.Sprintf("%s/admin/index.html", s.assetDir))
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
	adminInfo.Rooms = make([]*roomInfo, 0)
	adminInfo.Metrics = make(map[string]interface{}, 0)
	adminInfo.NextGCTime = s.roomsManager.GetNextGarbageCollectionTime().Format(time.DateTime)

	for _, roomVal := range s.roomsManager.Rooms {

		newRoom := &roomInfo{
			ChannelName:             roomVal.GetChannelName(),
			LastTimeUsed:            roomVal.GetLastChatterTime().Format(time.DateTime),
			Chatters:                make([]*chatter, 0),
			NumWebsocketConnections: roomVal.NumConnectedClients(),
			CreatedAt:               roomVal.CreatedAt().Format(time.DateTime),
			NextGCTime:              roomVal.GetNextGarbageCollectionTime().Format(time.DateTime),
			ConnectionAverageFps:    make(map[string]float64),
		}

		roomVal.ForEachChatter(func(chatUser *users.ChatUser) {
			newChatter := &chatter{
				Username:     chatUser.GetUsername(),
				Operator:     chatUser.GetOperatorInfo().OperatorDisplayName,
				LastChatTime: chatUser.GetLastChatTime().Format(time.DateTime),
			}
			newRoom.Chatters = append(newRoom.Chatters, newChatter)
		})

		slices.SortFunc(newRoom.Chatters, func(a, b *chatter) int {
			return strings.Compare(a.Username, b.Username)
		})
		adminInfo.Rooms = append(adminInfo.Rooms, newRoom)
	}
	slices.SortFunc(adminInfo.Rooms, func(a, b *roomInfo) int {
		return strings.Compare(a.ChannelName, b.ChannelName)
	})

	adminInfo.Metrics["NumRoomsCreated"] = misc.Monitor.NumRoomsCreated
	adminInfo.Metrics["NumWebsocketConnections"] = misc.Monitor.NumWebsocketConnections
	adminInfo.Metrics["NumUsers"] = misc.Monitor.NumUsers
	adminInfo.Metrics["NumCommands"] = misc.Monitor.NumCommands
	adminInfo.Metrics["Datetime"] = misc.Clock.Now().Format(time.DateTime)

	json.NewEncoder(w).Encode(adminInfo)
	return nil
}

func (s *AdminServer) HandleRemoveRoom(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	var reqBody removeRoomRequest
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
	var reqBody removeUserRequest
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
	return room.RemoveUserChibi(userName)
}
