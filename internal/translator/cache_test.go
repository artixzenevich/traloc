package translator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheKey(t *testing.T) {
	k1 := CacheKey("m", "en", "ru", "hello")
	k2 := CacheKey("m", "en", "ru", "hello")
	k3 := CacheKey("m", "en", "ru", "world")

	if k1 != k2 {
		t.Fatal("same input must produce same key")
	}
	if k1 == k3 {
		t.Fatal("different input must produce different key")
	}
}

func TestCacheKey_Length(t *testing.T) {
	k := CacheKey("model", "English", "Russian", "some text")
	if len(k) != 16 {
		t.Fatalf("expected key length 16, got %d", len(k))
	}
}

func TestSetGetCache(t *testing.T) {
	dir := t.TempDir()
	key := "testkey123"
	expected := "translated text"

	SetCache(dir, key, expected)
	got, ok := GetCache(dir, key)

	if !ok {
		t.Fatal("cache miss after set")
	}
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestGetCache_Miss(t *testing.T) {
	dir := t.TempDir()
	_, ok := GetCache(dir, "nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestSetCache_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	SetCache(dir, "key", "data")

	path := filepath.Join(dir, "key.txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("cache directory was not created")
	}
}
