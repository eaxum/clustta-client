package agent

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestOpenAIModelOptions(t *testing.T) {
	want := []string{openAIModelTerra, openAIModelSol, openAIModelLuna}

	if got := GetProviderModels("openai"); !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProviderModels(openai) = %v, want %v", got, want)
	}
	if got := GetDefaultModel("openai"); got != openAIModelTerra {
		t.Fatalf("GetDefaultModel(openai) = %q, want %q", got, openAIModelTerra)
	}
}

func TestResolveProviderModel(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     string
	}{
		{name: "openai default", provider: "openai", want: openAIModelTerra},
		{name: "current openai selection", provider: "openai", model: openAIModelSol, want: openAIModelSol},
		{name: "unknown openai selection", provider: "openai", model: "unknown", want: openAIModelTerra},
		{name: "other provider selection", provider: "ollama", model: "custom", want: "custom"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ResolveProviderModel(test.provider, test.model); got != test.want {
				t.Fatalf("ResolveProviderModel(%q, %q) = %q, want %q", test.provider, test.model, got, test.want)
			}
		})
	}
}

func TestReasoningEffortForOpenAIModels(t *testing.T) {
	for _, model := range []string{openAIModelTerra, openAIModelSol, openAIModelLuna} {
		t.Run(model, func(t *testing.T) {
			if got := reasoningEffortForModel(model); got != reasoningEffortNone {
				t.Fatalf("reasoningEffortForModel(%q) = %q, want %q", model, got, reasoningEffortNone)
			}
		})
	}
}

func TestReasoningEffortForCompatibleProviderModel(t *testing.T) {
	if got := reasoningEffortForModel("gemini-2.5-flash"); got != "" {
		t.Fatalf("reasoningEffortForModel(gemini-2.5-flash) = %q, want empty", got)
	}
}

func TestOpenAIRequestIncludesDisabledReasoning(t *testing.T) {
	request := buildOpenAIRequest(openAIModelTerra, nil, []openAITool{{Type: "function"}})
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(payload), `"reasoning_effort":"none"`) {
		t.Fatalf("request payload does not disable reasoning: %s", payload)
	}
}

func TestOpenAIRequestOmitsDisplayContent(t *testing.T) {
	messages := []Message{{Role: "user", Content: "expanded prompt", DisplayContent: "/who"}}
	request := buildOpenAIRequest(openAIModelTerra, messages, nil)
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(payload), "display_content") || strings.Contains(string(payload), "/who") {
		t.Fatalf("request payload includes UI-only content: %s", payload)
	}
}
