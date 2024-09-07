package spine

import "github.com/Stymphalian/ak_chibi_bot/internal/misc"

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
	UserName        string       `json:"user_name"`         // chonkyking
	UserNameDisplay string       `json:"user_name_display"` // ChonkyKing
	Operator        OperatorInfo `json:"operator"`
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

func (f *FacingData) HasFacing(facing ChibiFacingEnum) bool {
	if _, ok := f.Facings[facing]; !ok {
		return false
	}
	animations := f.Facings[facing]
	return len(animations) != 0
}

type SkinData struct {
	Stances map[ChibiTypeEnum]FacingData `json:"stance"`
}

func (s *SkinData) HasChibiType(chibiType ChibiTypeEnum) bool {
	if _, ok := s.Stances[chibiType]; !ok {
		return false
	}
	faceData := s.Stances[chibiType]
	return len(faceData.Facings) != 0
}

type GetOperatorResponse struct {
	SpineResponse
	OperatorId   string              `json:"operator_id"` // char_002_amiya
	OperatorName string              `json:"operator_name"`
	Skins        map[string]SkinData `json:"skins"` // build_char_002_amiya
}

// RemoveOperator
type RemoveOperatorRequest struct {
	UserName string `json:"user_name"`
}
type RemoveOperatorResponse struct {
	SpineResponse
}

type OperatorInfo struct {
	DisplayName       string          `json:"display_name"`
	Faction           FactionEnum     `json:"faction"`
	OperatorId        string          `json:"operator_id"`
	Skin              string          `json:"skin"`
	ChibiType         ChibiTypeEnum   `json:"chibi_type"`
	Facing            ChibiFacingEnum `json:"facing"`
	CurrentAnimations []string        `json:"animation"`
	AnimationSpeed    float64         `json:"animation_speed"`

	TargetPos misc.Option[misc.Vector2] `json:"target_pos"`
	StartPos  misc.Option[misc.Vector2] `json:"start_pos"`

	Skins      []string `json:"skins"`
	Animations []string `json:"animations"`
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

	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	Close()
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
