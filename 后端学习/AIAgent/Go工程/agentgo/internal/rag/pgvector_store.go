package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxHNSWVectorDimensions = 2000

type PGVectorStore struct {
	pool     *pgxpool.Pool
	embedder Embedder
	chunker  RuneChunker
}

func NewPGVectorStore(ctx context.Context, databaseURL string, embedder Embedder, chunker RuneChunker) (*PGVectorStore, error) {
	if embedder == nil {
		return nil, errors.New("embedder 不能为空")
	}
	if err := validateHNSWDimension(embedder.Dimension()); err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("创建 PostgreSQL 连接池失败: %w", err)
	}
	store := &PGVectorStore{pool: pool, embedder: embedder, chunker: chunker}
	if err := store.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *PGVectorStore) migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		return fmt.Errorf("启用 vector 扩展失败: %w", err)
	}
	dimension := s.embedder.Dimension()
	if err := validateHNSWDimension(dimension); err != nil {
		return err
	}
	ddl := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS rag_chunks (
    tenant_id   text        NOT NULL,
    document_id text        NOT NULL,
    chunk_id    text        NOT NULL,
    title       text        NOT NULL DEFAULT '',
    content     text        NOT NULL,
    metadata    jsonb       NOT NULL DEFAULT '{}'::jsonb,
    embedding   vector(%d)  NOT NULL,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, document_id, chunk_id)
);
CREATE INDEX IF NOT EXISTS rag_chunks_tenant_document_idx
    ON rag_chunks (tenant_id, document_id);
CREATE INDEX IF NOT EXISTS rag_chunks_embedding_hnsw_idx
    ON rag_chunks USING hnsw (embedding vector_cosine_ops);
`, dimension)
	if _, err := s.pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("初始化 rag_chunks 失败: %w", err)
	}
	return nil
}

func validateHNSWDimension(dimension int) error {
	if dimension < 1 || dimension > maxHNSWVectorDimensions {
		return fmt.Errorf("HNSW vector 维度必须在 1 到 %d 之间，当前为 %d", maxHNSWVectorDimensions, dimension)
	}
	return nil
}

func (s *PGVectorStore) Upsert(ctx context.Context, document Document) (int, error) {
	if strings.TrimSpace(document.TenantID) == "" || strings.TrimSpace(document.ID) == "" {
		return 0, errors.New("tenantID 和 document ID 不能为空")
	}
	parts := s.chunker.Split(document.Content)
	if len(parts) == 0 {
		return 0, errors.New("文档内容不能为空")
	}

	type preparedChunk struct {
		id        string
		content   string
		embedding string
	}
	prepared := make([]preparedChunk, 0, len(parts))
	for index, content := range parts {
		embedding, err := s.embedder.Embed(ctx, content)
		if err != nil {
			return 0, err
		}
		prepared = append(prepared, preparedChunk{
			id:        fmt.Sprintf("%s-%04d", document.ID, index),
			content:   content,
			embedding: vectorLiteral(embedding),
		})
	}
	metadata, err := json.Marshal(document.Metadata)
	if err != nil {
		return 0, fmt.Errorf("编码 metadata 失败: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM rag_chunks WHERE tenant_id=$1 AND document_id=$2`, document.TenantID, document.ID); err != nil {
		return 0, err
	}
	for _, chunk := range prepared {
		_, err := tx.Exec(ctx, `
INSERT INTO rag_chunks (tenant_id, document_id, chunk_id, title, content, metadata, embedding, updated_at)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::vector, now())
`, document.TenantID, document.ID, chunk.id, document.Title, chunk.content, metadata, chunk.embedding)
		if err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(prepared), nil
}

func (s *PGVectorStore) Search(ctx context.Context, tenantID, query string, topK int) ([]Result, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(query) == "" {
		return nil, errors.New("tenantID 和 query 不能为空")
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
	vector := vectorLiteral(embedding)

	rows, err := s.pool.Query(ctx, `
SELECT chunk_id, document_id, tenant_id, title, content, metadata,
       1 - (embedding <=> $1::vector) AS score
FROM rag_chunks
WHERE tenant_id = $2
ORDER BY embedding <=> $1::vector
LIMIT $3
`, vector, tenantID, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Result, 0, topK)
	for rows.Next() {
		var result Result
		var metadata []byte
		if err := rows.Scan(
			&result.Chunk.ID,
			&result.Chunk.DocumentID,
			&result.Chunk.TenantID,
			&result.Chunk.Title,
			&result.Chunk.Content,
			&metadata,
			&result.Score,
		); err != nil {
			return nil, err
		}
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &result.Chunk.Metadata)
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (s *PGVectorStore) Close() {
	s.pool.Close()
}

func vectorLiteral(vector []float32) string {
	var builder strings.Builder
	builder.WriteByte('[')
	for index, value := range vector {
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.FormatFloat(float64(value), 'g', -1, 32))
	}
	builder.WriteByte(']')
	return builder.String()
}
