package spine

import (
	"slices"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type ActionEnum string

const (
	ACTION_PLAY_ANIMATION = ActionEnum("PLAY_ANIMATION")
	ACTION_WANDER         = ActionEnum("WANDER")
	ACTION_WALK_TO        = ActionEnum("WALK_TO")
	ACTION_NONE           = ActionEnum("")
)

func IsActionEnum(a ActionEnum) bool {
	return slices.Contains([]ActionEnum{
		ACTION_PLAY_ANIMATION,
		ACTION_WANDER,
		ACTION_WALK_TO,
	}, a)
}

func IsWalkingAction(a ActionEnum) bool {
	return a == ACTION_WANDER || a == ACTION_WALK_TO
}

type ActionPlayAnimation struct {
	Animations []string `json:"animations"`
}

type ActionWander struct {
	WanderAnimation string `json:"wander_animation"`
}

type ActionWalkTo struct {
	TargetPos            misc.Option[misc.Vector2] `json:"target_pos"`
	WalkToAnimation      string                    `json:"walk_to_animation"`
	WalkToFinalAnimation string                    `json:"walk_to_final_animation"`
}

type ActionUnion struct {
	ActionPlayAnimation
	ActionWander
	ActionWalkTo
	IsSet         bool
	currentAction ActionEnum
}

func (a *ActionUnion) GetAnimations(action ActionEnum) []string {
	switch action {
	case ACTION_PLAY_ANIMATION:
		return a.Animations
	case ACTION_WANDER:
		return []string{a.WanderAnimation}
	case ACTION_WALK_TO:
		return []string{a.WalkToAnimation, a.WalkToFinalAnimation}
	case ACTION_NONE:
		fallthrough
	default:
		return []string{}
	}
}

func GetAvailableMoveAnimations(availableAnimations []string) []string {
	moveAnims := make([]string, 0)
	for _, anim := range availableAnimations {
		if strings.Contains(strings.ToLower(anim), "move") {
			moveAnims = append(moveAnims, anim)
		}
	}
	return moveAnims
}

func NewActionPlayAnimation(animations []string) (r ActionUnion) {
	r.Animations = animations
	r.IsSet = true
	r.currentAction = ACTION_PLAY_ANIMATION
	return r
}
func NewActionWander(animation string) (r ActionUnion) {
	r.WanderAnimation = animation
	r.IsSet = true
	r.currentAction = ACTION_WANDER
	return r
}
func NewActionWalkTo(TargetPos misc.Vector2, walkAnimation string, finalAnimation string) (r ActionUnion) {
	r.TargetPos = misc.NewOption(TargetPos)
	r.WalkToAnimation = walkAnimation
	r.WalkToFinalAnimation = finalAnimation
	r.IsSet = true
	r.currentAction = ACTION_WALK_TO
	return r
}
