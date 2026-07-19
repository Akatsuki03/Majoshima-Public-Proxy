package helper

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAndValidateTextRequestNormalizesCursorTool(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(
		`{"model":"claude-test","messages":[{"role":"user","content":"hi"}],"tools":[{"name":"lookup","description":"Lookup data","input_schema":{"type":"object"}}]}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")

	req, err := GetAndValidateTextRequest(c, relayconstant.RelayModeChatCompletions)

	require.NoError(t, err)
	require.Len(t, req.Tools, 1)
	assert.Equal(t, "function", req.Tools[0].Type)
	assert.Equal(t, "lookup", req.Tools[0].Function.Name)
	assert.Equal(t, "Lookup data", req.Tools[0].Function.Description)
	require.IsType(t, map[string]any{}, req.Tools[0].Function.Parameters)
	assert.Equal(t, "object", req.Tools[0].Function.Parameters.(map[string]any)["type"])
}

func TestGetAndValidateTextRequestRejectsEmptyFunctionToolName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		tool string
	}{
		{name: "missing tool fields", tool: `{}`},
		{name: "empty openai function name", tool: `{"type":"function","function":{"name":" "}}`},
		{name: "empty cursor function name", tool: `{"description":"Lookup data","input_schema":{"type":"object"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			body := `{"model":"claude-test","messages":[{"role":"user","content":"hi"}],"tools":[` + tt.tool + `]}`
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
			c.Request.Header.Set("Content-Type", "application/json")

			_, err := GetAndValidateTextRequest(c, relayconstant.RelayModeChatCompletions)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "tools[0].function.name is required")
		})
	}
}

func TestGetAndValidateTextRequestNormalizesAnthropicToolBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{
		"model":"claude-test",
		"messages":[
			{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"lookup","input":{"q":"x"}}]},
			{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"ok"}]}
		]
	}`))
	c.Request.Header.Set("Content-Type", "application/json")

	req, err := GetAndValidateTextRequest(c, relayconstant.RelayModeChatCompletions)

	require.NoError(t, err)
	require.Len(t, req.Messages, 2)
	assert.Equal(t, "assistant", req.Messages[0].Role)
	toolCalls := req.Messages[0].ParseToolCalls()
	require.Len(t, toolCalls, 1)
	assert.Equal(t, "lookup", toolCalls[0].Function.Name)
	assert.Equal(t, "tool", req.Messages[1].Role)
	assert.Equal(t, "toolu_1", req.Messages[1].ToolCallId)
	assert.Equal(t, "ok", req.Messages[1].StringContent())
}
