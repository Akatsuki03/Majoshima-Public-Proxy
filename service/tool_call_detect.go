package service

import (
	"github.com/QuantumNous/new-api/dto"
)

func isEmptyRawJson(raw []byte) bool {
	switch string(raw) {
	case "", "null", "[]", "{}":
		return true
	}
	return false
}

// RequestContainsToolCall reports whether a relay request involves tool calls:
// tools/functions declared in the request, or tool call / tool result messages
// carried in the conversation history.
func RequestContainsToolCall(req dto.Request) bool {
	if req == nil {
		return false
	}
	switch r := req.(type) {
	case *dto.GeneralOpenAIRequest:
		if len(r.Tools) > 0 || !isEmptyRawJson(r.Functions) {
			return true
		}
		for i := range r.Messages {
			m := &r.Messages[i]
			if m.Role == "tool" || m.ToolCallId != "" || !isEmptyRawJson(m.ToolCalls) {
				return true
			}
		}
	case *dto.ClaudeRequest:
		return len(r.GetTools()) > 0
	case *dto.OpenAIResponsesRequest:
		return !isEmptyRawJson(r.Tools)
	case *dto.GeminiChatRequest:
		return len(r.GetTools()) > 0
	}
	return false
}
