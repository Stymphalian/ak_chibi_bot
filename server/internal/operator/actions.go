package operator

import (
	"slices"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type ActionEnum string

// Wander: Walk around and then periodically stop
// Walkd: Walk around randomly never stopping
// WalkTo: Walk to point A and then stop
// PaceAround: Pace between point A and B
const (
	ACTION_PLAY_ANIMATION = ActionEnum("PLAY_ANIMATION")
	ACTION_WANDER         = ActionEnum("WANDER")
	ACTION_WALK           = ActionEnum("WALK")
	ACTION_WALK_TO        = ActionEnum("WALK_TO")
	ACTION_PACE_AROUND    = ActionEnum("PACE_AROUND")
	ACTION_FOLLOW         = ActionEnum("FOLLOW")
	ACTION_NONE           = ActionEnum("")
)

func IsActionEnum(a ActionEnum) bool {
	return slices.Contains([]ActionEnum{
		ACTION_PLAY_ANIMATION,
		ACTION_WANDER,
		ACTION_WALK,
		ACTION_WALK_TO,
		ACTION_PACE_AROUND,
		ACTION_FOLLOW,
	}, a)
}

func IsWalkingAction(a ActionEnum) bool {
	return (a == ACTION_WALK ||
		a == ACTION_WANDER ||
		a == ACTION_WALK_TO ||
		a == ACTION_PACE_AROUND ||
		a == ACTION_FOLLOW)
}

type ActionPlayAnimation struct {
	Animations []string `json:"animations"`
}

type ActionWander struct {
	WanderAnimation     string `json:"wander_animation"`
	WanderAnimationIdle string `json:"wander_animation_idle"`
}

type ActionWalk struct {
	WalkAnimation string `json:"walk_animation"`
}

type ActionWalkTo struct {
	TargetPos            misc.Option[misc.Vector2] `json:"target_pos"`
	WalkToAnimation      string                    `json:"walk_to_animation"`
	WalkToFinalAnimation string                    `json:"walk_to_final_animation"`
}

type ActionPaceAround struct {
	PaceStartPos        misc.Option[misc.Vector2] `json:"pace_start_pos"`
	PaceEndPos          misc.Option[misc.Vector2] `json:"pace_end_pos"`
	PaceAroundAnimation string                    `json:"pace_around_animation"`
}

type ActionFollow struct {
	ActionFollowTarget        string `json:"action_follow_target"`
	ActionFollowWalkAnimation string `json:"action_follow_walk_animation"`
	ActionFollowIdleAnimation string `json:"action_follow_idle_animation"`
}

type ActionUnion struct {
	ActionPlayAnimation
	ActionWander
	ActionWalk
	ActionWalkTo
	ActionPaceAround
	ActionFollow
	IsSet         bool       `json:"is_set"`
	CurrentAction ActionEnum `json:"current_action"`
}

func (a *ActionUnion) GetAnimations(action ActionEnum) []string {
	switch action {
	case ACTION_PLAY_ANIMATION:
		return a.Animations
	case ACTION_WANDER:
		return []string{a.WanderAnimation, a.WanderAnimationIdle}
	case ACTION_WALK:
		return []string{a.WalkAnimation}
	case ACTION_WALK_TO:
		return []string{a.WalkToAnimation, a.WalkToFinalAnimation}
	case ACTION_PACE_AROUND:
		return []string{a.PaceAroundAnimation}
	case ACTION_FOLLOW:
		return []string{a.ActionFollowWalkAnimation}
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
	r.CurrentAction = ACTION_PLAY_ANIMATION
	return r
}

func NewActionWander(animation string, animationIdle string) (r ActionUnion) {
	r.WanderAnimation = animation
	r.WanderAnimationIdle = animationIdle
	r.IsSet = true
	r.CurrentAction = ACTION_WANDER
	return r
}

func NewActionWalk(animation string) (r ActionUnion) {
	r.WalkAnimation = animation
	r.IsSet = true
	r.CurrentAction = ACTION_WALK
	return r
}

func NewActionWalkTo(TargetPos misc.Vector2, walkAnimation string, finalAnimation string) (r ActionUnion) {
	r.TargetPos = misc.NewOption(TargetPos)
	r.WalkToAnimation = walkAnimation
	r.WalkToFinalAnimation = finalAnimation
	r.IsSet = true
	r.CurrentAction = ACTION_WALK_TO
	return r
}

func NewActionPaceAround(startPos misc.Vector2, endPos misc.Vector2, animation string) (r ActionUnion) {
	r.PaceStartPos = misc.NewOption(startPos)
	r.PaceEndPos = misc.NewOption(endPos)
	r.PaceAroundAnimation = animation
	r.IsSet = true
	r.CurrentAction = ACTION_PACE_AROUND
	return r
}

func NewActionFollow(usernameTarget string, walkAnimation string, idleAnimation string) (r ActionUnion) {
	r.ActionFollowTarget = usernameTarget
	r.ActionFollowWalkAnimation = walkAnimation
	r.ActionFollowIdleAnimation = idleAnimation
	r.IsSet = true
	r.CurrentAction = ACTION_FOLLOW
	return r
}
