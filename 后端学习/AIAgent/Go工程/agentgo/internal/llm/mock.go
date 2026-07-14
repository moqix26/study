package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var expressionPattern = regexp.MustCompile(`[-+*/().0-9\s]{3,}`)

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Generate(ctx context.Context, request Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}

	for _, item := range request.Input {
		if item["type"] == "function_call_output" {
			output, _ := item["output"].(string)
			return Response{
				ID:   "mock-response-final",
				Text: "Mock Agent 已使用工具结果：" + output,
			}, nil
		}
	}

	message := LastUserText(request.Input)
	if len(request.Tools) > 0 {
		lower := strings.ToLower(message)
		if strings.Contains(message, "计算") || strings.Contains(lower, "calculate") {
			expression := expressionPattern.FindString(message)
			if strings.TrimSpace(expression) == "" {
				expression = "1+1"
			}
			arguments, _ := json.Marshal(map[string]string{"expression": strings.TrimSpace(expression)})
			return Response{
				ID: "mock-response-tool",
				ToolCalls: []ToolCall{{
					CallID:    "mock-call-calculator",
					Name:      "calculator",
					Arguments: arguments,
				}},
			}, nil
		}
		if strings.Contains(message, "知识库") || strings.Contains(lower, "knowledge") {
			arguments, _ := json.Marshal(map[string]string{"query": message})
			return Response{
				ID: "mock-response-rag-tool",
				ToolCalls: []ToolCall{{
					CallID:    "mock-call-search",
					Name:      "search_knowledge",
					Arguments: arguments,
				}},
			}, nil
		}
	}

	if message == "" {
		message = "空消息"
	}
	return Response{ID: "mock-response-text", Text: fmt.Sprintf("Mock 回答：%s", message)}, nil
}

func (m *MockClient) Stream(ctx context.Context, request Request) (<-chan StreamEvent, error) {
	response, err := m.Generate(ctx, request)
	if err != nil {
		return nil, err
	}
	events := make(chan StreamEvent)
	go func() {
		defer close(events)
		for _, part := range strings.Fields(response.Text) {
			select {
			case <-ctx.Done():
				events <- StreamEvent{Type: "error", Err: ctx.Err()}
				return
			case events <- StreamEvent{Type: "response.output_text.delta", Delta: part + " ", ResponseID: response.ID}:
			}
		}
		events <- StreamEvent{Type: "response.completed", ResponseID: response.ID}
	}()
	return events, nil
}
