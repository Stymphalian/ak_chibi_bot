package spine

import "testing"

func TestValidateUpdateSetDefaultOtherwise(t *testing.T) {
	// Default to operator if faction not provided
	// Error if operatorId not found
	// if No skin exists then use "default" skin
	// If the chibi doesn't have a stance, try the other stance. Default to base
	// If doesn't have facing default to front
	// animationSpeed clampts to 0 and 6.0
	// startPos if present, must clamp position
	// If faction operator, chibi stance is changed to battle and
	// Action.playAnimation must be a valid animations
	// Action.wander must have valid animation
	// Action.WalkTo must have a valid animation
	// test size clamping
	// test movementSpeed clamping
	// test unassigned currentAction/action
	// regression test for enemy->walk->operator->battle causes the chibi to slide across screen
}
