package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

type ApiServer struct {
	roomsManager *room.RoomsManager
	authService  auth.AuthServiceInterface
}

func NewApiServer(roomManager *room.RoomsManager, authService auth.AuthServiceInterface) *ApiServer {
	return &ApiServer{
		roomsManager: roomManager,
		authService:  authService,
	}
}

func (s *ApiServer) CheckAuth(h misc.HandlerWithErr, checkAdmin bool) misc.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) error {
		info, err := s.authService.IsAuthorized(w, r)
		if err != nil || !info.Authenticated {
			return misc.NewHumanReadableError(
				"not authorized",
				http.StatusUnauthorized,
				err,
			)
		}
		if checkAdmin {
			if !info.User.IsAdmin {
				return misc.NewHumanReadableError(
					"not authorized",
					http.StatusUnauthorized,
					err,
				)
			}
		}

		newContext := context.WithValue(
			r.Context(), auth.CONTEXT_TWITCH_USER_ID, info.User.TwitchUserId)
		newContext = context.WithValue(
			newContext, auth.CONTEXT_TWITCH_USER_NAME, info.User.Username)
		*r = *r.WithContext(newContext)
		return h(w, r)
	}
}

func (s *ApiServer) middleware(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(
		s.CheckAuth(h, false),
		5*time.Second,
	)
}

func (s *ApiServer) middlewareAdmin(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(
		s.CheckAuth(h, true),
		5*time.Second,
	)
}

func (s *ApiServer) RegisterHandlers() {
	mux := http.NewServeMux()
	mux.Handle("GET  /api/rooms/settings/{$}", s.middleware(s.HandleGetRoomSettings))
	mux.Handle("POST /api/rooms/settings/{$}", s.middleware(s.HandleUpdateRoomSettings))
	mux.Handle("POST /api/rooms/remove/{$}", s.middlewareAdmin(s.HandleRemoveRoom))
	mux.Handle("POST /api/rooms/users/remove/{$}", s.middlewareAdmin(s.HandleRemoveUser))
	mux.Handle("GET  /api/admin/info/{$}", s.middlewareAdmin(s.HandleAdminInfo))

	http.Handle("/api/", mux)
	// http.Handle("POST /api/rooms/users/add/{$}", s.middleware(s.HandleRoomAddOperator))
}

func (s *ApiServer) matchRequestChannel(r *http.Request, channelName string) error {
	twitchUserName := r.Context().Value(auth.CONTEXT_TWITCH_USER_NAME)
	if twitchUserName == nil {
		return misc.NewHumanReadableError(
			"Cannot modify ",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be provided"),
		)
	}
	if channelName != twitchUserName.(string) {
		return misc.NewHumanReadableError(
			"Cannot modify other user's room",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be provided"),
		)
	}
	return nil
}

func (s *ApiServer) HandleUpdateRoomSettings(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}

	decoder := json.NewDecoder(r.Body)
	var reqBody RoomUpdateRequest
	if err := decoder.Decode(&reqBody); err != nil {
		return misc.NewHumanReadableError(
			"Invalid request body",
			http.StatusBadRequest,
			fmt.Errorf("invalid request body: %w", err),
		)
	}

	channelName := reqBody.ChannelName
	if len(channelName) == 0 {
		return misc.NewHumanReadableError(
			"Channel name must be provided",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be provided"),
		)
	}
	if err := s.matchRequestChannel(r, channelName); err != nil {
		return err
	}

	if _, ok := s.roomsManager.Rooms[channelName]; !ok {
		return fmt.Errorf("room %s does not exist", channelName)
	}
	room := s.roomsManager.Rooms[channelName]

	config, err := room.GetSpineRuntimeConfig(r.Context())
	if err != nil {
		return err
	}
	// repo := room.RoomRepositoryPsql{}
	// config, err := room.GetSpineRuntimeConfigById(context.Background(), roomObj.GetRoomId())
	if reqBody.MinAnimationSpeed > 0 {
		config.MinAnimationSpeed = reqBody.MinAnimationSpeed
	}
	if reqBody.MaxAnimationSpeed > 0 {
		config.MaxAnimationSpeed = reqBody.MaxAnimationSpeed
	}
	if reqBody.MinVelocity > 0 {
		config.MinMovementSpeed = reqBody.MinVelocity
	}
	if reqBody.MaxVelocity > 0 {
		config.MaxMovementSpeed = reqBody.MaxVelocity
	}
	if reqBody.MinSpriteScale > 0 {
		config.MinScaleSize = reqBody.MinSpriteScale
	}
	if reqBody.MaxSpriteScale > 0 {
		config.MaxScaleSize = reqBody.MaxSpriteScale
	}
	if reqBody.MaxSpritePixelSize > 0 {
		config.MaxSpritePixelSize = reqBody.MaxSpritePixelSize
	}

	if err := misc.ValidateSpineRuntimeConfig(config); err != nil {
		return misc.NewHumanReadableError(
			"Invalid configuration settings",
			http.StatusBadRequest,
			err,
		)
	}
	err = room.UpdateSpineRuntimeConfig(r.Context(), config)
	if err != nil {
		return err
	}

	return nil
}

func (s *ApiServer) HandleGetRoomSettings(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return nil
	}

	channelName := r.URL.Query().Get("channel_name")
	if len(channelName) == 0 {
		return misc.NewHumanReadableError(
			"Channel name must be provided",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be provided"),
		)
	}
	if err := s.matchRequestChannel(r, channelName); err != nil {
		return err
	}

	if _, ok := s.roomsManager.Rooms[channelName]; !ok {
		return fmt.Errorf("room %s does not exist", channelName)
	}
	room := s.roomsManager.Rooms[channelName]
	config, err := room.GetSpineRuntimeConfig(r.Context())
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(GetRoomSettingsResponse{
		MinAnimationSpeed:  config.MinAnimationSpeed,
		MaxAnimationSpeed:  config.MaxAnimationSpeed,
		MinMovementSpeed:   config.MinMovementSpeed,
		MaxMovementSpeed:   config.MaxMovementSpeed,
		MinSpriteSize:      config.MinScaleSize,
		MaxSpriteSize:      config.MaxScaleSize,
		MaxSpritePixelSize: config.MaxSpritePixelSize,
	})
}

func (s *ApiServer) HandleRoomAddOperator(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	var reqBody RoomAddOperatorRequest
	if err := decoder.Decode(&reqBody); err != nil {
		return misc.NewHumanReadableError(
			"Invalid request body",
			http.StatusBadRequest,
			fmt.Errorf("invalid request body: %w", err),
		)
	}

	channelName := reqBody.ChannelName
	if len(channelName) == 0 {
		return misc.NewHumanReadableError(
			"Channel name must be provided",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be provided"),
		)
	}
	if _, ok := s.roomsManager.Rooms[channelName]; !ok {
		return fmt.Errorf("room %s does not exist", channelName)
	}
	room := s.roomsManager.Rooms[channelName]

	// TODO: TwitchUserId is not real?
	room.GiveChibiToUser(context.Background(), misc.UserInfo{
		Username:        reqBody.Username,
		UsernameDisplay: reqBody.UserDisplayName,
		TwitchUserId:    "0",
	})
	// faction, err := spine.FactionEnum_Parse(reqBody.Faction)
	// if err != nil {
	// 	return err
	// }
	// room.AddOperatorToRoom(reqBody.Username, reqBody.Username, reqBody.OperatorId, faction)

	return nil
}

func (s *ApiServer) HandleAdminInfo(w http.ResponseWriter, r *http.Request) error {
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

func (s *ApiServer) HandleRemoveRoom(w http.ResponseWriter, r *http.Request) error {
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

func (s *ApiServer) HandleRemoveUser(w http.ResponseWriter, r *http.Request) error {
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
	return room.RemoveUserChibi(context.Background(), userName)
}
