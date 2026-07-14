package rag

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

type HashEmbedder struct {
	dimension int
}

func NewHashEmbedder(dimension int) *HashEmbedder {
	if dimension < 16 {
		dimension = DefaultDimension
	}
	return &HashEmbedder{dimension: dimension}
}

func (e *HashEmbedder) Dimension() int {
	return e.dimension
}

func (e *HashEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	vector := make([]float32, e.dimension)
	for _, token := range tokenize(text) {
		hash := fnv.New64a()
		_, _ = hash.Write([]byte(token))
		value := hash.Sum64()
		index := int(value % uint64(e.dimension))
		sign := float32(1)
		if value&(1<<63) != 0 {
			sign = -1
		}
		vector[index] += sign
	}
	normalize(vector)
	return vector, nil
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
}

func normalize(vector []float32) {
	var sum float64
	for _, value := range vector {
		sum += float64(value * value)
	}
	if sum == 0 {
		return
	}
	norm := float32(math.Sqrt(sum))
	for i := range vector {
		vector[i] /= norm
	}
}

func cosine(left, right []float32) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}
	var dot float64
	for i := range left {
		dot += float64(left[i] * right[i])
	}
	return dot
}
