package rag

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	mu       sync.RWMutex
	embedder Embedder
	chunker  RuneChunker
	chunks   []Chunk
}

func NewMemoryStore(embedder Embedder, chunker RuneChunker) *MemoryStore {
	return &MemoryStore{embedder: embedder, chunker: chunker}
}

func (s *MemoryStore) Upsert(ctx context.Context, document Document) (int, error) {
	if strings.TrimSpace(document.TenantID) == "" || strings.TrimSpace(document.ID) == "" {
		return 0, errors.New("tenantID 和 document ID 不能为空")
	}
	parts := s.chunker.Split(document.Content)
	if len(parts) == 0 {
		return 0, errors.New("文档内容不能为空")
	}

	newChunks := make([]Chunk, 0, len(parts))
	for index, content := range parts {
		embedding, err := s.embedder.Embed(ctx, content)
		if err != nil {
			return 0, err
		}
		newChunks = append(newChunks, Chunk{
			ID:         fmt.Sprintf("%s-%04d", document.ID, index),
			DocumentID: document.ID,
			TenantID:   document.TenantID,
			Title:      document.Title,
			Content:    content,
			Metadata:   cloneMetadata(document.Metadata),
			Embedding:  embedding,
		})
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.chunks[:0]
	for _, chunk := range s.chunks {
		if chunk.TenantID == document.TenantID && chunk.DocumentID == document.ID {
			continue
		}
		kept = append(kept, chunk)
	}
	s.chunks = append(kept, newChunks...)
	return len(newChunks), nil
}

func (s *MemoryStore) Search(ctx context.Context, tenantID, query string, topK int) ([]Result, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, errors.New("tenantID 不能为空")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query 不能为空")
	}
	if topK < 1 {
		topK = 3
	}
	if topK > 20 {
		topK = 20
	}
	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	results := make([]Result, 0, len(s.chunks))
	for _, chunk := range s.chunks {
		if chunk.TenantID != tenantID {
			continue
		}
		results = append(results, Result{Chunk: cloneChunk(chunk), Score: cosine(embedding, chunk.Embedding)})
	}
	s.mu.RUnlock()

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

func (s *MemoryStore) Close() {}

func cloneChunk(chunk Chunk) Chunk {
	chunk.Metadata = cloneMetadata(chunk.Metadata)
	chunk.Embedding = append([]float32(nil), chunk.Embedding...)
	return chunk
}

func cloneMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}
	result := make(map[string]string, len(metadata))
	for key, value := range metadata {
		result[key] = value
	}
	return result
}
