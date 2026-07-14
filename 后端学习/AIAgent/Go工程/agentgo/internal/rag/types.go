package rag

import "context"

const DefaultDimension = 128

type Document struct {
	ID       string            `json:"id"`
	TenantID string            `json:"tenant_id"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Chunk struct {
	ID         string            `json:"id"`
	DocumentID string            `json:"document_id"`
	TenantID   string            `json:"tenant_id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Embedding  []float32         `json:"-"`
}

type Result struct {
	Chunk Chunk   `json:"chunk"`
	Score float64 `json:"score"`
}

type Store interface {
	Upsert(ctx context.Context, document Document) (int, error)
	Search(ctx context.Context, tenantID, query string, topK int) ([]Result, error)
	Close()
}

type Embedder interface {
	Dimension() int
	Embed(ctx context.Context, text string) ([]float32, error)
}
