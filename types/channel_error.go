package types

type ChannelError struct {
	ChannelId     int    `json:"channel_id"`
	ChannelType   int    `json:"channel_type"`
	ChannelName   string `json:"channel_name"`
	IsMultiKey    bool   `json:"is_multi_key"`
	AutoBan       bool   `json:"auto_ban"`
	UsingKey      string `json:"using_key"`
	UsingKeyIndex int    `json:"using_key_index"` // -1 = unknown; for multi-key prefer index over string match
}

func NewChannelError(channelId int, channelType int, channelName string, isMultiKey bool, usingKey string, usingKeyIndex int, autoBan bool) *ChannelError {
	return &ChannelError{
		ChannelId:     channelId,
		ChannelType:   channelType,
		ChannelName:   channelName,
		IsMultiKey:    isMultiKey,
		AutoBan:       autoBan,
		UsingKey:      usingKey,
		UsingKeyIndex: usingKeyIndex,
	}
}
