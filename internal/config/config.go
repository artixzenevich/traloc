package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Languages    map[string]string `yaml:"languages"`
	DefaultModel string            `yaml:"default_model"`
	ChunkSize    int               `yaml:"chunk_size"`
	CacheDir     string            `yaml:"cache_dir"`

	// Параметры из флагов (поверх конфига)
	InputPattern string
	OutputDir    string
	Model        string
	SourceLang   string
	TargetLang   string
	KeepOllama   bool
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать конфиг %s: %w", path, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфига: %w", err)
	}

	return cfg, nil
}

func Default() *Config {
	return &Config{
		Languages: map[string]string{
			"english":    "en",
			"russian":    "ru",
			"german":     "de",
			"french":     "fr",
			"spanish":    "es",
			"italian":    "it",
			"portuguese": "pt",
			"chinese":    "zh",
			"japanese":   "ja",
			"korean":     "ko",
			"ukrainian":  "uk",
			"polish":     "pl",
			"dutch":      "nl",
			"turkish":    "tr",
			"arabic":     "ar",
		},
		DefaultModel: "translategemma:4b",
		ChunkSize:    3500,
		CacheDir:     ".translate-cache",
	}
}

func (c *Config) LangCode(lang string) string {
	if code, ok := c.Languages[strings.ToLower(lang)]; ok {
		return code
	}
	return strings.ToLower(lang)
}

func (c *Config) Resolve() {
	if c.Model == "" {
		c.Model = c.DefaultModel
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = 3500
	}
	if c.CacheDir == "" {
		c.CacheDir = ".translate-cache"
	}
}

func FindConfig() string {
	// 1. Флаг -config
	if cfg := flag.Lookup("config"); cfg != nil {
		if v := cfg.Value.String(); v != "" {
			return v
		}
	}
	// 2. .traloc.yaml в текущей директории
	for _, name := range []string{".traloc.yaml", ".traloc.yml"} {
		if _, err := os.Stat(name); err == nil {
			abs, _ := filepath.Abs(name)
			return abs
		}
	}
	return ""
}
