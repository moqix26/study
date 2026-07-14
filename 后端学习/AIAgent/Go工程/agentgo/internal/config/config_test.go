package config

import (
	"testing"
	"time"
)

func TestValidateMockDefaults(t *testing.T) {
	cfg := Config{
		ServerAddr:     ":8080",
		AIProvider:     ProviderMock,
		RequestTimeout: 30 * time.Second,
		MaxAgentSteps:  4,
		RAGStore:       StoreMemory,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateOpenAIRequiresCredentials(t *testing.T) {
	cfg := Config{
		ServerAddr:     ":8080",
		AIProvider:     ProviderOpenAI,
		RequestTimeout: 30 * time.Second,
		MaxAgentSteps:  4,
		RAGStore:       StoreMemory,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected error")
	}
}
