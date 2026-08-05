package controller

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloseToolCallTicketResolvedPromotesUserGroup covers the full admin flow:
// closing a tool_call ticket as resolved must move the ticket owner into the
// configured target group.
func TestCloseToolCallTicketResolvedPromotesUserGroup(t *testing.T) {
	db := openTokenControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Ticket{}, &model.TicketMessage{}, &model.Log{}))

	guard := operation_setting.GetToolCallGuardSetting()
	originalGuard := *guard
	t.Cleanup(func() { *guard = originalGuard })
	guard.PromoteOnResolve = true
	guard.TargetGroup = "vip" // present in default group ratio

	ticketSetting := operation_setting.GetTicketSetting()
	originalTicket := *ticketSetting
	t.Cleanup(func() { *ticketSetting = originalTicket })
	ticketSetting.Enabled = true
	ticketSetting.DailyLimit = 10

	require.True(t, ratio_setting.ContainsGroupRatio("vip"))

	username := "tc_promote_" + common.GetRandomString(8)
	user := &model.User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	require.NoError(t, user.Insert(0))
	adminName := "tc_promote_admin_" + common.GetRandomString(8)
	admin := &model.User{
		Username:    adminName,
		Password:    "password123",
		DisplayName: adminName,
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, admin.Insert(0))
	t.Cleanup(func() {
		_ = model.DB.Unscoped().Delete(&model.TicketMessage{}, "user_id in ?", []int{user.Id, admin.Id})
		_ = model.DB.Unscoped().Delete(&model.Ticket{}, "user_id = ?", user.Id)
		_ = model.DB.Unscoped().Delete(&model.User{}, "id in ?", []int{user.Id, admin.Id})
	})

	ticket, err := model.CreateTicket(user.Id, model.TicketCategoryToolCall, "need tools", "please enable")
	require.NoError(t, err)

	ctx, recorder := newAuthenticatedContext(t, http.MethodPost, "/api/ticket/"+strconv.Itoa(ticket.Id)+"/close",
		map[string]any{"close_reason": model.TicketCloseReasonResolved}, admin.Id)
	ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: strconv.Itoa(ticket.Id)})
	CloseAdminTicket(ctx)

	response := decodeAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)

	reloadedUser, err := model.GetUserById(user.Id, false)
	require.NoError(t, err)
	assert.Equal(t, "vip", reloadedUser.Group)

	reloaded, err := model.GetTicketById(ticket.Id)
	require.NoError(t, err)
	assert.Equal(t, model.TicketStatusClosed, reloaded.Status)
}
