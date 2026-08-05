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

func TestHideTicketForUserKeepsDailyLimit(t *testing.T) {
	if DB == nil {
		t.Skip("database not initialized")
	}

	ticketSetting := operation_setting.GetTicketSetting()
	originalTicket := *ticketSetting
	t.Cleanup(func() { *ticketSetting = originalTicket })
	ticketSetting.Enabled = true
	ticketSetting.DailyLimit = 1

	username := "ticket_hide_" + common.GetRandomString(8)
	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, user.Insert(0))
	adminName := "ticket_hide_admin_" + common.GetRandomString(8)
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

	ticket, err := CreateTicket(user.Id, TicketCategoryOther, "hide me", "body")
	require.NoError(t, err)

	// An open ticket cannot be removed from the user's list.
	require.Error(t, HideTicketForUser(ticket.Id, user.Id))

	require.NoError(t, CloseTicket(ticket.Id, admin.Id, TicketCloseReasonResolved, ""))

	// A ticket belonging to somebody else is not visible to this user.
	require.Error(t, HideTicketForUser(ticket.Id, admin.Id))

	require.NoError(t, HideTicketForUser(ticket.Id, user.Id))

	// Hidden for the owner...
	tickets, total, err := GetUserTickets(user.Id, 1, 20)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, tickets)

	// ...but still counted against the daily limit, so hiding cannot be used to
	// create more tickets than allowed.
	_, err = CreateTicket(user.Id, TicketCategoryOther, "second", "body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "daily ticket limit")

	// ...and still visible to admins.
	adminTickets, adminTotal, err := GetAdminTickets(AdminTicketQuery{UserId: user.Id})
	require.NoError(t, err)
	assert.Equal(t, int64(1), adminTotal)
	require.Len(t, adminTickets, 1)
	assert.Equal(t, ticket.Id, adminTickets[0].Id)

	// Hiding twice is rejected, and the user can no longer reply to it.
	require.Error(t, HideTicketForUser(ticket.Id, user.Id))
	_, err = ReplyTicket(ticket.Id, user.Id, false, "still there?")
	require.Error(t, err)
}

func TestSetUserGroup(t *testing.T) {
	if DB == nil {
		t.Skip("database not initialized")
	}

	username := "group_set_" + common.GetRandomString(8)
	user := &User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, user.Insert(0))
	t.Cleanup(func() {
		_ = DB.Unscoped().Delete(&User{}, "id = ?", user.Id)
	})

	require.Error(t, SetUserGroup(user.Id, "  "))

	require.NoError(t, SetUserGroup(user.Id, "tool"))
	group, err := GetUserGroup(user.Id, true)
	require.NoError(t, err)
	assert.Equal(t, "tool", group)
}
