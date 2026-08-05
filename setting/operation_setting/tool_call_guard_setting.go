package operation_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// toolCallGuardBuiltinExemptGroup is always allowed to use tool calls,
// regardless of the configured allowed-groups list.
const toolCallGuardBuiltinExemptGroup = "admin"

type ToolCallGuardSetting struct {
	Enabled bool `json:"enabled"`
	// AllowedGroups is the whitelist: a request passes when the user group or
	// the billing group matches any entry (case-insensitive).
	AllowedGroups []string `json:"allowed_groups"`
	// TargetGroup is the group a user is moved into when an admin resolves
	// their tool_call ticket.
	TargetGroup string `json:"target_group"`
	// PromoteOnResolve controls whether resolving a tool_call ticket moves the
	// user into TargetGroup automatically.
	PromoteOnResolve bool `json:"promote_on_resolve"`
}

var toolCallGuardSetting = ToolCallGuardSetting{
	Enabled:          false,
	AllowedGroups:    []string{"tool"},
	TargetGroup:      "tool",
	PromoteOnResolve: true,
}

func init() {
	config.GlobalConfig.Register("tool_call_guard_setting", &toolCallGuardSetting)
}

func GetToolCallGuardSetting() *ToolCallGuardSetting {
	return &toolCallGuardSetting
}

func IsToolCallGuardEnabled() bool {
	return toolCallGuardSetting.Enabled
}

// IsToolCallAllowedForGroups reports whether a tool-call request from the
// given user group / billing group combination may pass the guard. The
// built-in "admin" group is always allowed so administrators can never lock
// themselves out via configuration.
func IsToolCallAllowedForGroups(userGroup, billingGroup string) bool {
	if strings.EqualFold(strings.TrimSpace(userGroup), toolCallGuardBuiltinExemptGroup) ||
		strings.EqualFold(strings.TrimSpace(billingGroup), toolCallGuardBuiltinExemptGroup) {
		return true
	}
	allowed := GetToolCallGuardSetting().AllowedGroups
	return containsStringIgnoreCase(allowed, userGroup) || containsStringIgnoreCase(allowed, billingGroup)
}
