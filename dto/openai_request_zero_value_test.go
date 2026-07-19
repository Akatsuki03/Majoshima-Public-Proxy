package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGeneralOpenAIRequestPreserveExplicitZeroValues(t *testing.T) {
	raw := []byte(`{
		"model":"gpt-4.1",
		"stream":false,
		"max_tokens":0,
		"max_completion_tokens":0,
		"top_p":0,
		"top_k":0,
		"n":0,
		"frequency_penalty":0,
		"presence_penalty":0,
		"seed":0,
		"logprobs":false,
		"top_logprobs":0,
		"dimensions":0,
		"return_images":false,
		"return_related_questions":false
	}`)

	var req GeneralOpenAIRequest
	err := common.Unmarshal(raw, &req)
	require.NoError(t, err)

	encoded, err := common.Marshal(req)
	require.NoError(t, err)

	require.True(t, gjson.GetBytes(encoded, "stream").Exists())
	require.True(t, gjson.GetBytes(encoded, "max_tokens").Exists())
	require.True(t, gjson.GetBytes(encoded, "max_completion_tokens").Exists())
	require.True(t, gjson.GetBytes(encoded, "top_p").Exists())
	require.True(t, gjson.GetBytes(encoded, "top_k").Exists())
	require.True(t, gjson.GetBytes(encoded, "n").Exists())
	require.True(t, gjson.GetBytes(encoded, "frequency_penalty").Exists())
	require.True(t, gjson.GetBytes(encoded, "presence_penalty").Exists())
	require.True(t, gjson.GetBytes(encoded, "seed").Exists())
	require.True(t, gjson.GetBytes(encoded, "logprobs").Exists())
	require.True(t, gjson.GetBytes(encoded, "top_logprobs").Exists())
	require.True(t, gjson.GetBytes(encoded, "dimensions").Exists())
	require.True(t, gjson.GetBytes(encoded, "return_images").Exists())
	require.True(t, gjson.GetBytes(encoded, "return_related_questions").Exists())
}

func TestOpenAIResponsesRequestPreserveExplicitZeroValues(t *testing.T) {
	raw := []byte(`{
		"model":"gpt-4.1",
		"max_output_tokens":0,
		"max_tool_calls":0,
		"stream":false,
		"top_p":0
	}`)

	var req OpenAIResponsesRequest
	err := common.Unmarshal(raw, &req)
	require.NoError(t, err)

	encoded, err := common.Marshal(req)
	require.NoError(t, err)

	require.True(t, gjson.GetBytes(encoded, "max_output_tokens").Exists())
	require.True(t, gjson.GetBytes(encoded, "max_tool_calls").Exists())
	require.True(t, gjson.GetBytes(encoded, "stream").Exists())
	require.True(t, gjson.GetBytes(encoded, "top_p").Exists())
}

func TestGeneralOpenAIRequestGetSystemRoleName(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{name: "o1 uses developer", model: "o1", want: "developer"},
		{name: "o3 family uses developer", model: "o3-mini-high", want: "developer"},
		{name: "o4 family uses developer", model: "o4-mini", want: "developer"},
		{name: "o1 mini stays system", model: "o1-mini", want: "system"},
		{name: "o1 preview stays system", model: "o1-preview", want: "system"},
		{name: "gpt 5 uses developer", model: "gpt-5", want: "developer"},
		{name: "omni is not o series", model: "omni-moderation-latest", want: "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := GeneralOpenAIRequest{Model: tt.model}

			require.Equal(t, tt.want, req.GetSystemRoleName())
		})
	}
}

func TestToolCallRequestNormalizesCompatibleFunctionShapes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "openai chat",
			raw:  `{"type":"function","function":{"name":"lookup","description":"Lookup data","parameters":{"type":"object"}}}`,
		},
		{
			name: "cursor anthropic style",
			raw:  `{"name":"lookup","description":"Lookup data","input_schema":{"type":"object"}}`,
		},
		{
			name: "responses flat style",
			raw:  `{"type":"function","name":"lookup","description":"Lookup data","parameters":{"type":"object"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tool ToolCallRequest
			require.NoError(t, common.Unmarshal([]byte(tt.raw), &tool))

			assert.Equal(t, "function", tool.Type)
			assert.Equal(t, "lookup", tool.Function.Name)
			assert.Equal(t, "Lookup data", tool.Function.Description)
			require.IsType(t, map[string]any{}, tool.Function.Parameters)
			assert.Equal(t, "object", tool.Function.Parameters.(map[string]any)["type"])

			encoded, err := common.Marshal(tool)
			require.NoError(t, err)
			assert.Equal(t, "function", gjson.GetBytes(encoded, "type").String())
			assert.Equal(t, "lookup", gjson.GetBytes(encoded, "function.name").String())
			assert.Equal(t, "object", gjson.GetBytes(encoded, "function.parameters.type").String())
			assert.False(t, gjson.GetBytes(encoded, "name").Exists())
			assert.False(t, gjson.GetBytes(encoded, "input_schema").Exists())
		})
	}
}

func TestNormalizeAnthropicToolMessageBlocks(t *testing.T) {
	raw := []byte(`{
		"model":"claude-test",
		"messages":[
			{"role":"user","content":"what is the weather?"},
			{"role":"assistant","content":[
				{"type":"text","text":"I will check."},
				{"type":"tool_use","id":"toolu_1","name":"lookup","input":{"q":"x"}}
			]},
			{"role":"user","content":[
				{"type":"tool_result","tool_use_id":"toolu_1","content":"{\"ok\":true}"},
				{"type":"text","text":"thanks"}
			]}
		]
	}`)

	var req GeneralOpenAIRequest
	require.NoError(t, common.Unmarshal(raw, &req))
	req.NormalizeAnthropicToolMessageBlocks()

	require.Len(t, req.Messages, 4)
	assert.Equal(t, "user", req.Messages[0].Role)
	assert.Equal(t, "what is the weather?", req.Messages[0].StringContent())

	assert.Equal(t, "assistant", req.Messages[1].Role)
	assert.Equal(t, "I will check.", req.Messages[1].StringContent())
	toolCalls := req.Messages[1].ParseToolCalls()
	require.Len(t, toolCalls, 1)
	assert.Equal(t, "toolu_1", toolCalls[0].ID)
	assert.Equal(t, "function", toolCalls[0].Type)
	assert.Equal(t, "lookup", toolCalls[0].Function.Name)
	assert.Equal(t, `{"q":"x"}`, toolCalls[0].Function.Arguments)

	assert.Equal(t, "tool", req.Messages[2].Role)
	assert.Equal(t, "toolu_1", req.Messages[2].ToolCallId)
	assert.Equal(t, `{"ok":true}`, req.Messages[2].StringContent())

	assert.Equal(t, "user", req.Messages[3].Role)
	assert.Equal(t, "thanks", req.Messages[3].StringContent())

	encoded, err := common.Marshal(req)
	require.NoError(t, err)
	assert.False(t, gjson.GetBytes(encoded, `messages.#(content.#(type=="tool_result"))`).Exists())
	assert.False(t, gjson.GetBytes(encoded, `messages.#(content.#(type=="tool_use"))`).Exists())
	assert.Equal(t, "tool", gjson.GetBytes(encoded, "messages.2.role").String())
	assert.Equal(t, "lookup", gjson.GetBytes(encoded, "messages.1.tool_calls.0.function.name").String())
}
