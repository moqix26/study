package rag

import "strings"

type RuneChunker struct {
	Size    int
	Overlap int
}

func NewRuneChunker(size, overlap int) RuneChunker {
	if size < 100 {
		size = 500
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 5
	}
	return RuneChunker{Size: size, Overlap: overlap}
}

func (c RuneChunker) Split(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	runes := []rune(text)
	step := c.Size - c.Overlap
	chunks := make([]string, 0, (len(runes)+step-1)/step)
	for start := 0; start < len(runes); start += step {
		end := start + c.Size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}
