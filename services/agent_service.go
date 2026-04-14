package services

import (
	"clustta/internal/agent"
	"clustta/internal/settings"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const agentIntegrationID = "clustta_agent_llm"

// AgentService exposes the AI agent to the frontend via Wails bindings.
type AgentService struct {
	mu      sync.Mutex
	history map[string][]agent.Message // keyed by projectPath
}

// AgentKeyStatus reports whether an API key is configured and which provider is selected.
type AgentKeyStatus struct {
	Configured bool   `json:"configured"`
	Provider   string `json:"provider"`
}

// ChatUIMessage represents a message in the format the frontend expects for rendering.
type ChatUIMessage struct {
	Type     string `json:"type"` // "user", "assistant", "tool-group"
	Content  string `json:"content"`
	ToolName string `json:"toolName,omitempty"` // tool function name for tool-group
	Count    int    `json:"count,omitempty"`    // number of calls for tool-group
}

// chatSessionFile holds persisted conversation history.
type chatSessionFile struct {
	Sessions map[string][]agent.Message `json:"sessions"`
}

// getChatSessionPath returns the path to the chat session JSON file.
func getChatSessionPath() (string, error) {
	roamingPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(roamingPath, "Clustta", ".agent_sessions.json"), nil
}

// loadChatSessions loads persisted chat sessions from disk.
func loadChatSessions() (map[string][]agent.Message, error) {
	path, err := getChatSessionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]agent.Message), nil
		}
		return nil, err
	}

	var file chatSessionFile
	if err := json.Unmarshal(data, &file); err != nil {
		return make(map[string][]agent.Message), nil
	}
	if file.Sessions == nil {
		return make(map[string][]agent.Message), nil
	}
	return file.Sessions, nil
}

// saveChatSessions persists chat sessions to disk.
func saveChatSessions(sessions map[string][]agent.Message) error {
	path, err := getChatSessionPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	data, err := json.MarshalIndent(chatSessionFile{Sessions: sessions}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// getHistory returns the conversation history for a project, loading from disk if needed.
func (a *AgentService) getHistory(projectPath string) []agent.Message {
	if a.history == nil {
		a.history = make(map[string][]agent.Message)
		if saved, err := loadChatSessions(); err == nil {
			a.history = saved
		}
	}
	return a.history[projectPath]
}

// SendMessage sends a user message to the agent, which calls the LLM and executes tools.
// Results are streamed back via Wails events: agent-status, agent-tool-start, agent-tool-result, agent-response, agent-error.
func (a *AgentService) SendMessage(projectPath, message, attachmentPath string) error {
	app := application.Get()

	// Load API key
	cred, err := settings.GetIntegrationCredential(agentIntegrationID)
	if err != nil || (cred.AccessToken == "" && cred.ApiUrl != "ollama") {
		app.Event.Emit("agent-error", "No API key configured. Please set your LLM provider in Settings > Advanced.")
		return fmt.Errorf("no API key configured")
	}

	provider := cred.ApiUrl // We store the provider name in ApiUrl field
	if provider == "" {
		provider = "openai"
	}

	// Read attachment if provided
	attachmentContent, err := agent.ReadAttachment(attachmentPath)
	if err != nil {
		app.Event.Emit("agent-error", fmt.Sprintf("Failed to read attachment: %s", err.Error()))
		return err
	}

	// Get existing conversation history
	a.mu.Lock()
	history := a.getHistory(projectPath)
	a.mu.Unlock()

	// Run agent in a goroutine so we don't block the UI
	go func() {
		emit := func(eventName string, data interface{}) {
			app.Event.Emit(eventName, data)
		}

		updatedHistory, err := agent.RunAgent(projectPath, history, message, attachmentContent, cred.AccessToken, provider, emit)
		if err != nil {
			app.Event.Emit("agent-error", err.Error())
		}

		// Persist updated history
		a.mu.Lock()
		a.history[projectPath] = updatedHistory
		_ = saveChatSessions(a.history)
		a.mu.Unlock()

		app.Event.Emit("agent-done", nil)
	}()

	return nil
}

// SetAPIKey stores the user's LLM API key in settings.
// Provider should be "openai", "anthropic", "gemini", "groq", or "ollama".
func (a *AgentService) SetAPIKey(provider, apiKey string) error {
	validProviders := map[string]bool{
		"openai": true, "anthropic": true, "gemini": true, "groq": true, "ollama": true,
	}
	if !validProviders[provider] {
		return fmt.Errorf("unsupported provider: %s", provider)
	}
	if provider != "ollama" && apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	cred := settings.IntegrationCredential{
		IntegrationId: agentIntegrationID,
		AccessToken:   apiKey,
		ApiUrl:        provider, // Store provider name in ApiUrl field
	}
	return settings.SaveIntegrationCredential(cred)
}

// GetAPIKeyStatus checks if an API key is configured. Never returns the key itself.
func (a *AgentService) GetAPIKeyStatus() AgentKeyStatus {
	cred, err := settings.GetIntegrationCredential(agentIntegrationID)
	if err != nil {
		return AgentKeyStatus{Configured: false, Provider: ""}
	}
	provider := cred.ApiUrl
	if provider == "" {
		provider = "openai"
	}
	// Ollama doesn't require an API key
	if cred.AccessToken == "" && provider != "ollama" {
		return AgentKeyStatus{Configured: false, Provider: ""}
	}
	return AgentKeyStatus{Configured: true, Provider: provider}
}

// RemoveAPIKey deletes the stored API key.
func (a *AgentService) RemoveAPIKey() error {
	return settings.DeleteIntegrationCredential(agentIntegrationID)
}

// ClearChatSession clears the conversation history for a project.
func (a *AgentService) ClearChatSession(projectPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.history == nil {
		a.history = make(map[string][]agent.Message)
	}
	delete(a.history, projectPath)
	return saveChatSessions(a.history)
}

// GetChatHistory returns the persisted conversation history mapped to UI message types.
// Tool-calling assistant messages are grouped by tool name.
func (a *AgentService) GetChatHistory(projectPath string) []ChatUIMessage {
	a.mu.Lock()
	history := a.getHistory(projectPath)
	a.mu.Unlock()

	var result []ChatUIMessage
	for _, msg := range history {
		switch msg.Role {
		case "user":
			content := fmt.Sprintf("%v", msg.Content)
			if idx := strings.Index(content, "\n\n--- Attached Content ---\n"); idx >= 0 {
				content = content[:idx]
			}
			result = append(result, ChatUIMessage{Type: "user", Content: content})
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				groups := make(map[string]int)
				var order []string
				for _, tc := range msg.ToolCalls {
					name := tc.Function.Name
					if groups[name] == 0 {
						order = append(order, name)
					}
					groups[name]++
				}
				for _, name := range order {
					result = append(result, ChatUIMessage{Type: "tool-group", ToolName: name, Count: groups[name]})
				}
			} else {
				content := fmt.Sprintf("%v", msg.Content)
				result = append(result, ChatUIMessage{Type: "assistant", Content: content})
			}
		}
	}
	return result
}
