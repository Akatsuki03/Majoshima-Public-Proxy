package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerMultiKeyUpdate_DisablesOnlyFailedKey(t *testing.T) {
	channel := &Channel{
		Id:     1,
		Status: common.ChannelStatusEnabled,
		Key:    "key-a\nkey-b\nkey-c",
		ChannelInfo: ChannelInfo{
			IsMultiKey: true,
		},
	}

	handlerMultiKeyUpdate(channel, "key-b", 1, common.ChannelStatusAutoDisabled, "no credit balance")

	assert.Equal(t, common.ChannelStatusEnabled, channel.Status, "channel must stay enabled while other keys remain")
	require.NotNil(t, channel.ChannelInfo.MultiKeyStatusList)
	assert.Equal(t, common.ChannelStatusAutoDisabled, channel.ChannelInfo.MultiKeyStatusList[1])
	assert.Len(t, channel.ChannelInfo.MultiKeyStatusList, 1)
	_, hasA := channel.ChannelInfo.MultiKeyStatusList[0]
	_, hasC := channel.ChannelInfo.MultiKeyStatusList[2]
	assert.False(t, hasA)
	assert.False(t, hasC)
}

func TestHandlerMultiKeyUpdate_PrefersIndexOverMismatchedKey(t *testing.T) {
	channel := &Channel{
		Id:     2,
		Status: common.ChannelStatusEnabled,
		Key:    "key-a\nkey-b\nkey-c",
		ChannelInfo: ChannelInfo{
			IsMultiKey: true,
		},
	}

	handlerMultiKeyUpdate(channel, "stale-or-rotated-key", 0, common.ChannelStatusAutoDisabled, "permission_error")

	assert.Equal(t, common.ChannelStatusEnabled, channel.Status)
	require.Equal(t, common.ChannelStatusAutoDisabled, channel.ChannelInfo.MultiKeyStatusList[0])
}

func TestHandlerMultiKeyUpdate_DoesNotDisableChannelWhenKeyUnresolved(t *testing.T) {
	channel := &Channel{
		Id:     3,
		Status: common.ChannelStatusEnabled,
		Key:    "key-a\nkey-b",
		ChannelInfo: ChannelInfo{
			IsMultiKey: true,
		},
	}

	handlerMultiKeyUpdate(channel, "missing-key", -1, common.ChannelStatusAutoDisabled, "should no-op")

	assert.Equal(t, common.ChannelStatusEnabled, channel.Status)
	assert.Empty(t, channel.ChannelInfo.MultiKeyStatusList)
}

func TestHandlerMultiKeyUpdate_AllKeysDisabledSetsChannelAutoDisabled(t *testing.T) {
	channel := &Channel{
		Id:     4,
		Status: common.ChannelStatusEnabled,
		Key:    "key-a\nkey-b",
		ChannelInfo: ChannelInfo{
			IsMultiKey: true,
		},
	}

	handlerMultiKeyUpdate(channel, "", 0, common.ChannelStatusAutoDisabled, "key-a dead")
	handlerMultiKeyUpdate(channel, "", 1, common.ChannelStatusAutoDisabled, "key-b dead")

	assert.Equal(t, common.ChannelStatusAutoDisabled, channel.Status)
	assert.Equal(t, "All keys are disabled", channel.GetOtherInfo()["status_reason"])
}

func TestHandlerMultiKeyUpdate_EmptyKeyWithoutIndexIsChannelLevel(t *testing.T) {
	channel := &Channel{
		Id:     5,
		Status: common.ChannelStatusEnabled,
		Key:    "key-a\nkey-b",
		ChannelInfo: ChannelInfo{
			IsMultiKey: true,
		},
	}

	handlerMultiKeyUpdate(channel, "", -1, common.ChannelStatusManuallyDisabled, "manual operation")

	assert.Equal(t, common.ChannelStatusManuallyDisabled, channel.Status)
	assert.Empty(t, channel.ChannelInfo.MultiKeyStatusList)
}
