package spine

import (
	"net/http"
)

const (
	// Request Type Strings
	SET_OPERATOR    = "SET_OPERATOR"
	REMOVE_OPERATOR = "REMOVE_OPERATOR"
	// Response Type Strings
	RUNTIME_DEBUG_UPDATE = "RUNTIME_DEBUG_UPDATE"
)

type SpineResponse struct {
	TypeName   string `json:"type_name"`
	ErrorMsg   string `json:"error_msg"`
	StatusCode int    `json:"staus_code"`
}

// SetOperator
type SetOperatorRequest struct {
	UserName        string       `json:"user_name"`         // chonkyking
	UserNameDisplay string       `json:"user_name_display"` // ChonkyKing
	Operator        OperatorInfo `json:"operator"`
}
type SetOperatorResponse struct {
	SpineResponse
}

// RemoveOperator
type RemoveOperatorRequest struct {
	UserName string `json:"user_name"`
}
type RemoveOperatorResponse struct {
	SpineResponse
}

// runtimeDebugUpdate
type RuntimeDebugUpdateRequest struct {
	TypeName   string  `json:"type_name"`
	AverageFps float64 `json:"average_fps"`
}

type ClientRequestCallback func(connId string, typeName string, message []byte)
type SpineRuntime interface {
	Close() error
	AddConnection(w http.ResponseWriter, r *http.Request, chatters []*ChatUser) error
	NumConnections() int

	// Add listeners for any incoming requests from the connected clients
	// Returns a function which can be used to remove the listener
	AddListenerToClientRequests(callback ClientRequestCallback) (func(), error)
}

type SpineClient interface {
	Close() error
	SetOperator(r *SetOperatorRequest) (*SetOperatorResponse, error)
	RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error)
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
