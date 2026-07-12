package service

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

var testPatterns = []string{
	"just say test",
	"just say \"test\"",
	"simply reply with \"test\"",
	"reply with the word \"test\"",
	"say the word \"test\"",
	"respond with \"test\"",
	"just respond with test",
}

func IsSillyTavernTestMessage(combineText string) bool {
	lower := strings.ToLower(strings.TrimSpace(combineText))
	for _, pattern := range testPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func ChargeTestPenalty(c *gin.Context, userId int, tokenId int, tokenName string, modelName string) *types.NewAPIError {
	settings := operation_setting.GetTestPenaltySetting()
	if !settings.Enabled || settings.Amount <= 0 {
		return nil
	}

	penaltyQuota := decimal.NewFromFloat(settings.Amount).
		Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
		Round(0).
		IntPart()
	if penaltyQuota <= 0 {
		return nil
	}

	err := model.DecreaseUserQuota(userId, int(penaltyQuota), false)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("failed to charge test penalty: %s", err.Error()))
	}

	model.RecordConsumeLog(c, userId, model.RecordConsumeLogParams{
		ModelName: modelName,
		TokenName: tokenName,
		Quota:     int(penaltyQuota),
		Content:   fmt.Sprintf("Test penalty: $%.2f deducted for sending test message", settings.Amount),
		TokenId:   tokenId,
	})

	logger.LogWarn(c, fmt.Sprintf("test penalty charged: user=%d, amount=$%.2f, quota=%d", userId, settings.Amount, penaltyQuota))

	return types.NewErrorWithStatusCode(
		fmt.Errorf("test message detected, penalty of $%.2f has been charged", settings.Amount),
		types.ErrorCodeInvalidRequest,
		403,
		types.ErrOptionWithSkipRetry(),
	)
}
