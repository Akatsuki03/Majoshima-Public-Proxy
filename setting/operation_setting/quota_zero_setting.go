package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type QuotaZeroSetting struct {
	Enabled      bool `json:"enabled"`
	CooldownDays int  `json:"cooldown_days"`
}

var quotaZeroSetting = QuotaZeroSetting{
	Enabled:      true,
	CooldownDays: 7,
}

func init() {
	config.GlobalConfig.Register("quota_zero_setting", &quotaZeroSetting)
}

func GetQuotaZeroSetting() *QuotaZeroSetting {
	return &quotaZeroSetting
}

func IsQuotaZeroEnabled() bool {
	return quotaZeroSetting.Enabled
}
