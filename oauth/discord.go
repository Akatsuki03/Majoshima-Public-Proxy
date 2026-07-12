package oauth

import (
        "context"
        "encoding/json"
        "fmt"
        "net/http"
        "net/url"
        "strings"
        "time"

        "github.com/QuantumNous/new-api/common"
        "github.com/QuantumNous/new-api/i18n"
        "github.com/QuantumNous/new-api/logger"
        "github.com/QuantumNous/new-api/model"
        "github.com/QuantumNous/new-api/setting/system_setting"
        "github.com/gin-gonic/gin"
)

func init() {
        Register("discord", &DiscordProvider{})
}

type DiscordProvider struct{}

type discordOAuthResponse struct {
        AccessToken  string `json:"access_token"`
        IDToken      string `json:"id_token"`
        RefreshToken string `json:"refresh_token"`
        TokenType    string `json:"token_type"`
        ExpiresIn    int    `json:"expires_in"`
        Scope        string `json:"scope"`
}

type discordUser struct {
        UID  string `json:"id"`
        ID   string `json:"username"`
        Name string `json:"global_name"`
}

type discordGuildMember struct {
        Roles []string `json:"roles"`
        User  struct {
                ID string `json:"id"`
        } `json:"user"`
}

func (p *DiscordProvider) GetName() string {
        return "Discord"
}

func (p *DiscordProvider) IsEnabled() bool {
        return system_setting.GetDiscordSettings().Enabled
}

func (p *DiscordProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
        if code == "" {
                return nil, NewOAuthError(i18n.MsgOAuthInvalidCode, nil)
        }

        logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken: code=%s...", code[:min(len(code), 10)])

        settings := system_setting.GetDiscordSettings()
        scheme := "http"
        if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
                scheme = "https"
        }
        redirectUri := fmt.Sprintf("%s://%s/oauth/discord", scheme, c.Request.Host)
        values := url.Values{}
        values.Set("client_id", settings.ClientId)
        values.Set("client_secret", settings.ClientSecret)
        values.Set("code", code)
        values.Set("grant_type", "authorization_code")
        values.Set("redirect_uri", redirectUri)

        logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken: redirect_uri=%s", redirectUri)

        req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(values.Encode()))
        if err != nil {
                return nil, err
        }
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        req.Header.Set("Accept", "application/json")

        client := http.Client{
                Timeout: 5 * time.Second,
        }
        res, err := client.Do(req)
        if err != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] ExchangeToken error: %s", err.Error()))
                return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Discord"}, err.Error())
        }
        defer res.Body.Close()

        logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken response status: %d", res.StatusCode)

        var discordResponse discordOAuthResponse
        err = json.NewDecoder(res.Body).Decode(&discordResponse)
        if err != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] ExchangeToken decode error: %s", err.Error()))
                return nil, err
        }

        if discordResponse.AccessToken == "" {
                logger.LogError(ctx, "[OAuth-Discord] ExchangeToken failed: empty access token")
                return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "Discord"})
        }

        logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken success: scope=%s", discordResponse.Scope)

        return &OAuthToken{
                AccessToken:  discordResponse.AccessToken,
                TokenType:    discordResponse.TokenType,
                RefreshToken: discordResponse.RefreshToken,
                ExpiresIn:    discordResponse.ExpiresIn,
                Scope:        discordResponse.Scope,
                IDToken:      discordResponse.IDToken,
        }, nil
}

func (p *DiscordProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
        logger.LogDebug(ctx, "[OAuth-Discord] GetUserInfo: fetching user info")

        req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/v10/users/@me", nil)
        if err != nil {
                return nil, err
        }
        req.Header.Set("Authorization", "Bearer "+token.AccessToken)

        client := http.Client{
                Timeout: 5 * time.Second,
        }
        res, err := client.Do(req)
        if err != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] GetUserInfo error: %s", err.Error()))
                return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Discord"}, err.Error())
        }
        defer res.Body.Close()

        logger.LogDebug(ctx, "[OAuth-Discord] GetUserInfo response status: %d", res.StatusCode)

        if res.StatusCode != http.StatusOK {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] GetUserInfo failed: status=%d", res.StatusCode))
                return nil, NewOAuthError(i18n.MsgOAuthGetUserErr, nil)
        }

        var discordUser discordUser
        err = json.NewDecoder(res.Body).Decode(&discordUser)
        if err != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] GetUserInfo decode error: %s", err.Error()))
                return nil, err
        }

        if discordUser.UID == "" || discordUser.ID == "" {
                logger.LogError(ctx, "[OAuth-Discord] GetUserInfo failed: empty user fields")
                return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "Discord"})
        }

        logger.LogDebug(ctx, "[OAuth-Discord] GetUserInfo success: uid=%s, username=%s, name=%s", discordUser.UID, discordUser.ID, discordUser.Name)

        settings := system_setting.GetDiscordSettings()
        guildConfigs := settings.GetEffectiveGuildConfigs()
        registerAllowed := true
        if len(guildConfigs) > 0 {
                registerAllowed, err = p.verifyGuildConfigs(ctx, token.AccessToken, discordUser.UID, guildConfigs)
                if err != nil {
                        return nil, err
                }
        } else if strings.TrimSpace(settings.RoleId) != "" {
                logger.LogError(ctx, "[OAuth-Discord] role_id is set but server_id is empty, denying login for safety")
                return nil, NewOAuthError(i18n.MsgDiscordGuildCheckFailed, nil)
        }

        return &OAuthUser{
                ProviderUserID: discordUser.UID,
                Username:       discordUser.ID,
                DisplayName:    discordUser.Name,
                Extra: map[string]any{
                        "register_allowed": registerAllowed,
                },
        }, nil
}

// verifyGuildConfigs 依次检查各条服务器/身份组规则。
// 返回值 registerAllowed 表示命中的规则中是否至少有一条允许注册。
// 用户未命中任何规则时返回错误（登录也被拒绝）。
func (p *DiscordProvider) verifyGuildConfigs(ctx context.Context, accessToken string, userID string, configs []system_setting.DiscordGuildConfig) (bool, error) {
        matched := false
        registerAllowed := false
        sawMembership := false
        sawRoleMissing := false
        var lastCheckErr error

        for _, gc := range configs {
                member, err := p.fetchGuildMember(ctx, accessToken, userID, gc.ServerId)
                if err != nil {
                        if _, ok := err.(*guildNotMemberError); ok {
                                continue
                        }
                        // 其他错误（网络/权限/解析）记录后继续尝试下一条规则，
                        // 避免单个服务器配置错误导致所有用户无法登录
                        lastCheckErr = err
                        continue
                }
                sawMembership = true

                if len(gc.RoleIds) == 0 {
                        matched = true
                        if gc.AllowRegister {
                                registerAllowed = true
                        }
                        continue
                }

                hasRole := false
                for _, required := range gc.RoleIds {
                        for _, role := range member.Roles {
                                if role == required {
                                        hasRole = true
                                        break
                                }
                        }
                        if hasRole {
                                break
                        }
                }
                if hasRole {
                        matched = true
                        if gc.AllowRegister {
                                registerAllowed = true
                        }
                } else {
                        sawRoleMissing = true
                        logger.LogDebug(ctx, "[OAuth-Discord] User %s in guild %s but missing required roles %v", userID, gc.ServerId, gc.RoleIds)
                }
        }

        if matched {
                logger.LogDebug(ctx, "[OAuth-Discord] User %s passed guild verification, registerAllowed=%v", userID, registerAllowed)
                return registerAllowed, nil
        }

        if sawRoleMissing {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] User %s is in guild(s) but has no required role", userID))
                return false, NewOAuthError(i18n.MsgDiscordMissingRole, nil)
        }
        if sawMembership {
                // 理论上不会到这里（有成员资格且无角色要求时必然 matched）
                return false, NewOAuthError(i18n.MsgDiscordGuildCheckFailed, nil)
        }
        if lastCheckErr != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] All guild checks failed for user %s: %s", userID, lastCheckErr.Error()))
                return false, NewOAuthError(i18n.MsgDiscordGuildCheckFailed, nil)
        }
        logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] User %s is not a member of any configured guild", userID))
        return false, NewOAuthError(i18n.MsgDiscordNotGuildMember, nil)
}

// guildNotMemberError 表示用户不在指定服务器内（Discord 返回 404）
type guildNotMemberError struct{ guildId string }

func (e *guildNotMemberError) Error() string {
        return "not a member of guild " + e.guildId
}

// fetchGuildMember 获取用户在指定服务器的成员信息
func (p *DiscordProvider) fetchGuildMember(ctx context.Context, accessToken string, userID string, serverId string) (*discordGuildMember, error) {
        logger.LogDebug(ctx, "[OAuth-Discord] Checking guild membership: guild=%s, user=%s", serverId, userID)

        apiURL := fmt.Sprintf("https://discord.com/api/v10/users/@me/guilds/%s/member", serverId)
        req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
        if err != nil {
                return nil, err
        }
        req.Header.Set("Authorization", "Bearer "+accessToken)

        client := http.Client{
                Timeout: 5 * time.Second,
        }
        res, err := client.Do(req)
        if err != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild membership check error: guild=%s, err=%s", serverId, err.Error()))
                return nil, err
        }
        defer res.Body.Close()

        if res.StatusCode == http.StatusNotFound {
                logger.LogDebug(ctx, "[OAuth-Discord] User %s is not a member of guild %s", userID, serverId)
                return nil, &guildNotMemberError{guildId: serverId}
        }

        if res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusUnauthorized {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild check permission error: status=%d, guild=%s — check OAuth scope or bot config", res.StatusCode, serverId))
                return nil, fmt.Errorf("guild check permission error: status=%d", res.StatusCode)
        }

        if res.StatusCode != http.StatusOK {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild membership check failed: status=%d, guild=%s", res.StatusCode, serverId))
                return nil, fmt.Errorf("guild membership check failed: status=%d", res.StatusCode)
        }

        var member discordGuildMember
        err = common.DecodeJson(res.Body, &member)
        if err != nil {
                logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild member decode error: guild=%s, err=%s", serverId, err.Error()))
                return nil, err
        }

        logger.LogDebug(ctx, "[OAuth-Discord] User %s is a member of guild %s, roles: %v", userID, serverId, member.Roles)
        return &member, nil
}

func (p *DiscordProvider) IsUserIDTaken(providerUserID string) bool {
        return model.IsDiscordIdAlreadyTaken(providerUserID)
}

func (p *DiscordProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
        user.DiscordId = providerUserID
        return user.FillUserByDiscordId()
}

func (p *DiscordProvider) SetProviderUserID(user *model.User, providerUserID string) {
        user.DiscordId = providerUserID
}

func (p *DiscordProvider) GetProviderPrefix() string {
        return "discord_"
}
