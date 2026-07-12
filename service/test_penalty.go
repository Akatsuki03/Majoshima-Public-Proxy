package service

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// IsTestInputBelowThreshold reports whether the estimated input token count is
// below the configured minimum, which is treated as a test/probe message.
func IsTestInputBelowThreshold(tokens int) bool {
	settings := operation_setting.GetTestPenaltySetting()
	if !settings.Enabled || settings.MinInputTokens <= 0 {
		return false
	}
	return tokens < settings.MinInputTokens
}

// ChargeTestPenalty deducts the configured penalty from the user and returns a
// non-retryable error describing the rejection.
func ChargeTestPenalty(c *gin.Context, userId int, tokenId int, tokenName string, modelName string, tokens int) *types.NewAPIError {
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
		Content:   fmt.Sprintf("Test penalty: $%.2f deducted, input tokens (%d) below minimum threshold (%d)", settings.Amount, tokens, settings.MinInputTokens),
		TokenId:   tokenId,
	})

	logger.LogWarn(c, fmt.Sprintf("test penalty charged: user=%d, amount=$%.2f, quota=%d, input_tokens=%d, threshold=%d", userId, settings.Amount, penaltyQuota, tokens, settings.MinInputTokens))

	return types.NewErrorWithStatusCode(
		fmt.Errorf("input tokens (%d) below minimum threshold (%d), test message detected, penalty of $%.2f has been charged", tokens, settings.MinInputTokens, settings.Amount),
		types.ErrorCodeInvalidRequest,
		403,
		types.ErrOptionWithSkipRetry(),
	)
}
