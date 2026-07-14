package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"study.local/agentgo/internal/llm"
)

type Principal struct {
	TenantID string
	UserID   string
}

type Handler func(ctx context.Context, principal Principal, arguments json.RawMessage) (any, error)

type RegisteredTool struct {
	Definition llm.ToolDefinition
	Handler    Handler
}

type Execution struct {
	Output  string
	Summary string
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]RegisteredTool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]RegisteredTool)}
}

func (r *Registry) Register(tool RegisteredTool) error {
	name := strings.TrimSpace(tool.Definition.Name)
	if name == "" || tool.Handler == nil {
		return errors.New("Tool 名称和 Handler 不能为空")
	}
	if tool.Definition.Type == "" {
		tool.Definition.Type = "function"
	}
	if tool.Definition.Parameters == nil {
		return fmt.Errorf("Tool %s 缺少 JSON Schema", name)
	}
	tool.Definition.Name = name
	tool.Definition.Strict = true

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("Tool %s 已注册", name)
	}
	r.tools[name] = tool
	return nil
}

func (r *Registry) Definitions() []llm.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	definitions := make([]llm.ToolDefinition, 0, len(names))
	for _, name := range names {
		definitions = append(definitions, r.tools[name].Definition)
	}
	return definitions
}

func (r *Registry) Execute(ctx context.Context, principal Principal, name string, arguments json.RawMessage) (Execution, error) {
	if principal.TenantID == "" || principal.UserID == "" {
		return Execution{}, errors.New("缺少已认证 Principal")
	}
	r.mu.RLock()
	registered, exists := r.tools[name]
	r.mu.RUnlock()
	if !exists {
		return Execution{}, fmt.Errorf("Tool %q 未注册或未授权", name)
	}
	value, err := registered.Handler(ctx, principal, arguments)
	if err != nil {
		return Execution{}, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return Execution{}, err
	}
	return Execution{Output: string(encoded), Summary: summarize(string(encoded), 280)}, nil
}

func summarize(value string, maxRunes int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes]) + "…"
}
