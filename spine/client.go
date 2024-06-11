package spine

const (
	SET_OPERATOR    = "SET_OPERATOR"
	REMOVE_OPERATOR = "REMOVE_OPERATOR"
	UPDATE_OPERATOR = "UPDATE_OPERATOR"
)

type SpineResponse struct {
	TypeName   string `json:"type_name"`
	ErrorMsg   string `json:"error_msg"`
	StatusCode int    `json:"staus_code"`
}

// SetOperator
type SetOperatorRequest struct {
	UserName        string          `json:"user_name"`         // chonkyking
	UserNameDisplay string          `json:"user_name_display"` // ChonkyKing
	OperatorId      string          `json:"operator_id"`       // char_002_amiya
	Faction         FactionEnum     `json:"faction"`           // operator or enemy
	Skin            string          `json:"skin"`              // build_char_002_amiya
	ChibiType       ChibiTypeEnum   `json:"chibi_type"`        // base
	Facing          ChibiFacingEnum `json:"facing"`            // Front
	Animation       string          `json:"animation"`         // Relax
}
type SetOperatorResponse struct {
	SpineResponse
}

// GetOperator
type GetOperatorRequest struct {
	OperatorId string      `json:"operator_id"` // char_002_amiya
	Faction    FactionEnum `json:"faction"`     // operator or enemy
}
type AnimationsList []string
type FacingData struct {
	Facings map[ChibiFacingEnum]AnimationsList `json:"facing"`
}
type SkinData struct {
	Stances map[ChibiTypeEnum]FacingData `json:"stance"`
}
type GetOperatorResponse struct {
	SpineResponse
	OperatorId string              `json:"operator_id"` // char_002_amiya
	Skins      map[string]SkinData `json:"skins"`       // build_char_002_amiya
}

// RemoveOperator
type RemoveOperatorRequest struct {
	UserName string `json:"user_name"`
}
type RemoveOperatorResponse struct {
	SpineResponse
}

type OperatorInfo struct {
	Name       string
	Faction    FactionEnum
	OperatorId string
	Skin       string
	ChibiType  ChibiTypeEnum
	Facing     ChibiFacingEnum
	Animation  string

	Skins      []string
	Animations []string
}

func EmptyOperatorInfo() *OperatorInfo {
	return &OperatorInfo{
		Skins:      make([]string, 0),
		Animations: make([]string, 0),
	}
}

type SpineClient interface {
	SetOperator(r *SetOperatorRequest) (*SetOperatorResponse, error)
	GetOperator(r *GetOperatorRequest) (*GetOperatorResponse, error)
	RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error)

	GetOperatorIds(faction FactionEnum) ([]string, error)
	GetOperatorIdFromName(name string, faction FactionEnum) (string, []string)
	CurrentInfo(userName string) (OperatorInfo, error)
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
