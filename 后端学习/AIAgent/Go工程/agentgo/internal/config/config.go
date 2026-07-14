package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ProviderMock   = "mock"
	ProviderOpenAI = "openai"
	StoreMemory    = "memory"
	StorePGVector  = "pgvector"
)

type Config struct {
	ServerAddr     string
	AIProvider     string
	AIModel        string
	OpenAIAPIKey   string
	OpenAIBaseURL  string
	RequestTimeout time.Duration
	MaxAgentSteps  int
	RAGStore       string
	DatabaseURL    string
}

func Load() Config {
	return Config{
		ServerAddr:     env("SERVER_ADDR", ":8080"),
		AIProvider:     strings.ToLower(env("AI_PROVIDER", ProviderMock)),
		AIModel:        strings.TrimSpace(os.Getenv("AI_MODEL")),
		OpenAIAPIKey:   strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		OpenAIBaseURL:  strings.TrimRight(env("OPENAI_BASE_URL", "https://api.openai.com"), "/"),
		RequestTimeout: durationEnv("REQUEST_TIMEOUT", 45*time.Second),
		MaxAgentSteps:  intEnv("MAX_AGENT_STEPS", 4),
		RAGStore:       strings.ToLower(env("RAG_STORE", StoreMemory)),
		DatabaseURL:    strings.TrimSpace(os.Getenv("DATABASE_URL")),
	}
}

func (c Config) Validate() error {
	if c.ServerAddr == "" {
		return errors.New("SERVER_ADDR 不能为空")
	}
	if c.RequestTimeout <= 0 || c.RequestTimeout > 10*time.Minute {
		return fmt.Errorf("REQUEST_TIMEOUT 必须在 0 到 10 分钟之间，当前为 %s", c.RequestTimeout)
	}
	if c.MaxAgentSteps < 1 || c.MaxAgentSteps > 12 {
		return fmt.Errorf("MAX_AGENT_STEPS 必须在 1 到 12 之间，当前为 %d", c.MaxAgentSteps)
	}

	switch c.AIProvider {
	case ProviderMock:
	case ProviderOpenAI:
		if c.OpenAIAPIKey == "" {
			return errors.New("AI_PROVIDER=openai 时必须设置 OPENAI_API_KEY")
		}
		if c.AIModel == "" {
			return errors.New("AI_PROVIDER=openai 时必须显式设置 AI_MODEL")
		}
		if c.OpenAIBaseURL == "" {
			return errors.New("OPENAI_BASE_URL 不能为空")
		}
	default:
		return fmt.Errorf("不支持的 AI_PROVIDER %q", c.AIProvider)
	}

	switch c.RAGStore {
	case StoreMemory:
	case StorePGVector:
		if c.DatabaseURL == "" {
			return errors.New("RAG_STORE=pgvector 时必须设置 DATABASE_URL")
		}
	default:
		return fmt.Errorf("不支持的 RAG_STORE %q", c.RAGStore)
	}
	return nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
