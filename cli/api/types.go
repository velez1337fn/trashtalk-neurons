package api

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type FunctionParameter struct {
	Type       string            `json:"type"`
	Properties map[string]PropertySchema `json:"properties"`
	Required   []string          `json:"required,omitempty"`
}

type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Function struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  FunctionParameter `json:"parameters"`
}

type Tool struct {
	Type     string `json:"type"`
	Function Function `json:"function"`
}

type ChatCompletionRequest struct {
	Model    string     `json:"model"`
	Messages []Message  `json:"messages"`
	Tools    []Tool     `json:"tools,omitempty"`
	Stream   bool       `json:"stream,omitempty"`
}

type AnthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

type ThinkingRequest struct {
	Model    string             `json:"model"`
	Messages []Message          `json:"messages"`
	Tools    []Tool             `json:"tools,omitempty"`
	Thinking *AnthropicThinking `json:"thinking,omitempty"`
	Stream   bool               `json:"stream,omitempty"`
}

type Choice struct {
	Index        int          `json:"index"`
	Message      Message      `json:"message"`
	FinishReason string       `json:"finish_reason"`
	Delta        Delta        `json:"delta,omitempty"`
	ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
}

type Delta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Index    int        `json:"index"`
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type StreamChunk struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage,omitempty"`
}
