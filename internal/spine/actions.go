package spine

import "github.com/Stymphalian/ak_chibi_bot/internal/misc"

type ActionEnum string

const (
	ACTION_PLAY_ANIMATION = ActionEnum("PLAY_ANIMATION")
	ACTION_WANDER         = ActionEnum("WANDER")
	ACTION_WALK_TO        = ActionEnum("WALK_TO")
	ACTION_NONE           = ActionEnum("")
)

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

func NewActionPlayAnimation(animations []string) (r ActionUnion) {
	r.Animations = animations
	return r
}
func NewActionWander(animation string) (r ActionUnion) {
	r.WanderAnimation = animation
	return r
}
func NewActionWalkTo(TargetPos misc.Vector2, walkAnimation string, finalAnimation string) (r ActionUnion) {
	r.TargetPos = misc.NewOption(TargetPos)
	r.WalkToAnimation = walkAnimation
	r.WalkToFinalAnimation = finalAnimation
	return r
}
