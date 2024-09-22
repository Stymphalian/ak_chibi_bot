package operator

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

const (
	DEFAULT_ANIM_BASE_RELAX = "Relax"
	DEFAULT_ANIM_BASE       = "Move"
	DEFAULT_ANIM_BATTLE     = "Idle"
	DEFAULT_SKIN_NAME       = "default"
	DEFAULT_MOVE_ANIM_NAME  = "Move"
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

func (oi *OperatorInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal OperatorInfo value:", value))
	}

	err := json.Unmarshal(bytes, oi)
	if err != nil {
		return err
	}
	return nil
}

func (oi OperatorInfo) Value() (driver.Value, error) {
	jsonData, err := json.Marshal(oi)
	return string(jsonData), err
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

type ChibiFacingEnum string

const (
	CHIBI_FACING_ENUM_FRONT ChibiFacingEnum = "Front"
	CHIBI_FACING_ENUM_BACK  ChibiFacingEnum = "Back"
)

func ChibiFacingEnum_Parse(str string) (ChibiFacingEnum, error) {
	switch strings.ToLower(str) {
	case "front":
		return CHIBI_FACING_ENUM_FRONT, nil
	case "back":
		return CHIBI_FACING_ENUM_BACK, nil
	default:
		return "", fmt.Errorf("invalid chibi facing (%s)", str)
	}
}

type ChibiStanceEnum string

const (
	CHIBI_STANCE_ENUM_BATTLE ChibiStanceEnum = "battle"
	CHIBI_STANCE_ENUM_BASE   ChibiStanceEnum = "base"
)

func ChibiStanceEnum_Parse(str string) (ChibiStanceEnum, error) {
	switch strings.ToLower(str) {
	case "battle":
		return CHIBI_STANCE_ENUM_BATTLE, nil
	case "base":
		return CHIBI_STANCE_ENUM_BASE, nil
	default:
		return "", fmt.Errorf("invalid chibi type (%s)", str)
	}
}

type FactionEnum string

const (
	FACTION_ENUM_OPERATOR FactionEnum = "operator"
	FACTION_ENUM_ENEMY    FactionEnum = "enemy"
)

func FactionEnum_Parse(str string) (FactionEnum, error) {
	switch strings.ToLower(str) {
	case "operator":
		return FACTION_ENUM_OPERATOR, nil
	case "enemy":
		return FACTION_ENUM_ENEMY, nil
	default:
		return "", fmt.Errorf("invalid faction type (%s)", str)
	}
}

func GetDefaultAnimForChibiStance(chibiStance ChibiStanceEnum) string {
	if chibiStance == CHIBI_STANCE_ENUM_BASE {
		return DEFAULT_ANIM_BASE
	} else {
		return DEFAULT_ANIM_BATTLE
	}
}
