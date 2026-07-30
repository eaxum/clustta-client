package agent

import (
	"encoding/json"
	"strings"
)

type turnContext struct {
	CurrentLocation map[string]interface{}   `json:"current_location"`
	HereScope       map[string]interface{}   `json:"here_scope"`
	Selection       []map[string]interface{} `json:"selection"`
}

func parseTurnContext(message string) turnContext {
	const marker = "[Context JSON]\n"
	start := strings.Index(message, marker)
	if start < 0 {
		return turnContext{}
	}
	line := message[start+len(marker):]
	if end := strings.IndexByte(line, '\n'); end >= 0 {
		line = line[:end]
	}
	var context turnContext
	_ = json.Unmarshal([]byte(line), &context)
	return context
}

// applyAuthoritativeScope makes browser context authoritative for contextual
// scopes. The model still chooses the operation, filters, types and recursion,
// but cannot omit or substitute the current location or selected entities.
func applyAuthoritativeScope(args map[string]interface{}, context turnContext) {
	raw, ok := args["scope"]
	if !ok {
		return
	}
	scopeArgs, ok := raw.(map[string]interface{})
	if !ok {
		return
	}
	normalizeScopeTypes(scopeArgs)
	source, _ := scopeArgs["source"].(string)
	switch source {
	case "here":
		delete(scopeArgs, "id")
		delete(scopeArgs, "entity_id")
		delete(scopeArgs, "path")
		if id := contextString(context.HereScope, "entity_id"); id != "" {
			scopeArgs["entity_id"] = id
		}
		if path := contextString(context.HereScope, "path"); path != "" {
			scopeArgs["path"] = path
		}
	case "selection":
		scopeArgs["selection"] = context.Selection
	}
}

func normalizeScopeTypes(scopeArgs map[string]interface{}) {
	if _, exists := scopeArgs["types"]; exists {
		return
	}
	if value, exists := scopeArgs["type"]; exists {
		if text, ok := value.(string); ok {
			scopeArgs["types"] = []interface{}{text}
		} else {
			scopeArgs["types"] = value
		}
		delete(scopeArgs, "type")
	}
}

func contextString(values map[string]interface{}, key string) string {
	value, _ := values[key].(string)
	return strings.TrimSpace(value)
}
