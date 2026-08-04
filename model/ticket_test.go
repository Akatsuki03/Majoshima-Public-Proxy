package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidTicketCategoryAndCloseReason(t *testing.T) {
	assert.True(t, IsValidTicketCategory(TicketCategoryBug))
	assert.False(t, IsValidTicketCategory("unknown"))
	assert.True(t, IsValidTicketCloseReason(TicketCloseReasonResolved))
	assert.False(t, IsValidTicketCloseReason("done"))
}

func TestCreateTicketDailyLimit(t *testing.T) {
	if DB == nil {
		t.Skip("database not initialized")
	}

	ticketSetting := operation_setting.GetTicketSetting()
	originalTicket := *ticketSetting
	t.Cleanup(func() { *ticketSetting = originalTicket })
	ticketSetting.Enabled = true
	ticketSetting.DailyLimit = 1

	username := "ticket_user_" + common.GetRandomString(8)
	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, user.Insert(0))
	t.Cleanup(func() {
		_ = DB.Unscoped().Delete(&TicketMessage{}, "user_id = ?", user.Id)
		_ = DB.Unscoped().Delete(&Ticket{}, "user_id = ?", user.Id)
		_ = DB.Unscoped().Delete(&User{}, "id = ?", user.Id)
	})

	first, err := CreateTicket(user.Id, TicketCategoryBug, "title1", "body1")
	require.NoError(t, err)
	require.NotNil(t, first)

	_, err = CreateTicket(user.Id, TicketCategoryBug, "title2", "body2")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "daily ticket limit")

	msg, err := ReplyTicket(first.Id, user.Id, false, "follow-up")
	require.NoError(t, err)
	require.NotNil(t, msg)

	user.TicketDisabled = true
	require.NoError(t, DB.Model(user).Update("ticket_disabled", true).Error)
	_, err = ReplyTicket(first.Id, user.Id, false, "blocked")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestZeroNegativeQuotaCooldown(t *testing.T) {
	if DB == nil {
		t.Skip("database not initialized")
	}

	setting := operation_setting.GetQuotaZeroSetting()
	original := *setting
	t.Cleanup(func() { *setting = original })
	setting.Enabled = true
	setting.CooldownDays = 7

	username := "quota_zero_" + common.GetRandomString(8)
	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, user.Insert(0))
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("quota", -100).Error)
	t.Cleanup(func() {
		_ = DB.Unscoped().Delete(&User{}, "id = ?", user.Id)
	})

	require.NoError(t, ZeroNegativeQuota(user.Id))
	reloaded, err := GetUserById(user.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 0, reloaded.Quota)
	assert.Greater(t, reloaded.LastQuotaZeroTime, int64(0))

	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("quota", -50).Error)
	err = ZeroNegativeQuota(user.Id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cooldown")

	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]interface{}{
		"quota":                -50,
		"last_quota_zero_time": time.Now().Unix() - 8*86400,
	}).Error)
	require.NoError(t, ZeroNegativeQuota(user.Id))
}

func TestCloseTicketRequiresCustomMessage(t *testing.T) {
	if DB == nil {
		t.Skip("database not initialized")
	}

	ticketSetting := operation_setting.GetTicketSetting()
	originalTicket := *ticketSetting
	t.Cleanup(func() { *ticketSetting = originalTicket })
	ticketSetting.Enabled = true
	ticketSetting.DailyLimit = 10

	username := "ticket_close_" + common.GetRandomString(8)
	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, user.Insert(0))
	adminName := "ticket_admin_" + common.GetRandomString(8)
	admin := &User{
		Username:    adminName,
		Password:    "password123",
		DisplayName: adminName,
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, admin.Insert(0))
	t.Cleanup(func() {
		_ = DB.Unscoped().Delete(&TicketMessage{}, "user_id in ?", []int{user.Id, admin.Id})
		_ = DB.Unscoped().Delete(&Ticket{}, "user_id = ?", user.Id)
		_ = DB.Unscoped().Delete(&User{}, "id in ?", []int{user.Id, admin.Id})
	})

	ticket, err := CreateTicket(user.Id, TicketCategoryOther, "close me", "body")
	require.NoError(t, err)

	err = CloseTicket(ticket.Id, admin.Id, TicketCloseReasonCustom, "")
	require.Error(t, err)

	require.NoError(t, CloseTicket(ticket.Id, admin.Id, TicketCloseReasonResolved, ""))
	reloaded, err := GetTicketById(ticket.Id)
	require.NoError(t, err)
	assert.Equal(t, TicketStatusClosed, reloaded.Status)
	assert.Equal(t, TicketCloseReasonResolved, reloaded.CloseReason)
}

func TestDeleteTicketOnlyWhenClosed(t *testing.T) {
	if DB == nil {
		t.Skip("database not initialized")
	}

	ticketSetting := operation_setting.GetTicketSetting()
	originalTicket := *ticketSetting
	t.Cleanup(func() { *ticketSetting = originalTicket })
	ticketSetting.Enabled = true
	ticketSetting.DailyLimit = 10

	username := "ticket_delete_" + common.GetRandomString(8)
	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, user.Insert(0))
	adminName := "ticket_del_admin_" + common.GetRandomString(8)
	admin := &User{
		Username:    adminName,
		Password:    "password123",
		DisplayName: adminName,
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, admin.Insert(0))
	t.Cleanup(func() {
		_ = DB.Unscoped().Delete(&TicketMessage{}, "user_id in ?", []int{user.Id, admin.Id})
		_ = DB.Unscoped().Delete(&Ticket{}, "user_id = ?", user.Id)
		_ = DB.Unscoped().Delete(&User{}, "id in ?", []int{user.Id, admin.Id})
	})

	ticket, err := CreateTicket(user.Id, TicketCategoryOther, "delete me", "body")
	require.NoError(t, err)

	err = DeleteTicket(ticket.Id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only closed tickets")

	require.NoError(t, CloseTicket(ticket.Id, admin.Id, TicketCloseReasonResolved, ""))
	require.NoError(t, DeleteTicket(ticket.Id))

	_, err = GetTicketById(ticket.Id)
	require.Error(t, err)

	messages, err := GetTicketMessages(ticket.Id)
	require.NoError(t, err)
	assert.Empty(t, messages)
}
