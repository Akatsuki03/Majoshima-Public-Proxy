package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type TestPenaltySetting struct {
	Enabled bool    `json:"enabled"`
	Amount  float64 `json:"amount"`
	// MinInputTokens: requests with estimated input tokens below this value are
	// treated as test messages and penalized. 0 disables the check.
	MinInputTokens int `json:"min_input_tokens"`
}

var testPenaltySetting = TestPenaltySetting{
	Enabled:        false,
	Amount:         1000,
	MinInputTokens: 0,
}

func init() {
	config.GlobalConfig.Register("test_penalty_setting", &testPenaltySetting)
}

func GetTestPenaltySetting() *TestPenaltySetting {
	return &testPenaltySetting
}
