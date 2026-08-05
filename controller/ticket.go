package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

type createTicketRequest struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type replyTicketRequest struct {
	Content string `json:"content"`
}

type closeTicketRequest struct {
	CloseReason  string `json:"close_reason"`
	CloseMessage string `json:"close_message"`
}

func GetSelfTickets(c *gin.Context) {
	userId := c.GetInt("id")
	page, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	tickets, total, err := model.GetUserTickets(userId, page, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user, _ := model.GetUserById(userId, false)
	ticketDisabled := false
	if user != nil {
		ticketDisabled = user.TicketDisabled
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":           tickets,
			"total":           total,
			"page":            page,
			"page_size":       pageSize,
			"ticket_disabled": ticketDisabled,
			"enabled":         operation_setting.IsTicketEnabled(),
			"daily_limit":     operation_setting.GetTicketSetting().DailyLimit,
		},
	})
}

func CreateSelfTicket(c *gin.Context) {
	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	userId := c.GetInt("id")
	ticket, err := model.CreateTicket(userId, strings.TrimSpace(req.Category), req.Title, req.Content)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    ticket,
	})
}

func GetSelfTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	userId := c.GetInt("id")
	ticket, err := model.GetTicketById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if ticket.UserId != userId || ticket.UserHidden {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "ticket not found"})
		return
	}
	messages, err := model.GetTicketMessages(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	ticket.Messages = messages
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    ticket,
	})
}

func ReplySelfTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req replyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	userId := c.GetInt("id")
	message, err := model.ReplyTicket(id, userId, false, strings.TrimSpace(req.Content))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    message,
	})
}

func DeleteSelfTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	userId := c.GetInt("id")
	if err := model.HideTicketForUser(id, userId); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func GetAdminTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	userId, _ := strconv.Atoi(c.Query("user_id"))
	tickets, total, err := model.GetAdminTickets(model.AdminTicketQuery{
		Status:   c.Query("status"),
		Category: c.Query("category"),
		UserId:   userId,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     tickets,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetAdminTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	ticket, err := model.GetTicketById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	messages, err := model.GetTicketMessages(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	for i := range messages {
		username, _ := model.GetUsernameById(messages[i].UserId, false)
		messages[i].Username = username
	}
	username, _ := model.GetUsernameById(ticket.UserId, false)
	ticket.Username = username
	ticket.Messages = messages
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    ticket,
	})
}

func ReplyAdminTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req replyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	adminId := c.GetInt("id")
	message, err := model.ReplyTicket(id, adminId, true, strings.TrimSpace(req.Content))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    message,
	})
}

func CloseAdminTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req closeTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	adminId := c.GetInt("id")
	closeReason := strings.TrimSpace(req.CloseReason)
	ticket, err := model.GetTicketById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	err = model.CloseTicket(id, adminId, closeReason, strings.TrimSpace(req.CloseMessage))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	promoteMessage := promoteUserGroupForResolvedToolCallTicket(c, ticket, closeReason)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": promoteMessage,
	})
}

// promoteUserGroupForResolvedToolCallTicket moves the ticket owner into the
// configured tool-call group after an admin resolves their tool_call ticket,
// so the tool-call guard stops blocking them. The returned message surfaces
// skip/failure reasons to the closing admin; the ticket close itself already
// succeeded, so failures never roll it back.
func promoteUserGroupForResolvedToolCallTicket(c *gin.Context, ticket *model.Ticket, closeReason string) string {
	setting := operation_setting.GetToolCallGuardSetting()
	targetGroup := strings.TrimSpace(setting.TargetGroup)
	if !setting.PromoteOnResolve || targetGroup == "" {
		return ""
	}
	if ticket.Category != model.TicketCategoryToolCall || closeReason != model.TicketCloseReasonResolved {
		return ""
	}
	if !ratio_setting.ContainsGroupRatio(targetGroup) {
		common.SysError(fmt.Sprintf("tool call ticket %d resolved but target group %q is not in group ratio, skip promoting user %d", ticket.Id, targetGroup, ticket.UserId))
		return fmt.Sprintf("ticket closed, but user was NOT moved to group %q: the group does not exist in group ratio settings", targetGroup)
	}
	owner, err := model.GetUserById(ticket.UserId, false)
	if err != nil {
		common.SysError(fmt.Sprintf("tool call ticket %d resolved but failed to load user %d: %s", ticket.Id, ticket.UserId, err.Error()))
		return fmt.Sprintf("ticket closed, but user group promotion failed: %s", err.Error())
	}
	currentGroup := owner.Group
	if strings.EqualFold(currentGroup, targetGroup) {
		return ""
	}
	if err := model.SetUserGroup(ticket.UserId, targetGroup); err != nil {
		common.SysError(fmt.Sprintf("tool call ticket %d resolved but failed to promote user %d to group %q: %s", ticket.Id, ticket.UserId, targetGroup, err.Error()))
		return fmt.Sprintf("ticket closed, but user group promotion failed: %s", err.Error())
	}
	recordManageAuditFor(c, ticket.UserId, "user.group_promote", map[string]interface{}{
		"ticket_id":  ticket.Id,
		"from_group": currentGroup,
		"to_group":   targetGroup,
	})
	return ""
}

func DeleteAdminTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.DeleteTicket(id); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func ZeroSelfQuota(c *gin.Context) {
	userId := c.GetInt("id")
	if err := model.ZeroNegativeQuota(userId); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func GetQuotaZeroStatus(c *gin.Context) {
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	setting := operation_setting.GetQuotaZeroSetting()
	cooldownDays := setting.CooldownDays
	var cooldownRemainingDays int64
	canZero := setting.Enabled && user.Quota < 0
	if canZero && user.LastQuotaZeroTime > 0 && cooldownDays > 0 {
		needed := int64(cooldownDays) * 86400
		elapsed := common.GetTimestamp() - user.LastQuotaZeroTime
		if elapsed < needed {
			canZero = false
			cooldownRemainingDays = (needed - elapsed + 86399) / 86400
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"enabled":                 setting.Enabled,
			"cooldown_days":           cooldownDays,
			"last_quota_zero_time":    user.LastQuotaZeroTime,
			"cooldown_remaining_days": cooldownRemainingDays,
			"can_zero":                canZero,
			"quota":                   user.Quota,
		},
	})
}
