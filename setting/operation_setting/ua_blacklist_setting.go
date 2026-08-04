package operation_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

type UABlacklistSetting struct {
	Enabled              bool     `json:"enabled"`
	Patterns             []string `json:"patterns"`
	ExemptUserGroups     []string `json:"exempt_user_groups"`
	ExemptBillingGroups  []string `json:"exempt_billing_groups"`
}

var uaBlacklistSetting = UABlacklistSetting{
	Enabled: true,
	Patterns: []string{
		"claude-cli",
		"codex-cli",
	},
	ExemptUserGroups:    []string{},
	ExemptBillingGroups: []string{},
}

func init() {
	config.GlobalConfig.Register("ua_blacklist_setting", &uaBlacklistSetting)
}

func GetUABlacklistSetting() *UABlacklistSetting {
	return &uaBlacklistSetting
}

func IsUABlacklistEnabled() bool {
	return uaBlacklistSetting.Enabled
}

// MatchUABlacklistPattern returns the first case-insensitive substring pattern
// matched by userAgent, or empty string if none match.
func MatchUABlacklistPattern(userAgent string) string {
	setting := GetUABlacklistSetting()
	if !setting.Enabled || userAgent == "" {
		return ""
	}
	uaLower := strings.ToLower(userAgent)
	for _, pattern := range setting.Patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.Contains(uaLower, strings.ToLower(pattern)) {
			return pattern
		}
	}
	return ""
}

func IsUABlacklistExemptUserGroup(userGroup string) bool {
	return containsStringIgnoreCase(GetUABlacklistSetting().ExemptUserGroups, userGroup)
}

func IsUABlacklistExemptBillingGroup(billingGroup string) bool {
	return containsStringIgnoreCase(GetUABlacklistSetting().ExemptBillingGroups, billingGroup)
}

func containsStringIgnoreCase(list []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), target) {
			return true
		}
	}
	return false
}
