package spine

import (
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type OperatorInfo struct {
	OperatorDisplayName string                    `json:"operator_display_name"`
	Faction             FactionEnum               `json:"faction"`
	OperatorId          string                    `json:"operator_id"`
	Skin                string                    `json:"skin"`
	ChibiStance         ChibiStanceEnum           `json:"chibi_stance"`
	Facing              ChibiFacingEnum           `json:"facing"`
	AnimationSpeed      float64                   `json:"animation_speed"`
	SpriteScale         misc.Option[misc.Vector2] `json:"sprite_scale"`
	Skins               []string                  `json:"skins"`
	AvailableAnimations []string                  `json:"available_animations"`
	StartPos            misc.Option[misc.Vector2] `json:"start_pos"`
	MovementSpeed       misc.Option[misc.Vector2] `json:"movement_speed"`

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
		SpriteScale:         misc.EmptyOption[misc.Vector2](),
		Skins:               AvailableSkins,
		AvailableAnimations: AvailableAnimations,
		StartPos:            StartPos,
		MovementSpeed:       misc.EmptyOption[misc.Vector2](),

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

type ChatUser struct {
	UserName        string
	UserNameDisplay string
	CurrentOperator OperatorInfo
	LastChatTime    time.Time
}

func NewChatUser(
	UserName string,
	UserNameDisplay string,
	LastChatTime time.Time,
) *ChatUser {
	return &ChatUser{
		UserName:        UserName,
		UserNameDisplay: UserNameDisplay,
		LastChatTime:    LastChatTime,
	}
}

func (c *ChatUser) IsActive(period time.Duration) bool {
	return misc.Clock.Since(c.LastChatTime) < period
}
