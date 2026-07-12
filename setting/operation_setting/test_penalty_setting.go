package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type TestPenaltySetting struct {
	Enabled bool    `json:"enabled"`
	Amount  float64 `json:"amount"`
}

var testPenaltySetting = TestPenaltySetting{
	Enabled: false,
	Amount:  1000,
}

func init() {
	config.GlobalConfig.Register("test_penalty_setting", &testPenaltySetting)
}

func GetTestPenaltySetting() *TestPenaltySetting {
	return &testPenaltySetting
}
