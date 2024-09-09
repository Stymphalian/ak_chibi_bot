package spine

import (
	"slices"
	"strings"
)

func FilterAnimations(availableAnimations []string) []string {
	animations := make([]string, 0)
	excludeAnimations := []string{
		"Default",
		"Start",
	}
	for _, animationName := range availableAnimations {
		if slices.Contains(excludeAnimations, animationName) {
			continue
		}
		if strings.Contains(animationName, "Default") {
			continue
		}
		if strings.HasSuffix(animationName, "_Begin") {
			continue
		}
		if strings.HasSuffix(animationName, "_End") {
			continue
		}
		animations = append(animations, animationName)
	}
	return animations
}
