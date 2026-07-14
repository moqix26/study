package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"study.local/agentgo/internal/llm"
	"study.local/agentgo/internal/tool"
)

type scriptedClient struct {
	requests  []llm.Request
	responses []llm.Response
}

func (c *scriptedClient) Generate(_ context.Context, request llm.Request) (llm.Response, error) {
	c.requests = append(c.requests, request)
	response := c.responses[0]
	c.responses = c.responses[1:]
	return response, nil
}

func (c *scriptedClient) Stream(context.Context, llm.Request) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

func TestRunnerUsesCalculator(t *testing.T) {
	registry := tool.NewRegistry()
	if err := registry.Register(tool.CalculatorTool()); err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(llm.NewMockClient(), registry, 4, time.Second)
	result, err := runner.Run(context.Background(), tool.Principal{TenantID: "u1", UserID: "u1"}, "请计算 12+8")
	if err != nil {
		t.Fatal(err)
	}
	if result.StopReason != StopCompleted || len(result.Steps) != 1 || result.Steps[0].Tool != "calculator" {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunnerCarriesFullOutputStatelessly(t *testing.T) {
	registry := tool.NewRegistry()
	if err := registry.Register(tool.CalculatorTool()); err != nil {
		t.Fatal(err)
	}
	arguments := json.RawMessage(`{"expression":"1+1"}`)
	secondArguments := json.RawMessage(`{"expression":"2+2"}`)
	client := &scriptedClient{responses: []llm.Response{
		{
			ID: "resp_tool",
			OutputItems: []llm.InputItem{
				{"type": "reasoning", "id": "rs_1", "summary": []any{}},
				{"type": "function_call", "call_id": "call_1", "name": "calculator", "arguments": string(arguments)},
			},
			ToolCalls: []llm.ToolCall{{CallID: "call_1", Name: "calculator", Arguments: arguments}},
		},
		{
			ID: "resp_tool_2",
			OutputItems: []llm.InputItem{
				{"type": "reasoning", "id": "rs_2", "encrypted_content": "encrypted"},
				{"type": "function_call", "call_id": "call_2", "name": "calculator", "arguments": string(secondArguments)},
			},
			ToolCalls: []llm.ToolCall{{CallID: "call_2", Name: "calculator", Arguments: secondArguments}},
		},
		{ID: "resp_final", Text: "2"},
	}}

	runner := NewRunner(client, registry, 4, time.Second)
	result, err := runner.Run(context.Background(), tool.Principal{TenantID: "u1", UserID: "u1"}, "计算 1+1")
	if err != nil {
		t.Fatal(err)
	}
	if result.FinalAnswer != "2" || len(client.requests) != 3 || len(result.Steps) != 2 {
		t.Fatalf("result=%#v requests=%d", result, len(client.requests))
	}
	second := client.requests[1]
	if second.Store || second.PreviousResponseID != "" || len(second.Input) != 4 {
		t.Fatalf("second request = %#v", second)
	}
	if second.Input[0]["role"] != "user" || second.Input[1]["type"] != "reasoning" || second.Input[2]["type"] != "function_call" || second.Input[3]["type"] != "function_call_output" {
		t.Fatalf("continuation input = %#v", second.Input)
	}
	third := client.requests[2]
	if third.Store || third.PreviousResponseID != "" || len(third.Input) != 7 {
		t.Fatalf("third request = %#v", third)
	}
	if third.Input[0]["role"] != "user" || third.Input[4]["type"] != "reasoning" || third.Input[5]["type"] != "function_call" || third.Input[6]["type"] != "function_call_output" {
		t.Fatalf("accumulated input = %#v", third.Input)
	}
}

func TestRunnerDoesNotExposeInvalidRawArguments(t *testing.T) {
	registry := tool.NewRegistry()
	if err := registry.Register(tool.CalculatorTool()); err != nil {
		t.Fatal(err)
	}
	client := &scriptedClient{responses: []llm.Response{
		{
			ID:        "resp_bad_tool",
			ToolCalls: []llm.ToolCall{{CallID: "call_1", Name: "calculator", Arguments: json.RawMessage(`{`)}},
		},
		{ID: "resp_final", Text: "参数无效"},
	}}

	runner := NewRunner(client, registry, 4, time.Second)
	result, err := runner.Run(context.Background(), tool.Principal{TenantID: "u1", UserID: "u1"}, "计算")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 1 || result.Steps[0].Status != "ERROR" || result.Steps[0].Arguments != nil {
		t.Fatalf("result = %#v", result)
	}
	if _, err := json.Marshal(result); err != nil {
		t.Fatalf("RunResult must remain JSON serializable: %v", err)
	}
}
