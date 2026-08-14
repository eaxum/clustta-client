package agent

import (
	"bytes"
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/agent/planning"
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

const (
	maxIterations       = 25
	openAIModelTerra    = "gpt-5.6-terra"
	openAIModelSol      = "gpt-5.6-sol"
	openAIModelLuna     = "gpt-5.6-luna"
	reasoningEffortNone = "none"
)

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

// ApprovalRequester asks the user to approve a destructive tool call and blocks until they respond.
// It returns true to allow execution, false to skip; preview is an optional structured summary.
type ApprovalRequester func(ctx context.Context, toolCallID, toolName, riskLevel string, args map[string]interface{}, preview interface{}) bool

// RunAgent executes the agent loop: user message â†’ LLM â†’ tool calls â†’ repeat,
// returning the updated history, whether any mutating tool succeeded, and any error.
func RunAgent(ctx context.Context, projectPath string, history []Message, userMessage, attachmentContent, apiKey, provider, model string, emit EventEmitter, requestApproval ApprovalRequester) ([]Message, bool, error) {
	if apiKey == "" && provider != "ollama" {
		return history, false, fmt.Errorf("no API key configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	projectContext, err := BuildProjectContext(projectPath)
	if err != nil {
		projectContext = "Could not load project context: " + err.Error()
	}

	systemPrompt := buildSystemPrompt(projectContext)

	messages := make([]Message, 0, len(history)+2)
	messages = append(messages, Message{Role: "system", Content: systemPrompt})

	for _, m := range history {
		if m.Role != "system" {
			messages = append(messages, m)
		}
	}

	userContent := userMessage
	authoritativeContext := parseTurnContext(userMessage)
	if attachmentContent != "" {
		userContent += "\n\n--- Attached Content ---\n" + attachmentContent
	}
	messages = append(messages, Message{Role: "user", Content: userContent})

	tools := buildOpenAITools()

	var executedTools []string
	mutated := false

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

		if len(choice.Message.ToolCalls) > 0 {
			messages = append(messages, Message{
				Role:      "assistant",
				Content:   choice.Message.Content,
				ToolCalls: choice.Message.ToolCalls,
			})

			for _, tc := range choice.Message.ToolCalls {
				if err := ctx.Err(); err != nil {
					return finalizeCancelled(messages, executedTools, emit), mutated, ErrCancelled
				}
				emit("agent-tool-start", map[string]string{
					"tool": tc.Function.Name,
					"id":   tc.ID,
				})

				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					args = map[string]interface{}{}
				}
				applyAuthoritativeScope(args, authoritativeContext)

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
					if preview.Blocked {
						reason := "No matching entities were found in the requested scope."
						for _, note := range preview.Notes {
							if note != "" && !strings.Contains(strings.ToLower(note), "manual sync") {
								reason = note
								break
							}
						}
						blocked := ToolResult{Success: false, Error: reason}
						emit("agent-tool-result", map[string]interface{}{
							"tool": tc.Function.Name, "id": tc.ID,
							"success": false, "blocked": true, "error": reason,
						})
						messages = append(messages, Message{
							Role: "tool", Content: SerializeToolResult(blocked), ToolCallID: tc.ID,
						})
						continue
					}
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

				emit("agent-command-progress", map[string]interface{}{
					"tool": tc.Function.Name, "id": tc.ID, "current": 0, "total": 1,
					"percentage": 0, "message": "Applying approved changes locally",
				})
				result := ExecuteToolContext(ctx, projectPath, tc.Function.Name, args)
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
					"data":    result.Data,
				})
				if def, ok := agentcommands.DefinitionFor(tc.Function.Name); ok && result.Success && def.Direct == nil {
					emit("agent-command-progress", map[string]interface{}{
						"tool": tc.Function.Name, "id": tc.ID, "current": 1, "total": 1,
						"percentage": 100, "message": "Local changes complete",
					})
					if planned, ok := result.Data.(planning.Result); ok && planned.RequiresSync {
						emit("agent-local-changes", map[string]interface{}{
							"tool": tc.Function.Name, "plan_id": planned.PlanID,
							"applied": planned.Applied, "requires_sync": true,
						})
					}
				}

				if result.Success && (tc.Function.Name == "add_ignore_pattern" || tc.Function.Name == "remove_ignore_pattern") {
					if dataMap, ok := result.Data.(map[string]interface{}); ok {
						if updatedList, ok := dataMap["ignore_list"]; ok {
							emit("ignore-list-updated", updatedList)
						}
					}
				}

				if result.Success && tc.Function.Name == "reveal_in_browser" {
					emit("agent-reveal-in-browser", result.Data)
				}

				if result.Success && tc.Function.Name == "apply_browser_filter" {
					emit("agent-apply-filter", result.Data)
				}
				if result.Success && tc.Function.Name == "clear_browser_filter" {
					emit("agent-clear-filter", result.Data)
				}

				messages = append(messages, Message{
					Role:       "tool",
					Content:    SerializeToolResult(result),
					ToolCallID: tc.ID,
				})
			}
			continue
		}

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

	"list_workflows":       true,
	"list_roles":           true,
	"reveal_asset_on_disk": true,
	"reveal_in_browser":    true,
	"search_project_text":  true,

	"list_project_collaborators": true,
	"list_studios":               true,
	"list_studio_users":          true,

	"list_filter_dimensions": true,
	"apply_browser_filter":   true,
	"clear_browser_filter":   true,
	"query_entities":         true,
	"audit_entities":         true,
	"dcc_open":               true,

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

	sb.WriteString(`You are Clustta Assistant, an AI that helps manage creative projects in Clustta â€” a version control and collaboration system for creative workflows (3D, VFX, animation, game dev).

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
15. Set up standard type presets for animation, game, VFX, or film pipelines; scaffold a full animation project (types + Production/Assets tree + EP/SEQ/SH hierarchy) in one call
16. Open asset files in DCC applications (Blender, Maya, Houdini, etc.)
17. Launch scoped Blender headless jobs (fire-and-forget)
18. Export scoped Blender files to FBX, OBJ, glTF, or USD
19. Run approved scoped Python scripts on .blend files
21. Run inline Python code on .blend files (create collections, modify materials, rename objects, etc.)
22. Batch-modify Blender render settings (engine, resolution, FPS, samples, output format)
23. Link or append objects from dependency .blend files into a target .blend file (auto-resolves from Clustta dependency graph)

## Rules
- Messages may begin with a [Context: ...] block describing what the user currently has selected in the UI. Use this to resolve ambiguous references like "this asset", "these items", "the selected collection", etc. The context provides item names, IDs, types, and other metadata â€” use these IDs directly when calling tools.
- Messages may instead begin with [Context JSON] followed by structured current_location and selection data. Prefer this structure when building a command scope:
  - selected items: source="selection" and copy the selected entity envelopes into scope.selection
  - "here": copy context.here_scope and only change recursive when the user explicitly requests recursion
  - one explicit entity: source="entity"
  - whole project: source="project"
- For rename, move, delete, status, type, assignment, tags, dependencies, and task/resource requests, use the batch_* commands. Do not loop legacy single-item tools.
- batch_* commands resolve scope, compute a deterministic local plan, show an approval preview, revalidate it, and apply locally. The user must manually sync afterward.
- For batch_add_dependency and batch_remove_dependency, target_scope is the one asset receiving or losing dependencies. Use source="entity" with the returned ID for a named target such as "Zeus". Use source="selection" only when exactly one selected asset is the target. Never use source="here" or source="project" for target_scope.
- For dependency prompts such as "make all assets here dependencies of Zeus", use context.here_scope for dependency_scope, use Zeus as target_scope, and let the command exclude Zeus from its own dependencies.
- Agent-created dependencies always use the project's "linked" dependency type. Do not call list_dependency_types before batch_add_dependency and never choose blocking, working, waiting, or another dependency type.
- Use entity type values exactly as provided: asset, collection, untracked_asset, untracked_collection.
- Status changes only support asset. Type changes support asset and collection. Assignment supports tracked asset and collection with their different semantics.
- When asset types should match asset-name suffixes, call batch_change_type once with asset_type_rules containing every suffix-to-type-ID mapping. Never emit a separate batch_change_type call for each suffix.
- For batch renames, use format, prepend_text, append_text, find_text/replace_text, remove_prefix, remove_suffix, template, numbering, or name_mappings as requested. Composable rules are applied in that order; name_mappings and a single explicit new_name are exclusive rules.
- Use numbering for sequential names. It supports start, step, padding, prefix/suffix position, and separator. A template can include {name} and {number}.

## Citing entities in your replies
- Whenever you mention a specific asset, collection, or user that exists in the project, write it inline using this exact token format so the UI can render it as an interactive chip:
  - Asset:      [[asset:<id>|Display Name]]
  - Collection: [[collection:<id>|Display Name]]
  - Untracked asset: [[untracked_asset:<id>|Display Name]]
  - Untracked collection: [[untracked_collection:<id>|Display Name]]
  - User:       [[user:<id>|Display Name]]
- Use the id returned by the most recent tool call. Never invent ids. If you do not have an id, write the name as plain text instead.
- Never write a raw id, UUID, or hash in your reply. The user must never see ids â€” they only ever see the display name inside the chip.
- The token replaces the name inline. Example: "I assigned [[asset:9f3...|shot_010_layout]] to [[user:b21...|Ada Lovelace]]." Do NOT also write the name outside the token.
- Never wrap a token in Markdown emphasis or code formatting. Do NOT write **[[asset:...|...]]**, *[[...]]*, _[[...]]_, or place tokens inside backticks. Plain tokens only â€” the UI styles the chip.
- In a response, render at most three entity chips. Summarize any remainder as "+N more"; never enumerate a large tool result.
- Short lists work the same way â€” render each item as a chip on its own line, e.g.
  - [[asset:<id>|Name]]
  - [[asset:<id>|Name]]
- When the user asks for Blender-internal operations, use the scoped dcc_run_python command.
- Blender Python best practices: when creating a Blender collection, always link it to the scene with bpy.context.scene.collection.children.link(). When creating objects, always link them to a collection. Data blocks not linked to the scene are invisible in the Outliner.
- Before any mutating operation (create, delete, rename, assign, etc.), FIRST call get_my_permissions to check the user's role. If the user lacks the required permission, tell them immediately â€” do not attempt the action.
- Use search_assets (with no filters) to list all assets directly. Do not assume assets must be inside collections â€” assets can exist at root level. search_assets returns paginated results (default 50). Use offset to page through large result sets.
- For destructive operations (delete, remove user), warn the user and ask for confirmation first
- For bulk operations, express the target through the command's structured scope and filters. Do not search, collect IDs, and loop.
- For listing assets or collections in a location, use query_entities with context.here_scope. It supports tracked and untracked entities; do not use legacy collection listing tools.
- For any count, existence, or contents question about the current location, call query_entities before answering. Never infer that an untracked collection is empty merely because it has no database row.
- For script generation, display the script for user review â€” never claim to execute it
- Use exact IDs from the project data when calling tools â€” never guess IDs
- Be concise and direct in responses
- When the user asks about Clustta features, use search_knowledge to find accurate information
- If creating multiple items, use batch tools (batch_create_collections, batch_create_assets) instead of calling single-item tools repeatedly.
- Use batch_distribute to distribute a structured asset scope across several users. Do not use legacy random_assign or enumerate IDs.
- DCC jobs are fire-and-forget. Inform the user when the scoped job was started.
- For DCC tool detection: .blend files use Blender, .ma/.mb use Maya, .hip use Houdini. Users can also set BLENDER_PATH, MAYA_PATH, etc. environment variables.
- dcc_link_dependencies auto-resolves source files from the target asset's dependency graph when source_scope is omitted.

## Filtering the browser view
- When the user asks to view, list, or filter assets/collections (e.g. "show all rigging assets", "list done tasks", "tasks assigned to me", "only modified files", "all .blend files"), call apply_browser_filter â€” do NOT use search_assets just to display things in the browser. search_assets is for analytical queries that return data to you; apply_browser_filter changes what the user sees.
- If you do not already know the available statuses / asset types / users / tags, call list_filter_dimensions FIRST so your filter terms match real values.
- Use the literal string "@me" inside assignees for the current user.
- Use no_assignees:true for "unassigned".
- Set deep:true when the user says things like "across the project", "everywhere", or "the whole project".
- Call clear_browser_filter to undo all filtering.
- Filtering does not modify any project data and never requires confirmation.
- The result of apply_browser_filter includes match_count and unmatched. If match_count is 0, tell the user no items matched (and mention any unmatched terms) instead of pretending the filter succeeded. If unmatched is non-empty but match_count > 0, mention which terms were ignored.

## Setting up an animation project
- When the user asks to set up or scaffold an animation project (often with a script/screenplay attached), call setup_animation_production. Do NOT loop create_collection, batch_create_collections, or create_asset.
- First read any attached script/screenplay with read_attachment, then derive episodes, sequences, shot counts, and the character/environment/prop lists from it. Pass them in ONE call.
- ALWAYS call list_templates FIRST so you know the template names available. Then pass the chosen one in template (e.g. template:"blender"). Only omit template if list_templates returns exactly one item.
- The tool creates BOTH task files inside every shot AND task files inside every library entry (character/env/prop). Defaults: shot_tasks=["Animation","Lighting","FX"], asset_tasks=["Model","Rig","Texture"]. Do NOT override unless the user asked for different tasks. If the user asks to skip task files, pass shot_tasks:[] or asset_tasks:[]. Never omit these arrays expecting "no tasks" â€” missing arrays mean "use defaults".
- Naming is fixed and handled by the tool: EP### (3-digit), SEQ### (3-digit, step 10 â€” SEQ010, SEQ020, â€¦), SH#### (4-digit, step 10 â€” SH0010, SH0020, â€¦). Do not pre-format names.
- For non-series projects pass is_series:false (the EP layer is skipped and sequences live directly under Production/).
- The call runs in a single transaction. If it fails, the project is left untouched â€” fix the arguments and call setup_animation_production again. Do NOT fall back to manual create_collection/create_asset loops to "patch up" perceived gaps; doing so produces malformed structures.
- setup_animation_production also creates the standard animation asset and collection types, so you do not need to call setup_project_types separately.
- After the scaffold, use the regular create_collection / create_asset tools for ad-hoc additions. For new shots, pick the next free multiple of 10 after the highest existing shot in that sequence; reserve letter suffixes (SH0010A, B, â€¦) for takes and variations of an existing shot.
- When you reply about the scaffold result, render Production, Assets, and each library bucket as collection chips using the IDs returned by setup_animation_production: production_id, assets_id, and bucket_ids["Characters"|"Environments"|"Props"]. Example: "I created [[collection:<production_id>|Production]] and [[collection:<assets_id>|Assets]] with [[collection:<bucket_ids.Characters>|Characters]], [[collection:<bucket_ids.Environments>|Environments]], and [[collection:<bucket_ids.Props>|Props]]." Never write raw paths like "Production/" or "Assets/Characters/" in plain text â€” always use chips so the user can click through.

`)

	sb.WriteString("## Current Project Context\n")
	sb.WriteString(projectContext)

	return sb.String()
}

// --- OpenAI API types ---

type openAIRequest struct {
	Model           string       `json:"model"`
	Messages        []Message    `json:"messages"`
	Tools           []openAITool `json:"tools,omitempty"`
	ReasoningEffort string       `json:"reasoning_effort,omitempty"`
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
	"openai":    {URL: "https://api.openai.com/v1/chat/completions", Model: openAIModelTerra},
	"anthropic": {URL: "https://api.anthropic.com/v1/messages", Model: "claude-sonnet-4-20250514"},
	"gemini":    {URL: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", Model: "gemini-2.5-flash"},
	"groq":      {URL: "https://api.groq.com/openai/v1/chat/completions", Model: "llama-3.3-70b-versatile"},
	"ollama":    {URL: "http://localhost:11434/v1/chat/completions", Model: "llama3.2"},
}

// providerModelOptions lists the user-selectable models per provider, shown in the
// console model dropdown. The first entry is treated as the default for that provider.
var providerModelOptions = map[string][]string{
	"openai": {
		openAIModelTerra,
		openAIModelSol,
		openAIModelLuna,
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
// the requested model is pulled, returning a user-friendly error otherwise.
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
		if m.Name == model || strings.HasPrefix(m.Name, model+":") {
			return nil
		}
	}
	return fmt.Errorf("Ollama model %q is not pulled locally. Run `ollama pull %s` then try again", model, model)
}

// ResolveProviderModel returns a supported model or the provider default.
func ResolveProviderModel(provider, model string) string {
	if model != "" {
		if provider != "openai" {
			return model
		}
		for _, availableModel := range providerModelOptions[provider] {
			if model == availableModel {
				return model
			}
		}
	}
	return GetDefaultModel(provider)
}

// callLLM dispatches to the right provider's API and returns the response.
// Falls back to the provider default when model is empty.
func callLLM(ctx context.Context, apiKey, provider, model string, messages []Message, tools []openAITool) (openAIResponse, error) {
	switch provider {
	case "anthropic":
		return callAnthropic(ctx, apiKey, ResolveProviderModel(provider, model), messages, tools)
	default:
		cfg, ok := providers[provider]
		if !ok {
			cfg = providers["openai"]
			provider = "openai"
		}
		return callOpenAICompat(ctx, apiKey, cfg.URL, ResolveProviderModel(provider, model), messages, tools)
	}
}

// callOpenAICompat calls an OpenAI-compatible chat completions API.
func callOpenAICompat(ctx context.Context, apiKey, endpoint, model string, messages []Message, tools []openAITool) (openAIResponse, error) {
	reqBody := buildOpenAIRequest(model, messages, tools)

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

func buildOpenAIRequest(model string, messages []Message, tools []openAITool) openAIRequest {
	return openAIRequest{
		Model:           model,
		Messages:        messages,
		Tools:           tools,
		ReasoningEffort: reasoningEffortForModel(model),
	}
}

func reasoningEffortForModel(model string) string {
	switch model {
	case openAIModelTerra, openAIModelSol, openAIModelLuna:
		return reasoningEffortNone
	default:
		return ""
	}
}

// parseAPIError extracts a user-friendly message from an LLM API error response.
func parseAPIError(statusCode int, body []byte, model string) error {
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

// ReadAttachment returns the text content of an attached file, supporting
// plain text files and PDFs.
func ReadAttachment(filePath string) (string, error) {
	if filePath == "" {
		return "", nil
	}

	if strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return readPDFAttachment(filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read attachment: %w", err)
	}

	content := string(data)
	if len(content) > 50000 {
		content = content[:50000] + "\n\n[Content truncated â€” file too large]"
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
			sb.WriteString("\n\n[Content truncated â€” PDF too large]")
			break
		}
	}

	content := sb.String()
	if content == "" {
		return "", fmt.Errorf("could not extract text from PDF â€” it may be image-based or scanned")
	}

	return content, nil
}
