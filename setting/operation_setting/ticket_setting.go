package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type TicketSetting struct {
	Enabled    bool `json:"enabled"`
	DailyLimit int  `json:"daily_limit"`
}

var ticketSetting = TicketSetting{
	Enabled:    true,
	DailyLimit: 1,
}

func init() {
	config.GlobalConfig.Register("ticket_setting", &ticketSetting)
}

func GetTicketSetting() *TicketSetting {
	return &ticketSetting
}

func IsTicketEnabled() bool {
	return ticketSetting.Enabled
}
