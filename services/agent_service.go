package services

import (
	"clustta/internal/agent"
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/settings"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const agentIntegrationID = "clustta-agent-llm"

// legacyAgentIntegrationID is the pre-kebab-case identifier used in earlier
// builds. Existing installs may still have credentials stored under this key
// in the OS keyring and settings.json; migrateLegacyAgentIntegrationID copies
// them to agentIntegrationID once, then removes the legacy entries.
const legacyAgentIntegrationID = "clustta_agent_llm"

var migrateLegacyAgentIDOnce sync.Once

// migrateLegacyAgentIntegrationID moves any credential stored under the old
// underscore identifier to the new dashed one. Idempotent and safe to call
// repeatedly; the actual work runs at most once per process.
func migrateLegacyAgentIntegrationID() {
	migrateLegacyAgentIDOnce.Do(func() {
		legacy, err := settings.GetIntegrationCredential(legacyAgentIntegrationID)
		if err != nil {
			return
		}
		// Don't clobber an existing new-style entry.
		if _, err := settings.GetIntegrationCredential(agentIntegrationID); err == nil {
			_ = settings.DeleteIntegrationCredential(legacyAgentIntegrationID)
			return
		}
		legacy.IntegrationId = agentIntegrationID
		if err := settings.SaveIntegrationCredential(legacy); err != nil {
			return
		}
		_ = settings.DeleteIntegrationCredential(legacyAgentIntegrationID)
	})
}

// AgentService exposes the AI agent to the frontend via Wails bindings.
type AgentService struct {
	mu        sync.Mutex
	history   map[string][]agent.Message       // keyed by projectPath
	cancels   map[string]context.CancelFunc    // active runs keyed by projectPath
	approvals map[string]chan approvalDecision // pending approval channels keyed by tool call ID
}

type approvalDecision struct {
	approved     bool
	selectedKeys []string
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

// maxHistoryMessages caps the per-project chat history that is kept in memory
// and persisted to disk. Beyond this, the oldest non-system messages are
// dropped. Without this cap the file grows without bound and eventually blows
// the LLM's context window on every send.
const maxHistoryMessages = 200

// trimHistory drops the oldest messages so the history stays under the cap.
// We keep the *last* maxHistoryMessages entries because the most recent turns
// matter most for the LLM's understanding.
//
// Tool-result messages must stay paired with their preceding assistant
// tool-call message; if we trim mid-pair the LLM rejects the request. We
// shift the start forward until it lands on a "user" or "system" boundary.
func trimHistory(history []agent.Message) []agent.Message {
	if len(history) <= maxHistoryMessages {
		return history
	}
	start := len(history) - maxHistoryMessages
	for start < len(history) && history[start].Role == "tool" {
		start++
	}
	if start >= len(history) {
		return history[len(history)-1:]
	}
	trimmed := make([]agent.Message, len(history)-start)
	copy(trimmed, history[start:])
	return trimmed
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

	migrateLegacyAgentIntegrationID()

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

	// Look up user-selected model for this provider; empty string means use provider default.
	model, _ := settings.GetAgentModel(provider)

	// For Ollama, verify reachability and that the model is pulled before
	// kicking off the run; surfaces a clear error instead of a generic 404.
	// Also reminds the user that prompts/data are sent to the local Ollama process.
	if provider == "ollama" {
		ollamaCtx, cancelVerify := context.WithTimeout(context.Background(), 5*time.Second)
		verifyErr := agent.VerifyOllamaModel(ollamaCtx, model)
		cancelVerify()
		if verifyErr != nil {
			app.Event.Emit("agent-error", verifyErr.Error())
			return verifyErr
		}
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
	// If a previous run for this project is still in flight, cancel it before starting a new one.
	if a.cancels == nil {
		a.cancels = map[string]context.CancelFunc{}
	}
	if prev, ok := a.cancels[projectPath]; ok {
		prev()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancels[projectPath] = cancel
	a.mu.Unlock()

	// Run agent in a goroutine so we don't block the UI
	go func() {
		emit := func(eventName string, data interface{}) {
			app.Event.Emit(eventName, data)
		}

		requestApproval := func(ctx context.Context, toolCallID, toolName, riskLevel string, args map[string]interface{}, preview interface{}) bool {
			return a.requestApproval(ctx, projectPath, toolCallID, toolName, riskLevel, args, preview)
		}
		updatedHistory, mutated, err := agent.RunAgent(ctx, projectPath, history, message, attachmentContent, cred.AccessToken, provider, model, emit, requestApproval)
		if err != nil && !errors.Is(err, agent.ErrCancelled) {
			app.Event.Emit("agent-error", err.Error())
		}

		// Persist updated history (capped to keep the file and the LLM context window bounded)
		a.mu.Lock()
		a.history[projectPath] = trimHistory(updatedHistory)
		_ = saveChatSessions(a.history)
		if a.cancels != nil {
			delete(a.cancels, projectPath)
		}
		a.mu.Unlock()

		payload := map[string]interface{}{"mutated": mutated}
		if errors.Is(err, agent.ErrCancelled) {
			app.Event.Emit("agent-cancelled", payload)
		}
		app.Event.Emit("agent-done", payload)
	}()

	return nil
}

// CancelRun cancels an in-flight agent run for the given project.
// Safe to call when no run is active (returns nil).
func (a *AgentService) CancelRun(projectPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancels == nil {
		return nil
	}
	if cancel, ok := a.cancels[projectPath]; ok {
		cancel()
		delete(a.cancels, projectPath)
	}
	// Also unblock any pending approval prompts so the goroutine exits cleanly.
	for id, ch := range a.approvals {
		select {
		case ch <- approvalDecision{}:
		default:
		}
		delete(a.approvals, id)
	}
	return nil
}

// requestApproval is the agent.ApprovalRequester implementation. It emits an
// agent-tool-approval-request event with full context and blocks until either
// ApproveToolCall is called for this tool call ID or the context is cancelled.
// Returns true to allow execution, false to deny.
func (a *AgentService) requestApproval(ctx context.Context, projectPath, toolCallID, toolName, riskLevel string, args map[string]interface{}, preview interface{}) bool {
	// Honor the auto-approve setting.
	_, isPlannedCommand := agentcommands.DefinitionFor(toolName)
	if auto, _ := settings.GetAgentAutoApproveDestructive(); auto && !isPlannedCommand {
		return true
	}

	a.mu.Lock()
	if a.approvals == nil {
		a.approvals = map[string]chan approvalDecision{}
	}
	ch := make(chan approvalDecision, 1)
	a.approvals[toolCallID] = ch
	a.mu.Unlock()

	application.Get().Event.Emit("agent-tool-approval-request", map[string]interface{}{
		"id":      toolCallID,
		"tool":    toolName,
		"risk":    riskLevel,
		"args":    args,
		"preview": preview,
	})

	select {
	case decision := <-ch:
		a.mu.Lock()
		delete(a.approvals, toolCallID)
		a.mu.Unlock()
		if !decision.approved {
			return false
		}
		if isPlannedCommand {
			if err := agentcommands.PrepareSelection(projectPath, toolName, args, decision.selectedKeys); err != nil {
				return false
			}
		}
		return true
	case <-ctx.Done():
		a.mu.Lock()
		delete(a.approvals, toolCallID)
		a.mu.Unlock()
		return false
	}
}

// ApproveToolCall responds to an agent-tool-approval-request from the frontend.
// Pass approved=true to allow the checked items to run, false to deny it.
// Safe to call with an unknown ID (returns nil).
func (a *AgentService) ApproveToolCall(toolCallID string, approved bool, selectedKeys []string) error {
	a.mu.Lock()
	ch, ok := a.approvals[toolCallID]
	if ok {
		delete(a.approvals, toolCallID)
	}
	a.mu.Unlock()
	if !ok {
		return nil
	}
	select {
	case ch <- approvalDecision{approved: approved, selectedKeys: selectedKeys}:
	default:
	}
	return nil
}

// GetAutoApproveDestructive reports whether destructive tool calls auto-execute
// without prompting the user.
func (a *AgentService) GetAutoApproveDestructive() bool {
	v, _ := settings.GetAgentAutoApproveDestructive()
	return v
}

// SetAutoApproveDestructive persists the auto-approve preference.
func (a *AgentService) SetAutoApproveDestructive(enabled bool) error {
	return settings.SetAgentAutoApproveDestructive(enabled)
}

// GetAvailableModels returns the list of selectable models for the given provider.
// The first entry is the provider's default. Empty slice for unknown providers.
func (a *AgentService) GetAvailableModels(provider string) []string {
	models := agent.GetProviderModels(provider)
	if models == nil {
		return []string{}
	}
	return models
}

// GetSelectedModel returns the user-selected model for the provider, or the
// provider default if no override has been set.
func (a *AgentService) GetSelectedModel(provider string) string {
	chosen, _ := settings.GetAgentModel(provider)
	return agent.ResolveProviderModel(provider, chosen)
}

// SetSelectedModel persists the chosen model for the given provider.
// Passing an empty model clears the override and reverts to the provider default.
func (a *AgentService) SetSelectedModel(provider, model string) error {
	return settings.SetAgentModel(provider, model)
}

// SetAPIKey stores the user's LLM API key in settings.
// Provider should be "openai", "anthropic", "gemini", "groq", or "ollama".
func (a *AgentService) SetAPIKey(provider, apiKey string) error {
	migrateLegacyAgentIntegrationID()
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
	migrateLegacyAgentIntegrationID()
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
	migrateLegacyAgentIntegrationID()
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

// RetryLastTurn cancels any running agent execution, deletes the last user message and subsequent assistant messages/errors, and restarts the run.
func (a *AgentService) RetryLastTurn(projectPath string) error {
	a.mu.Lock()
	history := a.getHistory(projectPath)

	// Find the last user message
	lastUserIdx := -1
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "user" {
			lastUserIdx = i
			break
		}
	}

	if lastUserIdx == -1 {
		a.mu.Unlock()
		return fmt.Errorf("no user message to retry")
	}

	// Extract the user message content
	userMsgContent := fmt.Sprintf("%v", history[lastUserIdx].Content)

	// Remove the last user message and everything after it
	history = history[:lastUserIdx]
	a.history[projectPath] = history
	_ = saveChatSessions(a.history)
	a.mu.Unlock()

	// SendMessage will cancel any in-flight run and start a new one
	return a.SendMessage(projectPath, userMsgContent, "")
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
