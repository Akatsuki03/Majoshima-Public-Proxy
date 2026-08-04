package claude

import (
	"regexp"

	kitutil "github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
)

// Anthropic requires tool_use.id / tool_result.tool_use_id to match
// ^[a-zA-Z0-9_-]+$. Clients relaying conversations that originated from other
// providers (or gateway-synthesized ids like "output:0") may carry ids with
// characters outside this set, which Claude rejects with a 400.
var invalidToolIdChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// SanitizeToolId makes a tool call id acceptable to the Anthropic API.
// The mapping is deterministic (pure), so a tool_use id and its matching
// tool_result.tool_use_id from the same request body always sanitize to the
// same value. Ids that are already valid are returned unchanged.
func SanitizeToolId(id string) string {
	if id == "" {
		return "call_" + invalidToolIdChars.ReplaceAllString(kitutil.GetUUID(), "_")
	}
	if !invalidToolIdChars.MatchString(id) {
		return id
	}
	return invalidToolIdChars.ReplaceAllString(id, "_")
}
