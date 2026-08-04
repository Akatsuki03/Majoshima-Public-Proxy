package service

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldDisableChannel_PermissionErrorNoCreditBalance(t *testing.T) {
	prev := common.AutomaticDisableChannelEnabled
	common.AutomaticDisableChannelEnabled = true
	t.Cleanup(func() { common.AutomaticDisableChannelEnabled = prev })

	err := types.WithOpenAIError(types.OpenAIError{
		Message: "team 'wilsongroup-178ad2' has no credit balance",
		Type:    "permission_error",
		Param:   "",
		Code:    nil,
	}, http.StatusForbidden)

	require.NotNil(t, err)
	assert.True(t, ShouldDisableChannel(err),
		"403 permission_error with no credit balance must auto-disable")
}

func TestShouldDisableChannel_OpenAIAuthAndBillingTypes(t *testing.T) {
	prev := common.AutomaticDisableChannelEnabled
	common.AutomaticDisableChannelEnabled = true
	t.Cleanup(func() { common.AutomaticDisableChannelEnabled = prev })

	cases := []struct {
		name string
		typ  string
		code any
	}{
		{name: "insufficient_quota", typ: "insufficient_quota"},
		{name: "authentication_error", typ: "authentication_error"},
		{name: "permission_error", typ: "permission_error"},
		{name: "forbidden", typ: "forbidden"},
		{name: "invalid_api_key", typ: "upstream_error", code: "invalid_api_key"},
		{name: "account_deactivated", typ: "upstream_error", code: "account_deactivated"},
		{name: "billing_not_active", typ: "upstream_error", code: "billing_not_active"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.WithOpenAIError(types.OpenAIError{
				Message: "test",
				Type:    tc.typ,
				Code:    tc.code,
			}, http.StatusForbidden)
			assert.True(t, ShouldDisableChannel(err))
		})
	}
}

func TestShouldDisableChannel_DisabledWhenFeatureOff(t *testing.T) {
	prev := common.AutomaticDisableChannelEnabled
	common.AutomaticDisableChannelEnabled = false
	t.Cleanup(func() { common.AutomaticDisableChannelEnabled = prev })

	err := types.WithOpenAIError(types.OpenAIError{
		Message: "team has no credit balance",
		Type:    "permission_error",
	}, http.StatusForbidden)
	assert.False(t, ShouldDisableChannel(err))
}
