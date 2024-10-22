package spine

import (
	"net/http"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

const (
	// Request Type Strings
	SET_OPERATOR      = "SET_OPERATOR"
	REMOVE_OPERATOR   = "REMOVE_OPERATOR"
	SHOW_CHAT_MESSAGE = "SHOW_CHAT_MESSAGE"
	// Response Type Strings
	RUNTIME_DEBUG_UPDATE  = "RUNTIME_DEBUG_UPDATE"
	RUNTIME_ROOM_SETTINGS = "RUNTIME_ROOM_SETTINGS"
)

type BridgeRequest struct {
	TypeName string `json:"type_name"`
}

type BridgeResponse struct {
	TypeName   string `json:"type_name"`
	ErrorMsg   string `json:"error_msg"`
	StatusCode int    `json:"staus_code"`
}

// SetOperator
type SetOperatorRequest struct {
	UserName        string                `json:"user_name"`         // chonkyking
	UserNameDisplay string                `json:"user_name_display"` // ChonkyKing
	Operator        operator.OperatorInfo `json:"operator"`
}
type SetOperatorInternalRequest struct {
	BridgeRequest
	UserName            string                    `json:"user_name"`         // chonkyking
	UserNameDisplay     string                    `json:"user_name_display"` // ChonkyKing
	OperatorId          string                    `json:"operator_id"`
	AtlasFile           string                    `json:"atlas_file"`
	PngFile             string                    `json:"png_file"`
	SkelFile            string                    `json:"skel_file"`
	StartPos            misc.Option[misc.Vector2] `json:"start_pos"`
	AnimationSpeed      float64                   `json:"animation_speed"`
	AvailableAnimations []string                  `json:"available_animations"`
	SpriteScale         misc.Option[misc.Vector2] `json:"sprite_scale"`
	MaxSpritePixelSize  int                       `json:"max_sprite_pixel_size"`
	MovementSpeedPx     int                       `json:"movement_speed_px"`
	MovementSpeed       misc.Option[misc.Vector2] `json:"movement_speed"`
	Action              operator.ActionEnum       `json:"action"`
	ActionData          operator.ActionUnion      `json:"action_data"`
}
type SetOperatorResponse struct {
	BridgeResponse
}

// RemoveOperator
type RemoveOperatorRequest struct {
	UserName string `json:"user_name"`
}
type RemoveOperatorResponse struct {
	BridgeResponse
}

// ShowChatMessage
type ShowChatMessageRequest struct {
	UserName string `json:"user_name"` // chonkyking
	Message  string `json:"message"`
}
type ShowChatMessageInternalRequest struct {
	BridgeRequest
	UserName string `json:"user_name"` // chonkyking
	Message  string `json:"message"`
}
type ShowChatMessageResponse struct {
	BridgeResponse
}

// runtimeDebugUpdate
type RuntimeDebugUpdateRequest struct {
	TypeName   string  `json:"type_name"`
	AverageFps float64 `json:"average_fps"`
}

type RuntimeRoomSettingsRequest struct {
	BridgeRequest
	ShowChatMessages bool `json:"show_chat_messages"`
}

type ChatterInfo struct {
	Username        string
	UsernameDisplay string
	// TODO: Make this a pointer
	OperatorInfo operator.OperatorInfo
}

type ClientRequestCallback func(connId string, typeName string, message []byte)
type SpineRuntime interface {
	Close() error
	AddConnection(w http.ResponseWriter, r *http.Request, chatters []*ChatterInfo) error
	NumConnections() int

	// Add listeners for any incoming requests from the connected clients
	// Returns a function which can be used to remove the listener
	AddListenerToClientRequests(callback ClientRequestCallback) (func(), error)
}

type SpineClient interface {
	Close() error
	SetOperator(r *SetOperatorRequest) (*SetOperatorResponse, error)
	RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error)
	ShowChatMessage(r *ShowChatMessageRequest) (*ShowChatMessageResponse, error)
}

type UserNotFound struct {
	message string
}

func (e *UserNotFound) Error() string {
	return e.message
}

func NewUserNotFound(message string) error {
	return &UserNotFound{message: message}
}
