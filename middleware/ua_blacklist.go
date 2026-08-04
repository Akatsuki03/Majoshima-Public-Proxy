package middleware

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// UABlacklistCheck disables accounts that match configured User-Agent patterns.
// Must run after TokenAuth so user id and using group are available.
func UABlacklistCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !operation_setting.IsUABlacklistEnabled() {
			c.Next()
			return
		}

		userId := c.GetInt("id")
		if userId == 0 {
			c.Next()
			return
		}

		if model.IsPrivilegedUser(userId) {
			c.Next()
			return
		}

		userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
		if operation_setting.IsUABlacklistExemptUserGroup(userGroup) {
			c.Next()
			return
		}

		billingGroup := common.GetContextKeyString(c, constant.ContextKeyUsingGroup)
		if operation_setting.IsUABlacklistExemptBillingGroup(billingGroup) {
			c.Next()
			return
		}

		userAgent := c.Request.UserAgent()
		matched := operation_setting.MatchUABlacklistPattern(userAgent)
		if matched == "" {
			c.Next()
			return
		}

		if err := model.DisableUserForUABlacklist(userId, matched); err != nil {
			common.SysError(fmt.Sprintf("UA blacklist disable user %d failed: %s", userId, err.Error()))
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "internal error")
			return
		}

		abortWithOpenAiMessage(c, http.StatusForbidden,
			fmt.Sprintf("Your account has been disabled due to disallowed client (matched: %s)", matched))
	}
}
