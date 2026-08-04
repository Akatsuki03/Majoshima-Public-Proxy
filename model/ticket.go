package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

const (
	TicketCategorySpecialUsage = "special_usage"
	TicketCategoryToolCall     = "tool_call"
	TicketCategoryBug          = "bug"
	TicketCategoryOther        = "other"

	TicketStatusOpen   = "open"
	TicketStatusClosed = "closed"

	TicketCloseReasonResolved   = "resolved"
	TicketCloseReasonUnresolved = "unresolved"
	TicketCloseReasonInvalid    = "invalid"
	TicketCloseReasonCustom     = "custom"
)

var validTicketCategories = map[string]bool{
	TicketCategorySpecialUsage: true,
	TicketCategoryToolCall:     true,
	TicketCategoryBug:          true,
	TicketCategoryOther:        true,
}

var validTicketCloseReasons = map[string]bool{
	TicketCloseReasonResolved:   true,
	TicketCloseReasonUnresolved: true,
	TicketCloseReasonInvalid:    true,
	TicketCloseReasonCustom:     true,
}

type Ticket struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int    `json:"user_id" gorm:"not null;index"`
	Category    string `json:"category" gorm:"type:varchar(32);not null;index"`
	Title       string `json:"title" gorm:"type:varchar(200);not null"`
	Status      string `json:"status" gorm:"type:varchar(16);not null;default:'open';index"`
	CloseReason string `json:"close_reason" gorm:"type:varchar(32);default:''"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;autoCreateTime"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint;autoUpdateTime"`
	ClosedAt    int64  `json:"closed_at" gorm:"bigint;default:0"`

	Username string          `json:"username,omitempty" gorm:"-"`
	Messages []TicketMessage `json:"messages,omitempty" gorm:"-"`
}

func (Ticket) TableName() string {
	return "tickets"
}

type TicketMessage struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TicketId  int    `json:"ticket_id" gorm:"not null;index"`
	UserId    int    `json:"user_id" gorm:"not null;index"`
	IsAdmin   bool   `json:"is_admin"`
	Content   string `json:"content" gorm:"type:text;not null"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;autoCreateTime"`

	Username string `json:"username,omitempty" gorm:"-"`
}

func (TicketMessage) TableName() string {
	return "ticket_messages"
}

func IsValidTicketCategory(category string) bool {
	return validTicketCategories[category]
}

func IsValidTicketCloseReason(reason string) bool {
	return validTicketCloseReasons[reason]
}

func CountUserTicketsCreatedToday(userId int) (int64, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Unix()
	endOfDay := startOfDay + 86400
	var count int64
	err := DB.Model(&Ticket{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userId, startOfDay, endOfDay).
		Count(&count).Error
	return count, err
}

func CreateTicket(userId int, category, title, content string) (*Ticket, error) {
	setting := operation_setting.GetTicketSetting()
	if !setting.Enabled {
		return nil, errors.New("ticket feature is disabled")
	}
	if !IsValidTicketCategory(category) {
		return nil, errors.New("invalid ticket category")
	}
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if len(title) == 0 || len(title) > 200 {
		return nil, errors.New("title must be between 1 and 200 characters")
	}
	if len(content) == 0 {
		return nil, errors.New("content is required")
	}

	user, err := GetUserById(userId, false)
	if err != nil {
		return nil, err
	}
	if user.TicketDisabled {
		return nil, errors.New("ticket access is disabled for this account")
	}

	dailyLimit := setting.DailyLimit
	if dailyLimit <= 0 {
		dailyLimit = 1
	}
	count, err := CountUserTicketsCreatedToday(userId)
	if err != nil {
		return nil, err
	}
	if count >= int64(dailyLimit) {
		return nil, fmt.Errorf("daily ticket limit reached (%d)", dailyLimit)
	}

	now := time.Now().Unix()
	ticket := &Ticket{
		UserId:    userId,
		Category:  category,
		Title:     title,
		Status:    TicketStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	message := &TicketMessage{
		UserId:    userId,
		IsAdmin:   false,
		Content:   content,
		CreatedAt: now,
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		message.TicketId = ticket.Id
		return tx.Create(message).Error
	})
	if err != nil {
		return nil, err
	}
	ticket.Messages = []TicketMessage{*message}
	return ticket, nil
}

func GetTicketById(id int) (*Ticket, error) {
	var ticket Ticket
	err := DB.First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func GetTicketMessages(ticketId int) ([]TicketMessage, error) {
	var messages []TicketMessage
	err := DB.Where("ticket_id = ?", ticketId).Order("id asc").Find(&messages).Error
	return messages, err
}

func GetUserTickets(userId int, page, pageSize int) ([]*Ticket, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var total int64
	query := DB.Model(&Ticket{}).Where("user_id = ?", userId)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tickets []*Ticket
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tickets).Error
	return tickets, total, err
}

type AdminTicketQuery struct {
	Status   string
	Category string
	UserId   int
	Page     int
	PageSize int
}

func GetAdminTickets(q AdminTicketQuery) ([]*Ticket, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	query := DB.Model(&Ticket{})
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}
	if q.Category != "" {
		query = query.Where("category = ?", q.Category)
	}
	if q.UserId > 0 {
		query = query.Where("user_id = ?", q.UserId)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tickets []*Ticket
	err := query.Order("id desc").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}
	for _, ticket := range tickets {
		username, _ := GetUsernameById(ticket.UserId, false)
		ticket.Username = username
	}
	return tickets, total, nil
}

func ReplyTicket(ticketId, userId int, isAdmin bool, content string) (*TicketMessage, error) {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil, errors.New("content is required")
	}
	ticket, err := GetTicketById(ticketId)
	if err != nil {
		return nil, err
	}
	if ticket.Status != TicketStatusOpen {
		return nil, errors.New("ticket is closed")
	}
	if !isAdmin {
		if ticket.UserId != userId {
			return nil, errors.New("ticket not found")
		}
		user, err := GetUserById(userId, false)
		if err != nil {
			return nil, err
		}
		if user.TicketDisabled {
			return nil, errors.New("ticket access is disabled for this account")
		}
	}

	now := time.Now().Unix()
	message := &TicketMessage{
		TicketId:  ticketId,
		UserId:    userId,
		IsAdmin:   isAdmin,
		Content:   content,
		CreatedAt: now,
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		return tx.Model(&Ticket{}).Where("id = ?", ticketId).Updates(map[string]interface{}{
			"updated_at": now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return message, nil
}

func CloseTicket(ticketId, adminUserId int, closeReason, closeMessage string) error {
	if !IsValidTicketCloseReason(closeReason) {
		return errors.New("invalid close reason")
	}
	if closeReason == TicketCloseReasonCustom && len(closeMessage) == 0 {
		return errors.New("custom close message is required")
	}

	ticket, err := GetTicketById(ticketId)
	if err != nil {
		return err
	}
	if ticket.Status != TicketStatusOpen {
		return errors.New("ticket is already closed")
	}

	content := closeMessage
	if content == "" {
		switch closeReason {
		case TicketCloseReasonResolved:
			content = "Ticket closed as resolved."
		case TicketCloseReasonUnresolved:
			content = "Ticket closed as unresolved."
		case TicketCloseReasonInvalid:
			content = "Ticket closed as invalid."
		}
	}

	now := time.Now().Unix()
	message := &TicketMessage{
		TicketId:  ticketId,
		UserId:    adminUserId,
		IsAdmin:   true,
		Content:   content,
		CreatedAt: now,
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		return tx.Model(&Ticket{}).Where("id = ?", ticketId).Updates(map[string]interface{}{
			"status":       TicketStatusClosed,
			"close_reason": closeReason,
			"closed_at":    now,
			"updated_at":   now,
		}).Error
	})
}

// ZeroNegativeQuota clears a user's negative wallet quota once per cooldown period.
func ZeroNegativeQuota(userId int) error {
	setting := operation_setting.GetQuotaZeroSetting()
	if !setting.Enabled {
		return errors.New("quota zero feature is disabled")
	}
	cooldownDays := setting.CooldownDays
	if cooldownDays < 0 {
		cooldownDays = 0
	}

	var previousQuota int
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, userId).Error; err != nil {
			return err
		}
		if user.Quota >= 0 {
			return errors.New("quota is not negative")
		}
		now := time.Now().Unix()
		if user.LastQuotaZeroTime > 0 && cooldownDays > 0 {
			elapsed := now - user.LastQuotaZeroTime
			needed := int64(cooldownDays) * 86400
			if elapsed < needed {
				remaining := (needed - elapsed + 86399) / 86400
				return fmt.Errorf("quota zero is on cooldown (%d day(s) remaining)", remaining)
			}
		}
		previousQuota = user.Quota
		return tx.Model(&User{}).Where("id = ?", userId).Updates(map[string]interface{}{
			"quota":                0,
			"last_quota_zero_time": now,
		}).Error
	})
	if err != nil {
		return err
	}
	RecordLog(userId, LogTypeManage, fmt.Sprintf("User zeroed negative quota (was %d)", previousQuota))
	return InvalidateUserCache(userId)
}

// DisableUserForUABlacklist disables the user and records BanReason without touching Remark.
func DisableUserForUABlacklist(userId int, matchedPattern string) error {
	reason := fmt.Sprintf("UA blacklist: %s", matchedPattern)
	if len(reason) > 255 {
		reason = reason[:255]
	}
	err := DB.Model(&User{}).Where("id = ?", userId).Updates(map[string]interface{}{
		"status":     common.UserStatusDisabled,
		"ban_reason": reason,
	}).Error
	if err != nil {
		return err
	}
	_ = InvalidateUserCache(userId)
	_ = InvalidateUserTokensCache(userId)
	RecordLog(userId, LogTypeSystem, reason)
	return nil
}
