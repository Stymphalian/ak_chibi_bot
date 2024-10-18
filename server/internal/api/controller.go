package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
	"github.com/google/uuid"
)

type ApiServer struct {
	roomsManager     *room.RoomsManager
	authService      auth.AuthServiceInterface
	roomRepo         room.RoomRepository
	usersRepo        users.UserRepository
	userPrefsRepo    users.UserPreferencesRepository
	operatorsService *operator.OperatorService
}

// TODO: Refactor the API handlers. More of the logic should be moved into
// services instead of being handled directly in the handlers
func NewApiServer(
	roomManager *room.RoomsManager,
	authService auth.AuthServiceInterface,
	roomRepo room.RoomRepository,
	usersRepo users.UserRepository,
	userPrefsRepo users.UserPreferencesRepository,
	operatorService *operator.OperatorService,
) *ApiServer {
	log.Println("NewApiServer created")
	return &ApiServer{
		roomsManager:     roomManager,
		authService:      authService,
		roomRepo:         roomRepo,
		usersRepo:        usersRepo,
		userPrefsRepo:    userPrefsRepo,
		operatorsService: operatorService,
	}
}

func (s *ApiServer) RegisterHandlers(rootMux *http.ServeMux) {
	mux := http.NewServeMux()
	mux.Handle("GET  /api/rooms/settings/{$}", s.middleware(s.HandleGetRoomSettings))
	mux.Handle("POST /api/rooms/settings/{$}", s.middleware(s.HandleUpdateRoomSettings))
	mux.Handle("POST /api/rooms/remove/{$}", s.middlewareAdmin(s.HandleRemoveRoom))
	mux.Handle("POST /api/rooms/users/remove/{$}", s.middlewareAdmin(s.HandleRemoveUser))
	mux.Handle("POST /api/rooms/users/give/{$}", s.middlewareAdmin(s.HandleRoomGiveOperator))
	mux.Handle("POST /api/rooms/users/set/{$}", s.middlewareAdmin(s.HandleRoomSetOperator))

	mux.Handle("GET /api/users/preferences/{$}", s.middleware(s.HandleGetUserPreferences))
	mux.Handle("POST /api/users/preferences/{$}", s.middleware(s.HandleUpdateUserPreferences))
	mux.Handle("DELETE /api/users/preferences/{$}", s.middleware(s.HandleDeleteUserPreferences))
	mux.Handle("GET  /api/admin/info/{$}", s.middlewareAdmin(s.HandleAdminInfo))

	// mux.Handle("GET  /api/vul/get/{$}", s.middleware(s.HandleVulGet))
	// mux.Handle("POST /api/vul/post/{$}", s.middleware(s.HandleVulPost))
	// mux.Handle("GET  /api/vul/csrf/{$}", s.middleware(s.HandleVulCsrf))

	rootMux.Handle("/api/", mux)
}

func (s *ApiServer) CheckForAuthToken(h misc.HandlerWithErr, checkAdmin bool) misc.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) error {
		bearerToken := r.Header.Get("Authorization")
		if len(bearerToken) > 7 && strings.ToLower(bearerToken[0:6]) == "bearer" {
			bearerToken = bearerToken[7:]
		} else {
			return misc.NewHumanReadableError(
				"invalid authorization header",
				http.StatusBadRequest,
				fmt.Errorf("invalid authorization header"),
			)
		}

		claims, err := s.authService.ValidateJWTToken(bearerToken)
		if err != nil {
			return misc.NewHumanReadableError(
				"invalid token",
				http.StatusUnauthorized,
				err,
			)
		}
		userId := claims.UserId
		twitchUserId := claims.TwitchUserId
		twitchUserName := claims.TwitchUserName

		userDb, err := s.usersRepo.GetById(r.Context(), userId)
		if err != nil {
			return misc.NewHumanReadableError(
				"user does not exist",
				http.StatusUnauthorized,
				err,
			)
		}

		if checkAdmin {
			if userDb.UserRole.Valid && userDb.UserRole.String != "admin" {
				return misc.NewHumanReadableError(
					"not authorized",
					http.StatusUnauthorized,
					nil,
				)
			}
		}

		newContext := context.WithValue(
			r.Context(), auth.CONTEXT_TWITCH_USER_ID, twitchUserId)
		newContext = context.WithValue(
			newContext, auth.CONTEXT_TWITCH_USER_NAME, twitchUserName)
		newContext = context.WithValue(
			newContext, auth.CONTEXT_USER_ID, userId)
		newContext = context.WithValue(
			newContext, auth.CONTEXT_USER_ROLE, userDb.UserRole)
		*r = *r.WithContext(newContext)
		return h(w, r)
	}
}

func (s *ApiServer) middleware(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(
		s.CheckForAuthToken(h, false),
		5*time.Second,
	)
}

func (s *ApiServer) middlewareAdmin(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(
		s.CheckForAuthToken(h, true),
		5*time.Second,
	)
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
	if (misc.ValidateChannelName(channelName)) != nil {
		return misc.NewHumanReadableError(
			"Invalid channel name",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be alphanumeric and between 1 and 100 characters, was '%s'", channelName),
		)
	}
	if err := s.matchRequestChannel(r, channelName); err != nil {
		return err
	}

	return s.updateRoomSettings(r.Context(), channelName, reqBody)
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
	if (misc.ValidateChannelName(channelName)) != nil {
		return misc.NewHumanReadableError(
			"Invalid channel name",
			http.StatusBadRequest,
			fmt.Errorf("channel name must be alphanumeric and between 1 and 100 characters, was '%s'", channelName),
		)
	}
	if err := s.matchRequestChannel(r, channelName); err != nil {
		return err
	}
	resp, err := s.getRoomSettings(r.Context(), channelName)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}

func (s *ApiServer) HandleRoomGiveOperator(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	var reqBody RoomGiveOperatorRequest
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
	id := uuid.New().String()
	room.GiveChibiToUser(context.Background(), misc.UserInfo{
		Username:        reqBody.Username,
		UsernameDisplay: reqBody.UserDisplayName,
		TwitchUserId:    id,
	})

	return nil
}

func (s *ApiServer) HandleRoomSetOperator(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	var reqBody RoomSetOperatorRequest
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

	id := uuid.New().String()
	err := room.AddOperatorToRoom(
		r.Context(),
		misc.UserInfo{
			Username:        reqBody.Username,
			UsernameDisplay: reqBody.UserDisplayName,
			TwitchUserId:    id,
		},
		reqBody.OperatorId,
		reqBody.Faction,
		reqBody.Skin,
		reqBody.Stance,
		misc.Vector2{
			X: reqBody.PositionX,
			Y: reqBody.PositionY,
		},
	)
	return err
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

func (s *ApiServer) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return nil
	}

	userIdStr := r.FormValue("user_id")
	userId64, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		return fmt.Errorf("user_id must be a number: %w", err)
	}
	userId := uint(userId64)
	currentUserId := r.Context().Value(auth.CONTEXT_USER_ID).(uint)
	if currentUserId != userId {
		return fmt.Errorf("user_id must match the current user")
	}

	prefsDb, err := s.userPrefsRepo.GetByUserIdOrNil(r.Context(), userId)
	if err != nil {
		http.NotFound(w, r)
		return nil
	}
	if prefsDb == nil {
		http.NotFound(w, r)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetUserPreferencesResponse{
		OperatorInfo: prefsDb.OperatorInfo,
	})
	return nil
}

func (s *ApiServer) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return nil
	}

	decoder := json.NewDecoder(r.Body)
	var reqBody UpdateUserPreferencesRequest
	if err := decoder.Decode(&reqBody); err != nil {
		return misc.NewHumanReadableError(
			"Invalid request body",
			http.StatusBadRequest,
			fmt.Errorf("invalid request body: %w", err),
		)
	}
	userId := reqBody.UserId
	currentUserId := r.Context().Value(auth.CONTEXT_USER_ID).(uint)
	if currentUserId != userId {
		return fmt.Errorf("user_id must match the current user")
	}

	err := s.operatorsService.ValidateUpdateSetDefaultOtherwise(&(reqBody.OperatorInfo))
	if err != nil {
		return err
	}
	err = s.userPrefsRepo.SetByUserId(r.Context(), userId, &reqBody.OperatorInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *ApiServer) HandleDeleteUserPreferences(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		http.NotFound(w, r)
		return nil
	}

	decoder := json.NewDecoder(r.Body)
	var reqBody DeleteUserPreferencesRquest
	if err := decoder.Decode(&reqBody); err != nil {
		return misc.NewHumanReadableError(
			"Invalid request body",
			http.StatusBadRequest,
			fmt.Errorf("invalid request body: %w", err),
		)
	}
	userId := reqBody.UserId
	currentUserId := r.Context().Value(auth.CONTEXT_USER_ID).(uint)
	if currentUserId != userId {
		return fmt.Errorf("user_id must match the current user")
	}

	err := s.userPrefsRepo.DeleteByUserId(r.Context(), userId)
	return err
}

// INTERNAL
// ----------------------------------------------
// TODO: Move these methods into a separate service
func (s *ApiServer) updateRoomSettings(ctx context.Context, channelName string, reqBody RoomUpdateRequest) error {
	roomDb, err := s.roomRepo.GetRoomByChannelName(ctx, channelName)
	if err != nil {
		return misc.NewHumanReadableError(
			"Room not found",
			http.StatusNotFound,
			fmt.Errorf("room not found: %w", err),
		)
	}
	config := roomDb.SpineRuntimeConfig

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
	if err := misc.ValidateSpineRuntimeConfig(&config); err != nil {
		return misc.NewHumanReadableError(
			"Invalid configuration settings",
			http.StatusBadRequest,
			err,
		)
	}

	err = s.roomsManager.UpdateSpineRuntimeConfig(ctx, roomDb.RoomId, &config)
	return err
}

func (s *ApiServer) getRoomSettings(ctx context.Context, channelName string) (*GetRoomSettingsResponse, error) {
	roomDb, err := s.roomRepo.GetRoomByChannelName(ctx, channelName)
	if err != nil {
		return nil, misc.NewHumanReadableError(
			"Room not found",
			http.StatusNotFound,
			fmt.Errorf("room not found: %w", err),
		)
	}
	config := roomDb.SpineRuntimeConfig
	resp := &GetRoomSettingsResponse{
		MinAnimationSpeed:  config.MinAnimationSpeed,
		MaxAnimationSpeed:  config.MaxAnimationSpeed,
		MinMovementSpeed:   config.MinMovementSpeed,
		MaxMovementSpeed:   config.MaxMovementSpeed,
		MinSpriteSize:      config.MinScaleSize,
		MaxSpriteSize:      config.MaxScaleSize,
		MaxSpritePixelSize: config.MaxSpritePixelSize,
	}
	return resp, nil
}

func (s *ApiServer) HandleVulGet(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("hello"))
	log.Println("handle vul get")
	log.Println("r.Header", r.Header)
	log.Println("r.Referer", r.Referer())
	log.Println("r.UserAgent", r.UserAgent())
	// log.Println("csrf_token: ", csrf.Token(r))
	return nil
}

func (s *ApiServer) HandleVulPost(w http.ResponseWriter, r *http.Request) error {
	log.Println("handle vul post")
	log.Println("r.Header", r.Header)
	log.Println("r.Referer", r.Referer())
	log.Println("r.UserAgent", r.UserAgent())
	// log.Println("csrf_token: ", csrf.Token(r))
	return nil
}
