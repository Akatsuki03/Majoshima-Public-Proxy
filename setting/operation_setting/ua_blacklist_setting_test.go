package operation_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchUABlacklistPattern(t *testing.T) {
	original := *GetUABlacklistSetting()
	t.Cleanup(func() {
		*GetUABlacklistSetting() = original
	})

	setting := GetUABlacklistSetting()
	setting.Enabled = true
	setting.Patterns = []string{"claude-cli", "codex-cli"}

	assert.Equal(t, "claude-cli", MatchUABlacklistPattern("claude-cli/1.0"))
	assert.Equal(t, "codex-cli", MatchUABlacklistPattern("Mozilla Codex-CLI xyz"))
	assert.Equal(t, "", MatchUABlacklistPattern("curl/8.0"))

	setting.Enabled = false
	assert.Equal(t, "", MatchUABlacklistPattern("claude-cli/1.0"))
}

func TestUABlacklistExemptGroups(t *testing.T) {
	original := *GetUABlacklistSetting()
	t.Cleanup(func() {
		*GetUABlacklistSetting() = original
	})

	setting := GetUABlacklistSetting()
	setting.ExemptUserGroups = []string{"vip", "staff"}
	setting.ExemptBillingGroups = []string{"special"}

	assert.True(t, IsUABlacklistExemptUserGroup("VIP"))
	assert.False(t, IsUABlacklistExemptUserGroup("default"))
	assert.True(t, IsUABlacklistExemptBillingGroup("special"))
	assert.False(t, IsUABlacklistExemptBillingGroup("default"))
}
