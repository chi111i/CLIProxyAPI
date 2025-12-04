package util

import (
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
)

const (
	// DefaultThinkingBudget is the default thinking budget value used when enabling
	// thinking for models via Antigravity. This follows the Antigravity2api reference.
	DefaultThinkingBudget = 1024
)

// IsAntigravityThinkingModel determines if a model should have thinking enabled
// when used through the Antigravity provider. This follows the Antigravity2api
// reference implementation logic.
//
// Thinking is enabled for:
// - Models ending with "-thinking" (e.g., gemini-claude-sonnet-4-5-thinking)
// - gemini-2.5-pro and gemini-2.5-pro-image
// - Models starting with "gemini-3-pro-"
func IsAntigravityThinkingModel(modelName string) bool {
	return strings.HasSuffix(modelName, "-thinking") ||
		modelName == "gemini-2.5-pro" ||
		modelName == "gemini-2.5-pro-image" ||
		strings.HasPrefix(modelName, "gemini-3-pro-")
}

// IsAntigravityClaudeModel determines if a model is a Claude model in the
// Antigravity context. Claude models require special handling when thinking
// is enabled (e.g., removal of topP parameter).
func IsAntigravityClaudeModel(modelName string) bool {
	return strings.Contains(modelName, "claude")
}

// ModelSupportsThinking reports whether the given model has Thinking capability
// according to the model registry metadata (provider-agnostic).
func ModelSupportsThinking(model string) bool {
	if model == "" {
		return false
	}
	if info := registry.GetGlobalRegistry().GetModelInfo(model); info != nil {
		return info.Thinking != nil
	}
	return false
}

// NormalizeThinkingBudget clamps the requested thinking budget to the
// supported range for the specified model using registry metadata only.
// If the model is unknown or has no Thinking metadata, returns the original budget.
// For dynamic (-1), returns -1 if DynamicAllowed; otherwise approximates mid-range
// or min (0 if zero is allowed and mid <= 0).
func NormalizeThinkingBudget(model string, budget int) int {
	if budget == -1 { // dynamic
		if found, min, max, zeroAllowed, dynamicAllowed := thinkingRangeFromRegistry(model); found {
			if dynamicAllowed {
				return -1
			}
			mid := (min + max) / 2
			if mid <= 0 && zeroAllowed {
				return 0
			}
			if mid <= 0 {
				return min
			}
			return mid
		}
		return -1
	}
	if found, min, max, zeroAllowed, _ := thinkingRangeFromRegistry(model); found {
		if budget == 0 {
			if zeroAllowed {
				return 0
			}
			return min
		}
		if budget < min {
			return min
		}
		if budget > max {
			return max
		}
		return budget
	}
	return budget
}

// thinkingRangeFromRegistry attempts to read thinking ranges from the model registry.
func thinkingRangeFromRegistry(model string) (found bool, min int, max int, zeroAllowed bool, dynamicAllowed bool) {
	if model == "" {
		return false, 0, 0, false, false
	}
	info := registry.GetGlobalRegistry().GetModelInfo(model)
	if info == nil || info.Thinking == nil {
		return false, 0, 0, false, false
	}
	return true, info.Thinking.Min, info.Thinking.Max, info.Thinking.ZeroAllowed, info.Thinking.DynamicAllowed
}
