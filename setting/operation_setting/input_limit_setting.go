package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type InputLimitSetting struct {
	MaxInputTokens int `json:"max_input_tokens"`
}

var inputLimitSetting = InputLimitSetting{
	MaxInputTokens: 200000,
}

func init() {
	config.GlobalConfig.Register("input_limit_setting", &inputLimitSetting)
}

func GetInputLimitSetting() *InputLimitSetting {
	return &inputLimitSetting
}
