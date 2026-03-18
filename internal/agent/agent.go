package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

const maxIterations = 25

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

// RunAgent executes the agent loop: user message → LLM → tool calls → repeat.
// It takes existing conversation history and returns the updated history after the interaction.
func RunAgent(projectPath string, history []Message, userMessage, attachmentContent, apiKey, provider string, emit EventEmitter) ([]Message, error) {
	if apiKey == "" && provider != "ollama" {
		return history, fmt.Errorf("no API key configured")
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

	// Agent loop
	for range maxIterations {
		emit("agent-status", "Thinking...")

		response, err := callLLM(apiKey, provider, messages, tools)
		if err != nil {
			return stripSystemMessages(messages), fmt.Errorf("LLM API error: %w", err)
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
				emit("agent-tool-start", map[string]string{
					"tool": tc.Function.Name,
					"id":   tc.ID,
				})

				// Parse arguments
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					args = map[string]interface{}{}
				}

				// Execute tool
				result := ExecuteTool(projectPath, tc.Function.Name, args)

				emit("agent-tool-result", map[string]interface{}{
					"tool":    tc.Function.Name,
					"id":      tc.ID,
					"success": result.Success,
				})

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
		return stripSystemMessages(messages), nil
	}

	emit("agent-response", "I reached the maximum number of steps. Here's what I accomplished so far. You may need to continue with additional requests.")
	return stripSystemMessages(messages), nil
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

## Rules
- Always use tools to look up current project data before making changes (list collections, statuses, users, etc.)
- For destructive operations (delete), warn the user and ask for confirmation first
- For bulk operations, present a summary plan before executing
- For script generation, display the script for user review — never claim to execute it
- Use exact IDs from the project data when calling tools — never guess IDs
- Be concise and direct in responses
- When the user asks about Clustta features, use search_knowledge to find accurate information
- If creating multiple items, execute the tool calls one by one and report progress

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

// providerConfig holds the endpoint URL and model for each LLM provider.
type providerConfig struct {
	URL   string
	Model string
}

var providers = map[string]providerConfig{
	"openai": {URL: "https://api.openai.com/v1/chat/completions", Model: "gpt-4o-mini"},
	"gemini": {URL: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", Model: "gemini-2.5-flash"},
	"groq":   {URL: "https://api.groq.com/openai/v1/chat/completions", Model: "llama-3.3-70b-versatile"},
	"ollama": {URL: "http://localhost:11434/v1/chat/completions", Model: "llama3.2"},
}

// callLLM calls the LLM API (OpenAI-compatible or Anthropic) and returns the response.
func callLLM(apiKey, provider string, messages []Message, tools []openAITool) (openAIResponse, error) {
	switch provider {
	case "anthropic":
		return callAnthropic(apiKey, messages, tools)
	default:
		cfg, ok := providers[provider]
		if !ok {
			cfg = providers["openai"]
		}
		return callOpenAICompat(apiKey, cfg.URL, cfg.Model, messages, tools)
	}
}

// callOpenAICompat calls an OpenAI-compatible chat completions API.
func callOpenAICompat(apiKey, endpoint, model string, messages []Message, tools []openAITool) (openAIResponse, error) {
	reqBody := openAIRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return openAIResponse{}, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return openAIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return openAIResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openAIResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return openAIResponse{}, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(respBody))
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

// callAnthropic calls the Anthropic messages API, translating to/from OpenAI format.
func callAnthropic(apiKey string, messages []Message, tools []openAITool) (openAIResponse, error) {
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
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return openAIResponse{}, err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return openAIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return openAIResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openAIResponse{}, err
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
