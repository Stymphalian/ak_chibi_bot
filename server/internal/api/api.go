package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
)

type ApiServer struct {
	roomsManager *room.RoomsManager
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

func NewApiServer(roomManager *room.RoomsManager) *ApiServer {
	return &ApiServer{
		roomsManager: roomManager,
	}
}

func (a *ApiServer) LoginAuth(h misc.HandlerWithErr) misc.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// TODO
		return h(w, r)
	}
}

func (s *ApiServer) middleware(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(s.LoginAuth(h), 5*time.Second)
}

func (s *ApiServer) RegisterHandlers() {
	http.Handle("POST /api/rooms/update", s.middleware(s.HandleRoomUpdate))
	// http.Handle("POST /api/rooms/add_operator", s.middleware(s.HandleRoomAddOperator))
}

func (s *ApiServer) HandleRoomUpdate(w http.ResponseWriter, r *http.Request) error {
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
	if _, ok := s.roomsManager.Rooms[channelName]; !ok {
		return fmt.Errorf("room %s does not exist", channelName)
	}
	room := s.roomsManager.Rooms[channelName]

	config, err := room.GetSpineRuntimeConfig(context.Background())
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
	err = room.UpdateSpineRuntimeConfig(context.Background(), config)
	if err != nil {
		return err
	}

	return nil
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

	room.GiveChibiToUser(context.Background(), reqBody.Username, reqBody.UserDisplayName)
	// faction, err := spine.FactionEnum_Parse(reqBody.Faction)
	// if err != nil {
	// 	return err
	// }
	// room.AddOperatorToRoom(reqBody.Username, reqBody.Username, reqBody.OperatorId, faction)

	return nil
}
