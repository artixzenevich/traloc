package translator

import (
	"testing"
)

func TestCodeStore_MaskInline(t *testing.T) {
	store := NewCodeStore()
	input := "text with `inline code` inside"
	masked := store.Mask(input)

	if masked == input {
		t.Fatal("inline code was not masked")
	}

	restored := store.Restore(masked)
	if restored != input {
		t.Fatalf("restored text differs:\n  got:  %q\n  want: %q", restored, input)
	}
}

func TestCodeStore_MaskBlock(t *testing.T) {
	store := NewCodeStore()
	input := "before\n```go\nfmt.Println(\"hello\")\n```\nafter"
	masked := store.Mask(input)

	if masked == input {
		t.Fatal("code block was not masked")
	}

	restored := store.Restore(masked)
	expected := "before\n\n```go\nfmt.Println(\"hello\")\n```\n\nafter"
	if restored != expected {
		t.Fatalf("restored text differs:\n  got:  %q\n  want: %q", restored, expected)
	}
}

func TestCodeStore_MaskMultiple(t *testing.T) {
	store := NewCodeStore()
	input := "a `b` c `d` e\n```\nblock1\n```\nf `g` h"
	masked := store.Mask(input)
	restored := store.Restore(masked)

	expected := "a `b` c `d` e\n\n```\nblock1\n```\n\nf `g` h"
	if restored != expected {
		t.Fatalf("restored text differs:\n  got:  %q\n  want: %q", restored, expected)
	}
}

func TestCodeStore_EmptyText(t *testing.T) {
	store := NewCodeStore()
	input := ""
	masked := store.Mask(input)
	if masked != "" {
		t.Fatalf("expected empty, got %q", masked)
	}
	restored := store.Restore(masked)
	if restored != "" {
		t.Fatalf("expected empty, got %q", restored)
	}
}

func TestCodeStore_NoCode(t *testing.T) {
	store := NewCodeStore()
	input := "обычный текст без кода"
	masked := store.Mask(input)
	if masked != input {
		t.Fatalf("expected no change, got %q", masked)
	}
	restored := store.Restore(masked)
	if restored != input {
		t.Fatalf("expected %q, got %q", input, restored)
	}
}
