package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"study.local/agentgo/internal/agent"
	"study.local/agentgo/internal/config"
	"study.local/agentgo/internal/httpapi"
	"study.local/agentgo/internal/llm"
	"study.local/agentgo/internal/memory"
	"study.local/agentgo/internal/rag"
	"study.local/agentgo/internal/tool"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置错误: %v", err)
	}

	client := buildClient(cfg)
	embedder := rag.NewHashEmbedder(rag.DefaultDimension)
	chunker := rag.NewRuneChunker(600, 100)
	ragStore := buildRAGStore(cfg, embedder, chunker)
	defer ragStore.Close()

	registry := tool.NewRegistry()
	mustRegister(registry, tool.CalculatorTool())
	mustRegister(registry, tool.KnowledgeSearchTool(ragStore))

	agentRunner := agent.NewRunner(client, registry, cfg.MaxAgentSteps, 8*time.Second)
	api := httpapi.New(httpapi.Options{
		Client:         client,
		Memory:         memory.NewWindowStore(20),
		RAG:            ragStore,
		Agent:          agentRunner,
		RequestTimeout: cfg.RequestTimeout,
		ProviderName:   cfg.AIProvider,
		StoreName:      cfg.RAGStore,
	})

	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	go func() {
		log.Printf("agentgo listening on %s provider=%s rag=%s", cfg.ServerAddr, cfg.AIProvider, cfg.RAGStore)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("优雅关闭失败: %v", err)
	}
}

func buildClient(cfg config.Config) llm.Client {
	switch cfg.AIProvider {
	case config.ProviderMock:
		return llm.NewMockClient()
	case config.ProviderOpenAI:
		return llm.NewOpenAIResponsesClient(
			cfg.OpenAIAPIKey,
			cfg.OpenAIBaseURL,
			cfg.AIModel,
			cfg.RequestTimeout,
		)
	default:
		panic("配置已验证，不应出现未知 Provider")
	}
}

func buildRAGStore(cfg config.Config, embedder rag.Embedder, chunker rag.RuneChunker) rag.Store {
	if cfg.RAGStore == config.StoreMemory {
		return rag.NewMemoryStore(embedder, chunker)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	store, err := rag.NewPGVectorStore(ctx, cfg.DatabaseURL, embedder, chunker)
	if err != nil {
		log.Fatalf("初始化 pgvector 失败: %v", err)
	}
	return store
}

func mustRegister(registry *tool.Registry, registered tool.RegisteredTool) {
	if err := registry.Register(registered); err != nil {
		log.Fatalf("注册 Tool 失败: %v", err)
	}
}
