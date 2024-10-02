package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
)

type ApiServer struct {
	roomsManager *room.RoomsManager
	authService  auth.AuthServiceInterface
}

type RoomUpdateRequest struct {
	ChannelName       string  `json:"channel_name"`
	MinAnimationSpeed float64 `json:"min_animation_speed"`
	MaxAnimationSpeed float64 `json:"max_animation_speed"`
	// DefaultAnimationSpeed float64 `json:"default_animation_speed"`
	MinVelocity float64 `json:"min_velocity"`
	MaxVelocity float64 `json:"max_velocity"`
	// DefaultVelocity       float64 `json:"default_velocity"`
	MinSpriteScale float64 `json:"min_sprite_scale"`
	MaxSpriteScale float64 `json:"max_sprite_scale"`
	// DefaultSpriteScale    float64 `json:"default_sprite_scale"`
	MaxSpritePixelSize int `json:"max_sprite_pixel_size"`
}

type RoomAddOperatorRequest struct {
	ChannelName     string `json:"channel_name"`
	Username        string `json:"username"`
	UserDisplayName string `json:"user_display_name"`
	// OperatorId      string `json:"operator_id"`
	// Faction         string `json:"faction"`
}

type GetRoomSettingsRequest struct {
	ChannelName string `json:"channel_name"`
}

type GetRoomSettingsResponse struct {
	MinAnimationSpeed  float64 `json:"min_animation_speed"`
	MaxAnimationSpeed  float64 `json:"max_animation_speed"`
	MinMovementSpeed   float64 `json:"min_movement_speed"`
	MaxMovementSpeed   float64 `json:"max_movement_speed"`
	MinSpriteSize      float64 `json:"min_sprite_size"`
	MaxSpriteSize      float64 `json:"max_sprite_size"`
	MaxSpritePixelSize int     `json:"max_sprite_pixel_size"`
}

func NewApiServer(roomManager *room.RoomsManager, authService auth.AuthServiceInterface) *ApiServer {
	return &ApiServer{
		roomsManager: roomManager,
		authService:  authService,
	}
}

func (s *ApiServer) CheckAuth(h misc.HandlerWithErr) misc.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) error {
		info, err := s.authService.IsAuthorized(w, r)
		if err != nil || !info.Authenticated {
			return misc.NewHumanReadableError(
				"not authorized",
				http.StatusUnauthorized,
				err,
			)
		}

		newContext := context.WithValue(r.Context(), auth.CONTEXT_TWITCH_USER_ID, info.TwitchUserId)
		newContext = context.WithValue(newContext, auth.CONTEXT_TWITCH_USER_NAME, info.Username)
		*r = *r.WithContext(newContext)
		return h(w, r)
	}
}

func (s *ApiServer) middleware(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(
		s.CheckAuth(h),
		5*time.Second,
	)
}

func (s *ApiServer) RegisterHandlers() {
	mux := http.NewServeMux()
	mux.Handle("GET  /api/rooms/settings/{$}", s.middleware(s.HandleGetRoomSettings))
	mux.Handle("POST /api/rooms/settings/{$}", s.middleware(s.HandleUpdateRoomSettings))
	http.Handle("/api/", mux)
	// http.Handle("POST /api/rooms/add_operator", s.middleware(s.HandleRoomAddOperator))
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
