package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIResponsesClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIResponsesClient(apiKey, baseURL, model string, timeout time.Duration) *OpenAIResponsesClient {
	return &OpenAIResponsesClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: timeout},
	}
}

type responsesRequest struct {
	Model              string           `json:"model"`
	Input              []InputItem      `json:"input"`
	Tools              []ToolDefinition `json:"tools,omitempty"`
	Include            []string         `json:"include,omitempty"`
	PreviousResponseID string           `json:"previous_response_id,omitempty"`
	Store              bool             `json:"store"`
	Stream             bool             `json:"stream,omitempty"`
}

type responsesOutputItem struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Content   []struct {
		Type    string `json:"type"`
		Text    string `json:"text"`
		Refusal string `json:"refusal"`
	} `json:"content"`
}

type responsesAPIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
	Param   string `json:"param"`
}

type responsesIncompleteDetails struct {
	Reason string `json:"reason"`
}

type responsesResponse struct {
	ID                string                      `json:"id"`
	Status            string                      `json:"status"`
	Output            []json.RawMessage           `json:"output"`
	Error             *responsesAPIError          `json:"error,omitempty"`
	IncompleteDetails *responsesIncompleteDetails `json:"incomplete_details,omitempty"`
}

func (c *OpenAIResponsesClient) Generate(ctx context.Context, request Request) (Response, error) {
	payload := c.payload(request, false)
	httpResponse, err := c.do(ctx, payload)
	if err != nil {
		return Response{}, err
	}
	defer httpResponse.Body.Close()

	if err := checkStatus(httpResponse); err != nil {
		return Response{}, err
	}
	var decoded responsesResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&decoded); err != nil {
		return Response{}, fmt.Errorf("解析 Responses API 响应失败: %w", err)
	}
	if err := responseStatusError(decoded); err != nil {
		return Response{}, err
	}

	result := Response{ID: decoded.ID, Status: decoded.Status}
	for _, rawItem := range decoded.Output {
		var preserved InputItem
		if err := json.Unmarshal(rawItem, &preserved); err != nil {
			return Response{}, fmt.Errorf("解析 Responses API output item 失败: %w", err)
		}
		result.OutputItems = append(result.OutputItems, preserved)

		var item responsesOutputItem
		if err := json.Unmarshal(rawItem, &item); err != nil {
			return Response{}, fmt.Errorf("解析 Responses API output item 失败: %w", err)
		}
		switch item.Type {
		case "message":
			for _, content := range item.Content {
				switch content.Type {
				case "output_text":
					result.Text += content.Text
				case "refusal":
					result.Refusal += content.Refusal
				}
			}
		case "function_call":
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				CallID:    item.CallID,
				Name:      item.Name,
				Arguments: json.RawMessage(item.Arguments),
			})
		}
	}
	if strings.TrimSpace(result.Text) == "" && result.Refusal != "" {
		result.Text = result.Refusal
	}
	return result, nil
}

func (c *OpenAIResponsesClient) Stream(ctx context.Context, request Request) (<-chan StreamEvent, error) {
	payload := c.payload(request, true)
	httpResponse, err := c.do(ctx, payload)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(httpResponse); err != nil {
		httpResponse.Body.Close()
		return nil, err
	}

	events := make(chan StreamEvent)
	go func() {
		defer close(events)
		defer httpResponse.Body.Close()

		scanner := bufio.NewScanner(httpResponse.Body)
		scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
		eventName := ""
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if data == "" || data == "[DONE]" {
					continue
				}
				var raw struct {
					Type     string `json:"type"`
					Delta    string `json:"delta"`
					Refusal  string `json:"refusal"`
					Message  string `json:"message"`
					Code     string `json:"code"`
					Param    string `json:"param"`
					Response *struct {
						ID                string                      `json:"id"`
						Status            string                      `json:"status"`
						Error             *responsesAPIError          `json:"error"`
						IncompleteDetails *responsesIncompleteDetails `json:"incomplete_details"`
					} `json:"response"`
					Error *responsesAPIError `json:"error"`
				}
				if err := json.Unmarshal([]byte(data), &raw); err != nil {
					sendStreamEvent(ctx, events, StreamEvent{Type: "error", Err: fmt.Errorf("解析 SSE 事件失败: %w", err)})
					return
				}
				if raw.Type == "" {
					raw.Type = eventName
				}
				responseID := ""
				if raw.Response != nil {
					responseID = raw.Response.ID
				}
				if raw.Type == "error" {
					message := strings.TrimSpace(raw.Message)
					if message == "" && raw.Error != nil {
						message = strings.TrimSpace(raw.Error.Message)
					}
					if message == "" {
						message = "Responses API 返回未说明的流式错误"
					}
					sendStreamEvent(ctx, events, StreamEvent{Type: "error", ResponseID: responseID, Err: errors.New(message)})
					return
				}
				if raw.Type == "response.failed" {
					message := "Responses API 流式响应失败"
					if raw.Response != nil && raw.Response.Error != nil && strings.TrimSpace(raw.Response.Error.Message) != "" {
						message = raw.Response.Error.Message
					}
					sendStreamEvent(ctx, events, StreamEvent{Type: "error", ResponseID: responseID, Err: errors.New(message)})
					return
				}
				if raw.Type == "response.incomplete" {
					message := "Responses API 流式响应不完整"
					if raw.Response != nil && raw.Response.IncompleteDetails != nil && raw.Response.IncompleteDetails.Reason != "" {
						message += ": " + raw.Response.IncompleteDetails.Reason
					}
					sendStreamEvent(ctx, events, StreamEvent{Type: "error", ResponseID: responseID, Err: errors.New(message)})
					return
				}
				if raw.Type == "response.refusal.done" && raw.Delta == "" {
					raw.Delta = raw.Refusal
				}
				if !sendStreamEvent(ctx, events, StreamEvent{Type: raw.Type, Delta: raw.Delta, ResponseID: responseID}) {
					return
				}
				if raw.Type == "response.completed" {
					return
				}
			}
		}
		if ctx.Err() != nil {
			return
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			sendStreamEvent(ctx, events, StreamEvent{Type: "error", Err: err})
			return
		}
		sendStreamEvent(ctx, events, StreamEvent{Type: "error", Err: errors.New("Responses API SSE 在终止事件前结束")})
	}()
	return events, nil
}

func responseStatusError(decoded responsesResponse) error {
	if decoded.Error != nil {
		message := strings.TrimSpace(decoded.Error.Message)
		if message == "" {
			message = "未说明的模型错误"
		}
		return fmt.Errorf("Responses API error %s: %s", decoded.Error.Code, message)
	}
	switch decoded.Status {
	case "", "completed":
		return nil
	case "incomplete":
		reason := "unknown"
		if decoded.IncompleteDetails != nil && decoded.IncompleteDetails.Reason != "" {
			reason = decoded.IncompleteDetails.Reason
		}
		return fmt.Errorf("Responses API 响应不完整: %s", reason)
	case "failed", "cancelled":
		return fmt.Errorf("Responses API 响应状态为 %s", decoded.Status)
	default:
		return fmt.Errorf("Responses API 返回非终态 %s", decoded.Status)
	}
}

func (c *OpenAIResponsesClient) payload(request Request, stream bool) responsesRequest {
	model := request.Model
	if model == "" {
		model = c.model
	}
	var include []string
	if !request.Store {
		include = []string{"reasoning.encrypted_content"}
	}
	return responsesRequest{
		Model:              model,
		Input:              request.Input,
		Tools:              request.Tools,
		Include:            include,
		PreviousResponseID: request.PreviousResponseID,
		Store:              request.Store,
		Stream:             stream,
	}
}

func (c *OpenAIResponsesClient) do(ctx context.Context, payload responsesRequest) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/responses", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	return c.client.Do(request)
}

func checkStatus(response *http.Response) error {
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	return fmt.Errorf("模型服务返回 HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
}

func sendStreamEvent(ctx context.Context, events chan<- StreamEvent, event StreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case events <- event:
		return true
	}
}
