package system_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// DiscordGuildConfig 单条服务器/身份组验证规则。
// 用户命中任意一条规则即可通过登录验证；
// 新用户注册要求命中的规则中至少有一条 AllowRegister=true。
type DiscordGuildConfig struct {
	// ServerId Discord 服务器(Guild) ID，必填
	ServerId string `json:"server_id"`
	// RoleIds 身份组 ID 列表，拥有其中任意一个即视为命中；为空表示仅要求加入服务器
	RoleIds []string `json:"role_ids"`
	// AllowRegister 该规则是否允许新用户注册（false 时老用户仍可登录）
	AllowRegister bool `json:"allow_register"`
}

type DiscordSettings struct {
	Enabled      bool   `json:"enabled"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	// Deprecated: 单服务器旧配置，仅在 GuildConfigs 为空时生效（兼容历史数据）
	ServerId string `json:"server_id"`
	// Deprecated: 单身份组旧配置，仅在 GuildConfigs 为空时生效（兼容历史数据）
	RoleId string `json:"role_id"`
	// GuildConfigs 多服务器/身份组验证规则列表
	GuildConfigs []DiscordGuildConfig `json:"guild_configs"`
}

var defaultDiscordSettings = DiscordSettings{}

func init() {
	config.GlobalConfig.Register("discord", &defaultDiscordSettings)
}

func GetDiscordSettings() *DiscordSettings {
	return &defaultDiscordSettings
}

// GetEffectiveGuildConfigs 返回生效的验证规则列表。
// 优先使用 GuildConfigs；为空时回退到旧的 ServerId/RoleId 单服务器配置（默认允许注册）。
// 返回空列表表示不做服务器校验。
func (s *DiscordSettings) GetEffectiveGuildConfigs() []DiscordGuildConfig {
	configs := make([]DiscordGuildConfig, 0, len(s.GuildConfigs))
	for _, gc := range s.GuildConfigs {
		gc.ServerId = strings.TrimSpace(gc.ServerId)
		if gc.ServerId == "" {
			continue
		}
		roleIds := make([]string, 0, len(gc.RoleIds))
		for _, rid := range gc.RoleIds {
			rid = strings.TrimSpace(rid)
			if rid != "" {
				roleIds = append(roleIds, rid)
			}
		}
		gc.RoleIds = roleIds
		configs = append(configs, gc)
	}
	if len(configs) > 0 {
		return configs
	}
	// 兼容旧的单服务器配置
	if serverId := strings.TrimSpace(s.ServerId); serverId != "" {
		legacy := DiscordGuildConfig{
			ServerId:      serverId,
			AllowRegister: true,
		}
		if roleId := strings.TrimSpace(s.RoleId); roleId != "" {
			legacy.RoleIds = []string{roleId}
		}
		return []DiscordGuildConfig{legacy}
	}
	return nil
}
