package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"study.local/agentgo/internal/llm"
	"study.local/agentgo/internal/tool"
)

const (
	StopCompleted       = "COMPLETED"
	StopMaxSteps        = "MAX_STEPS"
	StopRepeatedCall    = "REPEATED_TOOL_CALL"
	StopInvalidResponse = "INVALID_MODEL_RESPONSE"
)

type Step struct {
	Index        int    `json:"index"`
	Tool         string `json:"tool"`
	Arguments    any    `json:"arguments,omitempty"`
	Result       string `json:"result_summary"`
	Status       string `json:"status"`
	DurationMS   int64  `json:"duration_ms"`
	ModelCallID  string `json:"model_call_id"`
	ModelReplyID string `json:"model_response_id"`
}

type RunResult struct {
	FinalAnswer string `json:"final_answer"`
	Steps       []Step `json:"steps"`
	StopReason  string `json:"stop_reason"`
}

type Runner struct {
	client         llm.Client
	registry       *tool.Registry
	maxSteps       int
	perToolTimeout time.Duration
}

func NewRunner(client llm.Client, registry *tool.Registry, maxSteps int, perToolTimeout time.Duration) *Runner {
	if maxSteps < 1 {
		maxSteps = 4
	}
	if perToolTimeout <= 0 {
		perToolTimeout = 5 * time.Second
	}
	return &Runner{client: client, registry: registry, maxSteps: maxSteps, perToolTimeout: perToolTimeout}
}

func (r *Runner) Run(ctx context.Context, principal tool.Principal, message string) (RunResult, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return RunResult{}, errors.New("message 不能为空")
	}
	request := llm.Request{
		Input: []llm.InputItem{llm.UserText(message)},
		Tools: r.registry.Definitions(),
		Store: false,
	}
	steps := make([]Step, 0, r.maxSteps)
	seen := make(map[string]int)

	for round := 0; round < r.maxSteps; round++ {
		response, err := r.client.Generate(ctx, request)
		if err != nil {
			return RunResult{Steps: steps, StopReason: "MODEL_ERROR"}, err
		}
		if len(response.ToolCalls) == 0 {
			if strings.TrimSpace(response.Text) == "" {
				return RunResult{Steps: steps, StopReason: StopInvalidResponse}, errors.New("模型既没有文本也没有 Tool Call")
			}
			return RunResult{FinalAnswer: response.Text, Steps: steps, StopReason: StopCompleted}, nil
		}

		outputs := make([]llm.InputItem, 0, len(response.ToolCalls))
		for _, call := range response.ToolCalls {
			if strings.TrimSpace(call.CallID) == "" || strings.TrimSpace(call.Name) == "" {
				return RunResult{Steps: steps, StopReason: StopInvalidResponse}, errors.New("模型返回的 Tool Call 缺少 call_id 或 name")
			}

			var parsedArguments any
			argumentsErr := json.Unmarshal(call.Arguments, &parsedArguments)
			canonicalArguments := string(call.Arguments)
			if argumentsErr == nil {
				if encoded, err := json.Marshal(parsedArguments); err == nil {
					canonicalArguments = string(encoded)
				}
			}
			fingerprint := call.Name + "\x00" + canonicalArguments
			seen[fingerprint]++
			if seen[fingerprint] > 2 {
				return RunResult{
					FinalAnswer: "检测到重复工具调用，任务已安全停止。",
					Steps:       steps,
					StopReason:  StopRepeatedCall,
				}, nil
			}

			started := time.Now()
			var execution tool.Execution
			var executeErr error
			if argumentsErr != nil {
				executeErr = fmt.Errorf("Tool %s 参数不是合法 JSON: %w", call.Name, argumentsErr)
			} else {
				toolCtx, cancel := context.WithTimeout(ctx, r.perToolTimeout)
				execution, executeErr = r.registry.Execute(toolCtx, principal, call.Name, call.Arguments)
				cancel()
			}

			step := Step{
				Index:        len(steps) + 1,
				Tool:         call.Name,
				DurationMS:   time.Since(started).Milliseconds(),
				ModelCallID:  call.CallID,
				ModelReplyID: response.ID,
			}
			output := execution.Output
			if executeErr != nil {
				step.Status = "ERROR"
				step.Result = safeError(executeErr)
				encoded, _ := json.Marshal(map[string]any{"ok": false, "error": step.Result})
				output = string(encoded)
			} else {
				step.Status = "OK"
				step.Arguments = parsedArguments
				step.Result = execution.Summary
			}
			steps = append(steps, step)
			outputs = append(outputs, llm.FunctionOutput(call.CallID, output))
		}

		continuation := continuationItems(request.Input, response)
		continuation = append(continuation, outputs...)
		request = llm.Request{Input: continuation, Tools: r.registry.Definitions(), Store: false}
	}

	return RunResult{
		FinalAnswer: fmt.Sprintf("达到最大步骤数 %d，任务已停止。", r.maxSteps),
		Steps:       steps,
		StopReason:  StopMaxSteps,
	}, nil
}

func continuationItems(history []llm.InputItem, response llm.Response) []llm.InputItem {
	items := make([]llm.InputItem, 0, len(history)+len(response.OutputItems)+len(response.ToolCalls))
	for _, item := range history {
		copyItem := make(llm.InputItem, len(item))
		for key, value := range item {
			copyItem[key] = value
		}
		items = append(items, copyItem)
	}
	seenCallIDs := make(map[string]struct{}, len(response.ToolCalls))
	for _, item := range response.OutputItems {
		copyItem := make(llm.InputItem, len(item))
		for key, value := range item {
			copyItem[key] = value
		}
		items = append(items, copyItem)
		if item["type"] == "function_call" {
			if callID, _ := item["call_id"].(string); callID != "" {
				seenCallIDs[callID] = struct{}{}
			}
		}
	}
	for _, call := range response.ToolCalls {
		if _, exists := seenCallIDs[call.CallID]; exists {
			continue
		}
		items = append(items, llm.InputItem{
			"type":      "function_call",
			"call_id":   call.CallID,
			"name":      call.Name,
			"arguments": string(call.Arguments),
		})
	}
	return items
}

func safeError(err error) string {
	message := strings.TrimSpace(err.Error())
	if len([]rune(message)) > 240 {
		return string([]rune(message)[:240]) + "…"
	}
	return message
}
