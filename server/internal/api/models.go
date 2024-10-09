package api

import "github.com/Stymphalian/ak_chibi_bot/server/internal/operator"

type chatter struct {
	Username     string `json:"username"`
	Operator     string `json:"operator"`
	LastChatTime string `json:"last_chat_time"`
}

type roomInfo struct {
	ChannelName             string             `json:"channel_name"`
	CreatedAt               string             `json:"created_at"`
	LastTimeUsed            string             `json:"last_time_used"`
	Chatters                []*chatter         `json:"chatters"`
	NextGCTime              string             `json:"next_gc_time"`
	NumWebsocketConnections int                `json:"num_websocket_connections"`
	ConnectionAverageFps    map[string]float64 `json:"connection_average_fps"`
}

type AdminInfo struct {
	Rooms      []*roomInfo            `json:"rooms"`
	NextGCTime string                 `json:"next_gc_time"`
	Metrics    map[string]interface{} `json:"metrics"`
}

type removeRoomRequest struct {
	ChannelName string `json:"channel_name"`
}

type removeUserRequest struct {
	ChannelName string `json:"channel_name"`
	Username    string `json:"username"`
}

type RoomUpdateRequest struct {
	ChannelName       string  `json:"channel_name"`
	MinAnimationSpeed float64 `json:"min_animation_speed"`
	MaxAnimationSpeed float64 `json:"max_animation_speed"`
	// DefaultAnimationSpeed float64 `json:"default_animation_speed"`
	MinVelocity float64 `json:"min_movement_speed"`
	MaxVelocity float64 `json:"max_movement_speed"`
	// DefaultVelocity       float64 `json:"default_velocity"`
	MinSpriteScale float64 `json:"min_sprite_size"`
	MaxSpriteScale float64 `json:"max_sprite_size"`
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

type GetUserPreferencesRequest struct {
	UserId uint `json:"user_id"`
}
type GetUserPreferencesResponse struct {
	OperatorInfo operator.OperatorInfo `json:"operator_info"`
}
type UpdateUserPreferencesRequest struct {
	UserId       uint                  `json:"user_id"`
	OperatorInfo operator.OperatorInfo `json:"operator_info"`
}
type DeleteUserPreferencesRquest struct {
	UserId uint `json:"user_id"`
}
