package operation_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsToolCallAllowedForGroups(t *testing.T) {
	original := *GetToolCallGuardSetting()
	t.Cleanup(func() {
		*GetToolCallGuardSetting() = original
	})

	setting := GetToolCallGuardSetting()
	setting.AllowedGroups = []string{"tool", "vip"}

	tests := []struct {
		name         string
		userGroup    string
		billingGroup string
		want         bool
	}{
		{"user group allowed", "tool", "default", true},
		{"billing group allowed", "default", "tool", true},
		{"case insensitive", "TOOL", "default", true},
		{"neither allowed", "default", "default", false},
		{"empty groups", "", "", false},
		{"admin user group always allowed", "admin", "default", true},
		{"admin billing group always allowed", "default", "Admin", true},
		{"admin allowed even with empty whitelist", "admin", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsToolCallAllowedForGroups(tc.userGroup, tc.billingGroup))
		})
	}

	// admin bypass must not depend on the configured whitelist
	setting.AllowedGroups = []string{}
	assert.True(t, IsToolCallAllowedForGroups("admin", "default"))
	assert.False(t, IsToolCallAllowedForGroups("tool", "default"))
}
