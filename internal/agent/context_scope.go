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
	for _, key := range []string{"scope", "target_scope", "dependency_scope", "source_scope"} {
		applyAuthoritativeScopeField(args, key, context)
	}
}

func applyAuthoritativeScopeField(args map[string]interface{}, key string, context turnContext) {
	raw, ok := args[key]
	if !ok {
		return
	}
	scopeArgs, ok := raw.(map[string]interface{})
	if !ok {
		return
	}
	// Some providers occasionally place scope members beside `scope` even
	// though the generated schema nests them. Normalize those fields before
	// parsing so intent such as asset-only scope cannot be silently ignored.
	for _, member := range []string{"types", "type", "filters", "recursive", "limit"} {
		if _, nested := scopeArgs[member]; nested {
			continue
		}
		if value, misplaced := args[member]; misplaced && key == "scope" {
			scopeArgs[member] = value
			delete(args, member)
		}
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
		// Browser asset-only mode presents descendant assets as the effective
		// contents of "here". Preserve an explicitly recursive model request,
		// and otherwise promote the scope when the browser context is recursive.
		if contextBool(context.HereScope, "recursive") {
			scopeArgs["recursive"] = true
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

func contextBool(values map[string]interface{}, key string) bool {
	value, _ := values[key].(bool)
	return value
}
