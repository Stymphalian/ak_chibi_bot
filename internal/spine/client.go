package spine

import (
	"io"

	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
)

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
	Stances map[ChibiStanceEnum]FacingData `json:"stance"`
}

func (s *SkinData) HasChibiStance(chibiStance ChibiStanceEnum) bool {
	if _, ok := s.Stances[chibiStance]; !ok {
		return false
	}
	faceData := s.Stances[chibiStance]
	return len(faceData.Facings) != 0
}

type GetOperatorResponse struct {
	SpineResponse
	OperatorId   string              `json:"operator_id"` // char_002_amiya
	OperatorName string              `json:"operator_name"`
	Skins        map[string]SkinData `json:"skins"` // build_char_002_amiya
}

func (r *GetOperatorResponse) GetSkinNames() []string {
	skins := make([]string, len(r.Skins))
	i := 0
	for skinName := range r.Skins {
		skins[i] = skinName
		i += 1
	}
	return skins
}

// RemoveOperator
type RemoveOperatorRequest struct {
	UserName string `json:"user_name"`
}
type RemoveOperatorResponse struct {
	SpineResponse
}

type OperatorInfo struct {
	OperatorDisplayName string                    `json:"operator_display_name"`
	Faction             FactionEnum               `json:"faction"`
	OperatorId          string                    `json:"operator_id"`
	Skin                string                    `json:"skin"`
	ChibiStance         ChibiStanceEnum           `json:"chibi_stance"`
	Facing              ChibiFacingEnum           `json:"facing"`
	AnimationSpeed      float64                   `json:"animation_speed"`
	Skins               []string                  `json:"skins"`
	AvailableAnimations []string                  `json:"available_animations"`
	StartPos            misc.Option[misc.Vector2] `json:"start_pos"`

	CurrentAction ActionEnum  `json:"current_action"`
	Action        ActionUnion `json:"action"`
}

func NewOperatorInfo(
	OperatorDisplayName string,
	Faction FactionEnum,
	OperatorId string,
	Skin string,
	ChibiStance ChibiStanceEnum,
	Facing ChibiFacingEnum,
	AvailableSkins []string,
	AvailableAnimations []string,
	AnimationSpeed float64,
	StartPos misc.Option[misc.Vector2],
	CurrentAction ActionEnum,
	Action ActionUnion,
) OperatorInfo {
	return OperatorInfo{
		OperatorDisplayName: OperatorDisplayName,
		Faction:             Faction,
		OperatorId:          OperatorId,
		Skin:                Skin,
		ChibiStance:         ChibiStance,
		Facing:              Facing,
		AnimationSpeed:      AnimationSpeed,
		Skins:               AvailableSkins,
		AvailableAnimations: AvailableAnimations,
		StartPos:            StartPos,

		CurrentAction: CurrentAction,
		Action:        Action,
	}
}

func EmptyOperatorInfo() *OperatorInfo {
	return &OperatorInfo{
		Skins:               make([]string, 0),
		AvailableAnimations: make([]string, 0),
	}
}

type SpineClient interface {
	io.Closer
	SetOperator(r *SetOperatorRequest) (*SetOperatorResponse, error)
	GetOperator(r *GetOperatorRequest) (*GetOperatorResponse, error)
	RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error)

	// GetOperatorIds(faction FactionEnum) ([]string, error)
	GetOperatorIdFromName(name string, faction FactionEnum) (string, []string)
	CurrentInfo(userName string) (OperatorInfo, error)

	SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails)
	GetRandomOperator() (*OperatorInfo, error)
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
