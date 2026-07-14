package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAIResponsesClientGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["model"] != "test-model" {
			t.Fatalf("model = %#v", body["model"])
		}
		include, _ := body["include"].([]any)
		if len(include) != 1 || include[0] != "reasoning.encrypted_content" {
			t.Fatalf("include = %#v", body["include"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
          "id":"resp_test",
          "status":"completed",
          "output":[
            {"type":"reasoning","id":"rs_1","summary":[]},
            {"type":"message","content":[{"type":"output_text","text":"hello"}]},
            {"type":"function_call","call_id":"call_1","name":"calculator","arguments":"{\"expression\":\"1+1\"}"}
          ]
        }`))
	}))
	defer server.Close()

	client := NewOpenAIResponsesClient("test-key", server.URL, "test-model", time.Second)
	response, err := client.Generate(context.Background(), Request{Input: []InputItem{UserText("hi")}})
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "resp_test" || response.Status != "completed" || response.Text != "hello" || len(response.ToolCalls) != 1 {
		t.Fatalf("response = %#v", response)
	}
	if len(response.OutputItems) != 3 || response.OutputItems[0]["type"] != "reasoning" {
		t.Fatalf("output items = %#v", response.OutputItems)
	}
}

func TestOpenAIResponsesClientGeneratePreservesRefusal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
          "id":"resp_refusal",
          "status":"completed",
          "output":[{"type":"message","content":[{"type":"refusal","refusal":"I cannot help with that."}]}]
        }`))
	}))
	defer server.Close()

	client := NewOpenAIResponsesClient("test-key", server.URL, "test-model", time.Second)
	response, err := client.Generate(context.Background(), Request{Input: []InputItem{UserText("hi")}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Refusal != "I cannot help with that." || response.Text != response.Refusal {
		t.Fatalf("response = %#v", response)
	}
}

func TestOpenAIResponsesClientGenerateRejectsIncomplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
          "id":"resp_incomplete",
          "status":"incomplete",
          "incomplete_details":{"reason":"max_output_tokens"},
          "output":[]
        }`))
	}))
	defer server.Close()

	client := NewOpenAIResponsesClient("test-key", server.URL, "test-model", time.Second)
	_, err := client.Generate(context.Background(), Request{Input: []InputItem{UserText("hi")}})
	if err == nil || !strings.Contains(err.Error(), "max_output_tokens") {
		t.Fatalf("error = %v", err)
	}
}

func TestOpenAIResponsesClientStreamParsesTopLevelError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: error\ndata: {\"type\":\"error\",\"code\":\"server_error\",\"message\":\"boom\",\"param\":null,\"sequence_number\":1}\n\n"))
	}))
	defer server.Close()

	client := NewOpenAIResponsesClient("test-key", server.URL, "test-model", time.Second)
	events, err := client.Stream(context.Background(), Request{Input: []InputItem{UserText("hi")}})
	if err != nil {
		t.Fatal(err)
	}
	event := <-events
	if event.Err == nil || !strings.Contains(event.Err.Error(), "boom") {
		t.Fatalf("event = %#v", event)
	}
}

func TestOpenAIResponsesClientStreamRequiresTerminalEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}\n\n"))
	}))
	defer server.Close()

	client := NewOpenAIResponsesClient("test-key", server.URL, "test-model", time.Second)
	events, err := client.Stream(context.Background(), Request{Input: []InputItem{UserText("hi")}})
	if err != nil {
		t.Fatal(err)
	}
	first := <-events
	second := <-events
	if first.Delta != "partial" || second.Err == nil || !strings.Contains(second.Err.Error(), "终止事件") {
		t.Fatalf("events = %#v %#v", first, second)
	}
}

func TestOpenAIResponsesClientStreamCancellationClosesDownstreamRequest(t *testing.T) {
	canceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.created\ndata: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_1\"}}\n\n"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(canceled)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewOpenAIResponsesClient("test-key", server.URL, "test-model", time.Second)
	events, err := client.Stream(ctx, Request{Input: []InputItem{UserText("hi")}})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	for range events {
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("downstream request was not canceled")
	}
}
