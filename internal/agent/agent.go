package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	pdf "github.com/ledongthuc/pdf"
)

// ErrCancelled is returned when the agent run was cancelled by the user.
var ErrCancelled = errors.New("agent run cancelled")

const maxIterations = 25

// llmHTTPClient is the shared client for LLM API calls. Bounded timeouts at
// every layer prevent a slow or hung provider from leaking goroutines.
var llmHTTPClient = &http.Client{
	Timeout: 120 * time.Second,
	Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	},
}

// Message represents a chat message in the conversation.
type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall holds the function name and arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// EventEmitter is a callback for sending events to the frontend.
type EventEmitter func(eventName string, data interface{})

// ApprovalRequester is a callback that asks the user for permission before executing
// a destructive tool call. Implementations should block until the user responds (or
// the context is cancelled). It returns true to allow execution, false to skip.
// preview is an optional structured summary describing what would happen.
type ApprovalRequester func(ctx context.Context, toolCallID, toolName, riskLevel string, args map[string]interface{}, preview interface{}) bool

// RunAgent executes the agent loop: user message → LLM → tool calls → repeat.
// It takes existing conversation history and returns the updated history after the interaction.
// The context can be cancelled to stop the run; on cancel a summary assistant message is appended
// listing the tool calls that had already been executed.
// requestApproval may be nil; when non-nil it is consulted before running each destructive tool.
// The returned mutated flag is true if at least one tool call that modifies the project DB succeeded.
func RunAgent(ctx context.Context, projectPath string, history []Message, userMessage, attachmentContent, apiKey, provider, model string, emit EventEmitter, requestApproval ApprovalRequester) ([]Message, bool, error) {
	if apiKey == "" && provider != "ollama" {
		return history, false, fmt.Errorf("no API key configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// Build project context
	projectContext, err := BuildProjectContext(projectPath)
	if err != nil {
		projectContext = "Could not load project context: " + err.Error()
	}

	systemPrompt := buildSystemPrompt(projectContext)

	// Start from existing history or create new
	messages := make([]Message, 0, len(history)+2)
	messages = append(messages, Message{Role: "system", Content: systemPrompt})

	// Add prior conversation turns (skip old system messages)
	for _, m := range history {
		if m.Role != "system" {
			messages = append(messages, m)
		}
	}

	// Build user message content
	userContent := userMessage
	if attachmentContent != "" {
		userContent += "\n\n--- Attached Content ---\n" + attachmentContent
	}
	messages = append(messages, Message{Role: "user", Content: userContent})

	// Get tool definitions in OpenAI format
	tools := buildOpenAITools()

	// Track executed tool calls for cancellation summary.
	var executedTools []string
	// Track whether any successful tool call mutated the project DB.
	mutated := false

	// Agent loop
	for range maxIterations {
		if err := ctx.Err(); err != nil {
			return finalizeCancelled(messages, executedTools, emit), mutated, ErrCancelled
		}
		emit("agent-status", "Thinking...")

		response, err := callLLM(ctx, apiKey, provider, model, messages, tools)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return finalizeCancelled(messages, executedTools, emit), mutated, ErrCancelled
			}
			return stripSystemMessages(messages), mutated, err
		}

		choice := response.Choices[0]

		// If the LLM wants to call tools
		if len(choice.Message.ToolCalls) > 0 {
			// Add assistant message with tool calls
			messages = append(messages, Message{
				Role:      "assistant",
				Content:   choice.Message.Content,
				ToolCalls: choice.Message.ToolCalls,
			})

			// Execute each tool call
			for _, tc := range choice.Message.ToolCalls {
				if err := ctx.Err(); err != nil {
					return finalizeCancelled(messages, executedTools, emit), mutated, ErrCancelled
				}
				emit("agent-tool-start", map[string]string{
					"tool": tc.Function.Name,
					"id":   tc.ID,
				})

				// Parse arguments
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					args = map[string]interface{}{}
				}

				// Approval gate for destructive tools.
				// Fail-closed: if the caller didn't supply an approver, refuse to run
				// the tool rather than silently executing it. Callers that want
				// auto-approve must pass an explicit always-approve function.
				if IsDestructive(tc.Function.Name) {
					if requestApproval == nil {
						denied := ToolResult{Success: false, Error: "Destructive tool blocked: no approver configured for this agent run."}
						emit("agent-tool-result", map[string]interface{}{
							"tool":    tc.Function.Name,
							"id":      tc.ID,
							"success": false,
							"denied":  true,
						})
						messages = append(messages, Message{
							Role:       "tool",
							Content:    SerializeToolResult(denied),
							ToolCallID: tc.ID,
						})
						continue
					}
					preview := buildToolPreview(projectPath, tc.Function.Name, args)
					approved := requestApproval(ctx, tc.ID, tc.Function.Name, RiskDestructive, args, preview)
					if err := ctx.Err(); err != nil {
						return finalizeCancelled(messages, executedTools, emit), mutated, ErrCancelled
					}
					if !approved {
						denied := ToolResult{Success: false, Error: "User denied this action."}
						emit("agent-tool-result", map[string]interface{}{
							"tool":    tc.Function.Name,
							"id":      tc.ID,
							"success": false,
							"denied":  true,
						})
						messages = append(messages, Message{
							Role:       "tool",
							Content:    SerializeToolResult(denied),
							ToolCallID: tc.ID,
						})
						continue
					}
					// Re-verify that the targets the user saw in the preview still
					// match what we're about to do. Closes the TOCTOU window.
					if verifyErr := verifyToolPreview(projectPath, tc.Function.Name, args, preview); verifyErr != nil {
						aborted := ToolResult{Success: false, Error: verifyErr.Error()}
						emit("agent-tool-result", map[string]interface{}{
							"tool":    tc.Function.Name,
							"id":      tc.ID,
							"success": false,
							"denied":  true,
						})
						messages = append(messages, Message{
							Role:       "tool",
							Content:    SerializeToolResult(aborted),
							ToolCallID: tc.ID,
						})
						continue
					}
				}

				// Execute tool
				result := ExecuteTool(projectPath, tc.Function.Name, args)
				if result.Success {
					executedTools = append(executedTools, tc.Function.Name)
					if isMutatingTool(tc.Function.Name) {
						mutated = true
					}
				}

				emit("agent-tool-result", map[string]interface{}{
					"tool":    tc.Function.Name,
					"id":      tc.ID,
					"success": result.Success,
				})

				// Sync updated ignore list to frontend when modified
				if result.Success && (tc.Function.Name == "add_ignore_pattern" || tc.Function.Name == "remove_ignore_pattern") {
					if dataMap, ok := result.Data.(map[string]interface{}); ok {
						if updatedList, ok := dataMap["ignore_list"]; ok {
							emit("ignore-list-updated", updatedList)
						}
					}
				}

				// Forward reveal-in-browser navigation requests to the frontend
				if result.Success && tc.Function.Name == "reveal_in_browser" {
					emit("agent-reveal-in-browser", result.Data)
				}

				// Add tool result message
				messages = append(messages, Message{
					Role:       "tool",
					Content:    SerializeToolResult(result),
					ToolCallID: tc.ID,
				})
			}
			continue
		}

		// LLM returned a text response — we're done
		if choice.Message.Content != nil {
			content := fmt.Sprintf("%v", choice.Message.Content)
			messages = append(messages, Message{Role: "assistant", Content: content})
			emit("agent-response", content)
		}
		return stripSystemMessages(messages), mutated, nil
	}

	emit("agent-response", "I reached the maximum number of steps. Here's what I accomplished so far. You may need to continue with additional requests.")
	return stripSystemMessages(messages), mutated, nil
}

// stripSystemMessages removes system messages from history before persisting.
func stripSystemMessages(messages []Message) []Message {
	result := make([]Message, 0, len(messages))
	for _, m := range messages {
		if m.Role != "system" {
			result = append(result, m)
		}
	}
	return result
}

// finalizeCancelled appends a synthetic assistant message summarizing the cancelled run
// and emits agent-response so the frontend renders the summary in the chat.
func finalizeCancelled(messages []Message, executedTools []string, emit EventEmitter) []Message {
	var summary string
	if len(executedTools) == 0 {
		summary = "Cancelled. No tool calls had run yet."
	} else {
		// Group identical tool names with counts.
		counts := map[string]int{}
		order := []string{}
		for _, name := range executedTools {
			if _, ok := counts[name]; !ok {
				order = append(order, name)
			}
			counts[name]++
		}
		parts := make([]string, 0, len(order))
		for _, name := range order {
			if counts[name] > 1 {
				parts = append(parts, fmt.Sprintf("%s \u00d7%d", name, counts[name]))
			} else {
				parts = append(parts, name)
			}
		}
		summary = fmt.Sprintf("Cancelled. %d tool call(s) had already been applied: %s.", len(executedTools), strings.Join(parts, ", "))
	}
	messages = append(messages, Message{Role: "assistant", Content: summary})
	emit("agent-response", summary)
	return stripSystemMessages(messages)
}

// readOnlyTools lists tools that only query data and never modify the project.
var readOnlyTools = map[string]bool{
	"list_collections":          true,
	"list_assets_in_collection": true,
	"get_asset_details":         true,
	"list_users":                true,
	"list_statuses":             true,
	"list_task_types":           true,
	"list_tags":                 true,
	"list_templates":            true,
	"search_knowledge":          true,
	"get_my_permissions":        true,
	"get_user_activity":         true,
	"list_checkpoints":          true,
	"get_asset_tags":            true,
	"list_dependencies":         true,
	"list_dependency_types":     true,
	"list_collection_types":     true,
	"search_assets":             true,
	"get_project_summary":       true,
	"list_ignore_patterns":      true,
	"generate_script":           true,

	// Phase 1a read-only additions
	"list_workflows":       true,
	"list_roles":           true,
	"reveal_asset_on_disk": true,
	"reveal_in_browser":    true,
	"search_project_text":  true,

	// Phase 1a-ii read-only additions (collaborator listings)
	"list_project_collaborators": true,
	"list_studios":               true,
	"list_studio_users":          true,

	// DCC tools don't modify the project database
	"open_in_dcc":    true,
	"blender_render": true,
	"blender_export": true, "blender_run_script": true,
	"blender_run_python":   true,
	"blender_set_settings": true,
	"blender_link":         true,
	"run_terminal_command": true,
}

// isMutatingTool returns true if the tool modifies project data.
func isMutatingTool(toolName string) bool {
	return !readOnlyTools[toolName]
}

// buildSystemPrompt constructs the system prompt with project context and knowledge.
func buildSystemPrompt(projectContext string) string {
	var sb strings.Builder

	sb.WriteString(`You are Clustta Assistant, an AI that helps manage creative projects in Clustta — a version control and collaboration system for creative workflows (3D, VFX, animation, game dev).

## Your Capabilities
1. Answer questions about how Clustta works using the search_knowledge tool
2. Create, rename, delete, and manage collections and assets
3. Assign users, change statuses, move assets between collections
4. Generate scripts for batch file operations (rendering, conversion, exports)
5. Analyze attached content (like screenplays) and create project structures
6. Search and filter assets across the entire project by name, status, type, assignee, or tag
7. View checkpoint (version) history for any asset
8. Manage tags: create tags, add/remove tags on assets
9. Manage dependencies between assets: list, add, remove
10. Manage asset types and collection types: create, delete, list
11. Get project summaries with breakdowns by status, assignee, and type
12. Bulk assign, unassign, or randomly distribute assets among users
13. Remove users/collaborators from the project
14. Manage the ignore list: add, remove, and list patterns
15. Set up standard type presets for animation, game, VFX, or film pipelines
16. Open asset files in DCC applications (Blender, Maya, Houdini, etc.)
17. Launch Blender headless renders in a terminal window (fire-and-forget)
18. Export Blender files to FBX, OBJ, glTF, or USD via terminal
19. Run arbitrary commands in a visible terminal window
20. Run custom Python scripts on .blend files via terminal
21. Run inline Python code on .blend files (create collections, modify materials, rename objects, etc.)
22. Batch-modify Blender render settings (engine, resolution, FPS, samples, output format)
23. Link or append objects from dependency .blend files into a target .blend file (auto-resolves from Clustta dependency graph)

## Rules
- Messages may begin with a [Context: ...] block describing what the user currently has selected in the UI. Use this to resolve ambiguous references like "this asset", "these items", "the selected collection", etc. The context provides item names, IDs, types, and other metadata — use these IDs directly when calling tools.
- When the user asks for Blender-internal operations (creating Blender collections, modifying objects, changing materials, etc.), use blender_run_python to execute inline Python code directly — do not use generate_script.
- Blender Python best practices: when creating a Blender collection, always link it to the scene with bpy.context.scene.collection.children.link(). When creating objects, always link them to a collection. Data blocks not linked to the scene are invisible in the Outliner.
- Before any mutating operation (create, delete, rename, assign, etc.), FIRST call get_my_permissions to check the user's role. If the user lacks the required permission, tell them immediately — do not attempt the action.
- Use search_assets (with no filters) to list all assets directly. Do not assume assets must be inside collections — assets can exist at root level. search_assets returns paginated results (default 50). Use offset to page through large result sets.
- For destructive operations (delete, remove user), warn the user and ask for confirmation first
- For bulk operations (assign all X, change status of all Y), use bulk_assign or bulk_change_status with filter parameters — these operate server-side and do NOT require searching first. Never search+collect IDs+loop when a bulk tool with filters can do the job in one call.
- For script generation, display the script for user review — never claim to execute it
- Use exact IDs from the project data when calling tools — never guess IDs
- Be concise and direct in responses
- When the user asks about Clustta features, use search_knowledge to find accurate information
- If creating multiple items, use batch tools (batch_create_collections, batch_create_assets) instead of calling single-item tools repeatedly
- DCC tools (open_in_dcc, blender_render, blender_export) are fire-and-forget — they launch a terminal or process and return immediately. Inform the user the operation was started.
- For DCC tool detection: .blend files use Blender, .ma/.mb use Maya, .hip use Houdini. Users can also set BLENDER_PATH, MAYA_PATH, etc. environment variables.
- run_terminal_command launches any command in a visible terminal — use it for custom scripts or operations not covered by other tools
- blender_link auto-resolves source files from the target asset's Clustta dependency graph when source_asset_ids is omitted. Prefer this for linking dependent assets. Use data_names to link only specific named data blocks (e.g., the asset name) — without data_names, ALL data blocks of the specified types are linked from each source file.

`)

	sb.WriteString("## Current Project Context\n")
	sb.WriteString(projectContext)

	return sb.String()
}

// --- OpenAI API types ---

type openAIRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Tools    []openAITool `json:"tools,omitempty"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIMessage struct {
	Role      string      `json:"role"`
	Content   interface{} `json:"content"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// buildOpenAITools converts tool definitions to OpenAI's format.
func buildOpenAITools() []openAITool {
	defs := GetToolDefinitions()
	tools := make([]openAITool, 0, len(defs))
	for _, d := range defs {
		tools = append(tools, openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  d.Parameters,
			},
		})
	}
	return tools
}

// providerConfig holds the endpoint URL and default model for each LLM provider.
type providerConfig struct {
	URL   string
	Model string
}

var providers = map[string]providerConfig{
	"openai":    {URL: "https://api.openai.com/v1/chat/completions", Model: "gpt-4o-mini"},
	"anthropic": {URL: "https://api.anthropic.com/v1/messages", Model: "claude-sonnet-4-20250514"},
	"gemini":    {URL: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", Model: "gemini-2.5-flash"},
	"groq":      {URL: "https://api.groq.com/openai/v1/chat/completions", Model: "llama-3.3-70b-versatile"},
	"ollama":    {URL: "http://localhost:11434/v1/chat/completions", Model: "llama3.2"},
}

// providerModelOptions lists the user-selectable models per provider, shown in the
// console model dropdown. The first entry is treated as the default for that provider.
var providerModelOptions = map[string][]string{
	"openai": {
		"gpt-4o-mini",
		"gpt-4o",
		"gpt-4.1",
		"gpt-4.1-mini",
		"o3-mini",
	},
	"anthropic": {
		"claude-sonnet-4-20250514",
		"claude-opus-4-20250514",
		"claude-3-5-haiku-20241022",
	},
	"gemini": {
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-2.0-flash",
	},
	"groq": {
		"llama-3.3-70b-versatile",
		"llama-3.1-70b-versatile",
		"llama-3.1-8b-instant",
	},
	"ollama": {
		"llama3.2",
		"llama3.1",
		"qwen2.5",
		"mistral",
	},
}

// GetProviderModels returns the list of selectable models for the given provider.
// The first entry is the default. Returns an empty slice for unknown providers.
func GetProviderModels(provider string) []string {
	opts, ok := providerModelOptions[provider]
	if !ok {
		return nil
	}
	out := make([]string, len(opts))
	copy(out, opts)
	return out
}

// GetDefaultModel returns the default model for a provider, or empty string if unknown.
func GetDefaultModel(provider string) string {
	if cfg, ok := providers[provider]; ok {
		return cfg.Model
	}
	return ""
}

// VerifyOllamaModel checks that the local Ollama server is reachable and that
// the requested model is pulled. Returns a user-friendly error otherwise so
// the agent doesn't surface a generic 404 from the LLM call.
func VerifyOllamaModel(ctx context.Context, model string) error {
	if model == "" {
		model = GetDefaultModel("ollama")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := llmHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach Ollama at localhost:11434 (is it running?): %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama responded with status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("could not parse Ollama tag list: %w", err)
	}
	for _, m := range payload.Models {
		// Ollama returns names like "llama3.2:latest"; accept exact match or prefix-with-colon.
		if m.Name == model || strings.HasPrefix(m.Name, model+":") {
			return nil
		}
	}
	return fmt.Errorf("Ollama model %q is not pulled locally. Run `ollama pull %s` then try again", model, model)
}

// resolveModel returns model if non-empty, otherwise the provider default.
func resolveModel(provider, model string) string {
	if model != "" {
		return model
	}
	return GetDefaultModel(provider)
}

// callLLM calls the LLM API (OpenAI-compatible or Anthropic) and returns the response.
// If model is empty the provider's default is used.
func callLLM(ctx context.Context, apiKey, provider, model string, messages []Message, tools []openAITool) (openAIResponse, error) {
	switch provider {
	case "anthropic":
		return callAnthropic(ctx, apiKey, resolveModel(provider, model), messages, tools)
	default:
		cfg, ok := providers[provider]
		if !ok {
			cfg = providers["openai"]
			provider = "openai"
		}
		return callOpenAICompat(ctx, apiKey, cfg.URL, resolveModel(provider, model), messages, tools)
	}
}

// callOpenAICompat calls an OpenAI-compatible chat completions API.
func callOpenAICompat(ctx context.Context, apiKey, endpoint, model string, messages []Message, tools []openAITool) (openAIResponse, error) {
	reqBody := openAIRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return openAIResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return openAIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := llmHTTPClient.Do(req)
	if err != nil {
		return openAIResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openAIResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return openAIResponse{}, parseAPIError(resp.StatusCode, respBody, model)
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Include a snippet of the response body for debugging
		snippet := string(respBody)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return openAIResponse{}, fmt.Errorf("failed to parse LLM response: %w\nBody: %s", err, snippet)
	}

	if result.Error != nil {
		return openAIResponse{}, fmt.Errorf("LLM API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return openAIResponse{}, fmt.Errorf("LLM returned no choices")
	}

	return result, nil
}

// parseAPIError extracts a user-friendly message from an LLM API error response.
func parseAPIError(statusCode int, body []byte, model string) error {
	// Try to extract the message from JSON error response
	var parsed struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	msg := ""
	if json.Unmarshal(body, &parsed) == nil && parsed.Error.Message != "" {
		msg = parsed.Error.Message
	}

	switch statusCode {
	case 429:
		return fmt.Errorf("Rate limit exceeded for %s. Please wait a moment and try again.", model)
	case 401:
		return fmt.Errorf("Invalid API key. Please check your LLM provider settings.")
	case 403:
		return fmt.Errorf("Access denied. Your API key may not have permission for %s.", model)
	case 500, 502, 503:
		return fmt.Errorf("The LLM service (%s) is temporarily unavailable. Please try again later.", model)
	default:
		if msg != "" {
			// Truncate long messages
			if len(msg) > 200 {
				msg = msg[:200] + "..."
			}
			return fmt.Errorf("LLM API error (%d): %s", statusCode, msg)
		}
		return fmt.Errorf("LLM API returned an unexpected error (status %d). Please try again.", statusCode)
	}
}

// callAnthropic calls the Anthropic messages API, translating to/from OpenAI format.
func callAnthropic(ctx context.Context, apiKey, model string, messages []Message, tools []openAITool) (openAIResponse, error) {
	// Extract system message
	var systemPrompt string
	var anthropicMessages []anthropicMessage

	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = fmt.Sprintf("%v", m.Content)
			continue
		}

		msg := anthropicMessage{Role: m.Role}

		if m.Role == "tool" {
			msg.Role = "user"
			msg.Content = []anthropicContent{{
				Type:        "tool_result",
				ToolUseID:   m.ToolCallID,
				ToolContent: fmt.Sprintf("%v", m.Content),
			}}
		} else if len(m.ToolCalls) > 0 {
			contents := []anthropicContent{}
			if m.Content != nil && fmt.Sprintf("%v", m.Content) != "" {
				contents = append(contents, anthropicContent{
					Type: "text",
					Text: fmt.Sprintf("%v", m.Content),
				})
			}
			for _, tc := range m.ToolCalls {
				var input map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
					input = map[string]interface{}{}
				}
				contents = append(contents, anthropicContent{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: input,
				})
			}
			msg.Content = contents
		} else {
			msg.Content = fmt.Sprintf("%v", m.Content)
		}

		anthropicMessages = append(anthropicMessages, msg)
	}

	// Convert tools
	anthropicTools := make([]anthropicTool, 0, len(tools))
	for _, t := range tools {
		anthropicTools = append(anthropicTools, anthropicTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		})
	}

	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return openAIResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return openAIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := llmHTTPClient.Do(req)
	if err != nil {
		return openAIResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openAIResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return openAIResponse{}, parseAPIError(resp.StatusCode, respBody, model)
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return openAIResponse{}, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	if anthropicResp.Error != nil {
		return openAIResponse{}, fmt.Errorf("Anthropic API error: %s", anthropicResp.Error.Message)
	}

	// Convert back to OpenAI format
	return convertAnthropicResponse(anthropicResp), nil
}

// --- Anthropic API types ---

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type anthropicContent struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	ToolUseID   string                 `json:"tool_use_id,omitempty"`
	ToolContent interface{}            `json:"content,omitempty"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	Content    []anthropicContent `json:"content"`
	StopReason string             `json:"stop_reason"`
	Error      *anthropicError    `json:"error,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// convertAnthropicResponse maps Anthropic's response to OpenAI's format.
func convertAnthropicResponse(resp anthropicResponse) openAIResponse {
	var textContent string
	var toolCalls []ToolCall

	for _, c := range resp.Content {
		switch c.Type {
		case "text":
			textContent += c.Text
		case "tool_use":
			argsJSON, _ := json.Marshal(c.Input)
			toolCalls = append(toolCalls, ToolCall{
				ID:   c.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      c.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	var content interface{} = textContent
	if textContent == "" && len(toolCalls) > 0 {
		content = nil
	}

	return openAIResponse{
		Choices: []openAIChoice{{
			Message: openAIMessage{
				Role:      "assistant",
				Content:   content,
				ToolCalls: toolCalls,
			},
		}},
	}
}

// ReadAttachment reads and returns the text content of an attached file.
// Supports plain text files and PDFs.
func ReadAttachment(filePath string) (string, error) {
	if filePath == "" {
		return "", nil
	}

	// Handle PDF files
	if strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return readPDFAttachment(filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read attachment: %w", err)
	}

	// Limit to ~50KB to avoid overwhelming the context
	content := string(data)
	if len(content) > 50000 {
		content = content[:50000] + "\n\n[Content truncated — file too large]"
	}

	return content, nil
}

// readPDFAttachment extracts text content from a PDF file.
func readPDFAttachment(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPages := r.NumPage()
	for i := 1; i <= totalPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		if sb.Len() > 50000 {
			sb.WriteString("\n\n[Content truncated — PDF too large]")
			break
		}
	}

	content := sb.String()
	if content == "" {
		return "", fmt.Errorf("could not extract text from PDF — it may be image-based or scanned")
	}

	return content, nil
}
