package httpapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"study.local/agentgo/internal/agent"
	"study.local/agentgo/internal/llm"
	"study.local/agentgo/internal/memory"
	"study.local/agentgo/internal/rag"
	"study.local/agentgo/internal/tool"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	return newTestServerWith(t, llm.NewMockClient(), memory.NewWindowStore(10), 0)
}

func newTestServerWith(t *testing.T, client llm.Client, memoryStore memory.Store, maxBodyBytes int64) *Server {
	t.Helper()
	ragStore := rag.NewMemoryStore(rag.NewHashEmbedder(64), rag.NewRuneChunker(120, 20))
	registry := tool.NewRegistry()
	if err := registry.Register(tool.CalculatorTool()); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(tool.KnowledgeSearchTool(ragStore)); err != nil {
		t.Fatal(err)
	}
	return New(Options{
		Client:         client,
		Memory:         memoryStore,
		RAG:            ragStore,
		Agent:          agent.NewRunner(client, registry, 4, time.Second),
		RequestTimeout: 5 * time.Second,
		MaxBodyBytes:   maxBodyBytes,
		ProviderName:   "mock",
		StoreName:      "memory",
	})
}

type failingClient struct{}

func (failingClient) Generate(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{}, errors.New("model failed")
}

func (failingClient) Stream(context.Context, llm.Request) (<-chan llm.StreamEvent, error) {
	return nil, errors.New("model failed")
}

type blockingClient struct {
	mu            sync.Mutex
	calls         int
	inFlight      int
	overlapped    bool
	inputs        [][]llm.InputItem
	firstStarted  chan struct{}
	secondStarted chan struct{}
	releaseFirst  chan struct{}
}

func newBlockingClient() *blockingClient {
	return &blockingClient{
		firstStarted:  make(chan struct{}),
		secondStarted: make(chan struct{}),
		releaseFirst:  make(chan struct{}),
	}
}

func (c *blockingClient) Generate(_ context.Context, request llm.Request) (llm.Response, error) {
	c.mu.Lock()
	c.calls++
	call := c.calls
	c.inFlight++
	if c.inFlight > 1 {
		c.overlapped = true
	}
	c.inputs = append(c.inputs, append([]llm.InputItem(nil), request.Input...))
	if call == 1 {
		close(c.firstStarted)
	} else if call == 2 {
		close(c.secondStarted)
	}
	c.mu.Unlock()

	if call == 1 {
		<-c.releaseFirst
	}

	c.mu.Lock()
	c.inFlight--
	c.mu.Unlock()
	return llm.Response{ID: fmt.Sprintf("resp_%d", call), Text: fmt.Sprintf("answer-%d", call)}, nil
}

func (c *blockingClient) Stream(context.Context, llm.Request) (<-chan llm.StreamEvent, error) {
	return nil, errors.New("not implemented")
}

type streamClient struct {
	events []llm.StreamEvent
}

func (c streamClient) Generate(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{ID: "unused", Text: "unused"}, nil
}

func (c streamClient) Stream(context.Context, llm.Request) (<-chan llm.StreamEvent, error) {
	events := make(chan llm.StreamEvent, len(c.events))
	for _, event := range c.events {
		events <- event
	}
	close(events)
	return events, nil
}

func performChat(server *Server, userID, conversationID, message string) *httptest.ResponseRecorder {
	body := fmt.Sprintf(`{"message":%q,"conversation_id":%q}`, message, conversationID)
	request := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", userID)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	return response
}

func TestHealth(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestChatRequiresIdentity(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"hi"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestChatMock(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"你好"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", "u1")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRequestBodyLimit(t *testing.T) {
	server := newTestServerWith(t, llm.NewMockClient(), memory.NewWindowStore(10), 64)
	request := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"`+strings.Repeat("x", 100)+`"}`))
	request.ContentLength = -1
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", "u1")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRejectsClientTenantHeader(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"hi"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", "u1")
	request.Header.Set("X-Tenant-ID", "other")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "untrusted_tenant_header") {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestFailedChatDoesNotCommitTurn(t *testing.T) {
	store := memory.NewWindowStore(10)
	server := newTestServerWith(t, failingClient{}, store, 0)
	response := performChat(server, "u1", "c1", "secret")
	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	history, err := store.History("u1", "u1", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 0 {
		t.Fatalf("failed turn was committed: %#v", history)
	}
}

func TestChatSerializesSameConversation(t *testing.T) {
	client := newBlockingClient()
	store := memory.NewWindowStore(10)
	server := newTestServerWith(t, client, store, 0)

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() { firstDone <- performChat(server, "u1", "c1", "first") }()
	<-client.firstStarted
	secondDone := make(chan *httptest.ResponseRecorder, 1)
	go func() { secondDone <- performChat(server, "u1", "c1", "second") }()

	select {
	case <-client.secondStarted:
		t.Fatal("second model call started before the first turn committed")
	case <-time.After(100 * time.Millisecond):
	}
	close(client.releaseFirst)
	if response := <-firstDone; response.Code != http.StatusOK {
		t.Fatalf("first status = %d body=%s", response.Code, response.Body.String())
	}
	if response := <-secondDone; response.Code != http.StatusOK {
		t.Fatalf("second status = %d body=%s", response.Code, response.Body.String())
	}

	client.mu.Lock()
	overlapped := client.overlapped
	inputs := append([][]llm.InputItem(nil), client.inputs...)
	client.mu.Unlock()
	if overlapped || len(inputs) != 2 || len(inputs[1]) != 3 {
		t.Fatalf("overlapped=%v inputs=%#v", overlapped, inputs)
	}
	if inputs[1][0]["content"] != "first" || inputs[1][1]["content"] != "answer-1" || inputs[1][2]["content"] != "second" {
		t.Fatalf("second input = %#v", inputs[1])
	}
}

func TestStreamRequiresTerminalAndCommitsOnlyCompletedTurn(t *testing.T) {
	t.Run("completed refusal", func(t *testing.T) {
		store := memory.NewWindowStore(10)
		client := streamClient{events: []llm.StreamEvent{
			{Type: "response.refusal.done", Delta: "cannot", ResponseID: "resp_1"},
			{Type: "response.completed", ResponseID: "resp_1"},
		}}
		server := newTestServerWith(t, client, store, 0)
		request := httptest.NewRequest(http.MethodPost, "/api/chat/stream", bytes.NewBufferString(`{"message":"bad","conversation_id":"c1"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-User-ID", "u1")
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"refusal":true`) || !strings.Contains(response.Body.String(), "event: done") {
			t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
		}
		history, _ := store.History("u1", "u1", "c1")
		if len(history) != 2 || history[1]["content"] != "cannot" {
			t.Fatalf("history = %#v", history)
		}
	})

	t.Run("missing terminal", func(t *testing.T) {
		store := memory.NewWindowStore(10)
		client := streamClient{events: []llm.StreamEvent{{Type: "response.output_text.delta", Delta: "partial"}}}
		server := newTestServerWith(t, client, store, 0)
		request := httptest.NewRequest(http.MethodPost, "/api/chat/stream", bytes.NewBufferString(`{"message":"hi","conversation_id":"c1"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-User-ID", "u1")
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if !strings.Contains(response.Body.String(), "model_stream_error") || strings.Contains(response.Body.String(), "event: done") {
			t.Fatalf("body=%s", response.Body.String())
		}
		history, _ := store.History("u1", "u1", "c1")
		if len(history) != 0 {
			t.Fatalf("incomplete stream was committed: %#v", history)
		}
	})
}
