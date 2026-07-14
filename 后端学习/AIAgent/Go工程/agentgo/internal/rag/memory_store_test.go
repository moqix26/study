package rag

import (
	"context"
	"strings"
	"testing"
)

func TestChunkerUsesRealOverlap(t *testing.T) {
	chunker := NewRuneChunker(100, 20)
	text := strings.Repeat("甲", 180)
	chunks := chunker.Split(text)
	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d", len(chunks))
	}
	if got := len([]rune(chunks[1])); got != 100 {
		t.Fatalf("second chunk runes = %d", got)
	}
}

func TestMemoryStoreTenantIsolation(t *testing.T) {
	store := NewMemoryStore(NewHashEmbedder(64), NewRuneChunker(120, 20))
	ctx := context.Background()
	_, _ = store.Upsert(ctx, Document{ID: "doc", TenantID: "u1", Title: "Go", Content: "Gin 是 Go Web 框架"})
	_, _ = store.Upsert(ctx, Document{ID: "doc", TenantID: "u2", Title: "Private", Content: "其他用户私有内容"})

	results, err := store.Search(ctx, "u1", "Gin Go", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Chunk.TenantID != "u1" {
		t.Fatalf("results = %#v", results)
	}
}
