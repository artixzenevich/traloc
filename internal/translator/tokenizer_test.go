package translator

import (
	"strings"
	"testing"
)

func TestEstimateTokens_Empty(t *testing.T) {
	if n := EstimateTokens(""); n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
}

func TestEstimateTokens_Latin(t *testing.T) {
	// ~4 latin chars per token
	n := EstimateTokens("hello world")
	if n < 2 || n > 5 {
		t.Fatalf("unexpected token count: %d", n)
	}
}

func TestEstimateTokens_Cyrillic(t *testing.T) {
	// ~2 cyrillic chars per token
	n := EstimateTokens("привет мир")
	if n < 4 || n > 8 {
		t.Fatalf("unexpected token count: %d", n)
	}
}

func TestSplitIntoChunks_Small(t *testing.T) {
	text := "один параграф\n\nвторой параграф"
	chunks := SplitIntoChunks(text, 10000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_Multiple(t *testing.T) {
	var paragraphs []string
	for i := 0; i < 10; i++ {
		paragraphs = append(paragraphs, strings.Repeat("word ", 50))
	}
	text := strings.Join(paragraphs, "\n\n")

	chunks := SplitIntoChunks(text, 100)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	joined := strings.Join(chunks, "\n\n")
	for _, p := range paragraphs {
		if !strings.Contains(joined, strings.TrimSpace(p)) {
			t.Fatal("missing paragraph in output")
		}
	}
}

func TestSplitIntoChunks_LongParagraph(t *testing.T) {
	longText := strings.Repeat("long sentence. ", 200)
	chunks := SplitIntoChunks(longText, 100)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for long paragraph, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_Empty(t *testing.T) {
	chunks := SplitIntoChunks("", 100)
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}
