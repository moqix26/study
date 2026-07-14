package llm

import (
	"context"
	"encoding/json"
)

type InputItem map[string]any

func UserText(text string) InputItem {
	return InputItem{"role": "user", "content": text}
}

func AssistantText(text string) InputItem {
	return InputItem{"role": "assistant", "content": text}
}

func FunctionOutput(callID, output string) InputItem {
	return InputItem{
		"type":    "function_call_output",
		"call_id": callID,
		"output":  output,
	}
}

type ToolDefinition struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
	Strict      bool           `json:"strict"`
}

type ToolCall struct {
	CallID    string          `json:"call_id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type Request struct {
	Model              string
	Input              []InputItem
	Tools              []ToolDefinition
	PreviousResponseID string
	Store              bool
}

type Response struct {
	ID          string
	Status      string
	Text        string
	Refusal     string
	ToolCalls   []ToolCall
	OutputItems []InputItem
}

type StreamEvent struct {
	Type       string
	Delta      string
	ResponseID string
	Err        error
}

type Client interface {
	Generate(ctx context.Context, request Request) (Response, error)
	Stream(ctx context.Context, request Request) (<-chan StreamEvent, error)
}

func LastUserText(items []InputItem) string {
	for i := len(items) - 1; i >= 0; i-- {
		role, _ := items[i]["role"].(string)
		content, _ := items[i]["content"].(string)
		if role == "user" {
			return content
		}
	}
	return ""
}
