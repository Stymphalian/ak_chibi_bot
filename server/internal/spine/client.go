package spine

import (
	"net/http"
)

const (
	SET_OPERATOR    = "SET_OPERATOR"
	REMOVE_OPERATOR = "REMOVE_OPERATOR"
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

type SpineRuntime interface {
	Close() error
	AddConnection(w http.ResponseWriter, r *http.Request, chatters []*ChatUser) error
	NumConnections() int
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
