package rag

import (
	"context"
	"strings"
	"testing"
)

func TestVectorLiteral(t *testing.T) {
	got := vectorLiteral([]float32{0.5, -1, 0})
	if got != "[0.5,-1,0]" {
		t.Fatalf("vectorLiteral() = %q", got)
	}
}

func TestValidateHNSWDimension(t *testing.T) {
	if err := validateHNSWDimension(maxHNSWVectorDimensions); err != nil {
		t.Fatal(err)
	}
	if err := validateHNSWDimension(maxHNSWVectorDimensions + 1); err == nil {
		t.Fatal("expected dimensions above pgvector HNSW limit to fail")
	}
}

func TestNewPGVectorStoreRejectsDimensionBeforeDatabaseSetup(t *testing.T) {
	_, err := NewPGVectorStore(
		context.Background(),
		"this database URL must not be parsed",
		NewHashEmbedder(maxHNSWVectorDimensions+1),
		NewRuneChunker(600, 100),
	)
	if err == nil || !strings.Contains(err.Error(), "HNSW vector 维度") {
		t.Fatalf("error = %v", err)
	}
}
