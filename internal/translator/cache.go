package translator

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

func CacheKey(model, src, tgt, text string) string {
	h := sha256.New()
	h.Write([]byte(model + "|" + src + "|" + tgt + "|" + text))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func GetCache(cacheDir, key string) (string, bool) {
	path := filepath.Join(cacheDir, key+".txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func SetCache(cacheDir, key, text string) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return
	}
	os.WriteFile(filepath.Join(cacheDir, key+".txt"), []byte(text), 0644)
}
