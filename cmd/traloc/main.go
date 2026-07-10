package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/artixzenevich/traloc/internal/app"
	"github.com/artixzenevich/traloc/internal/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Показать версию и выйти")
	cfgPath := flag.String("config", "", "Путь к конфигурационному файлу")
	inputPattern := flag.String("in", "", "Входные файлы (паттерн или список)")
	outputDir := flag.String("outdir", "", "Директория для выходных файлов")
	cacheDir := flag.String("cache", "", "Путь к кэшу переводов")
	model := flag.String("model", "", "Ollama модель (переопределяет конфиг)")
	src := flag.String("from", "English", "Исходный язык")
	tgt := flag.String("to", "Russian", "Целевой язык")
	chunkSize := flag.Int("chunk", 0, "Максимум токенов на чанк (переопределяет конфиг)")
	keepOllama := flag.Bool("keep-ollama", false, "Не останавливать Ollama после работы")
	flag.Parse()

	if *showVersion {
		fmt.Printf("traloc %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	cfg := loadConfig(*cfgPath, *inputPattern, *outputDir, *cacheDir, *model, *src, *tgt, *chunkSize, *keepOllama)

	files := resolveFiles(*inputPattern, flag.Args())
	app.Run(cfg, files)
}

func loadConfig(cfgPath, inputPattern, outputDir, cacheDir, model, src, tgt string, chunkSize int, keepOllama bool) *config.Config {
	var cfg *config.Config

	if cfgPath == "" {
		cfgPath = config.FindConfig()
	}

	if cfgPath != "" {
		var err error
		cfg, err = config.Load(cfgPath)
		if err != nil {
			log.Warnf("Не удалось загрузить конфиг: %v, используем настройки по умолчанию", err)
			cfg = config.Default()
		}
	} else {
		cfg = config.Default()
	}

	cfg.InputPattern = inputPattern
	cfg.OutputDir = outputDir
	cfg.SourceLang = src
	cfg.TargetLang = tgt
	cfg.KeepOllama = keepOllama

	if model != "" {
		cfg.Model = model
	}
	if chunkSize > 0 {
		cfg.ChunkSize = chunkSize
	}
	if cacheDir != "" {
		cfg.CacheDir = cacheDir
	}

	if cfg.InputPattern == "" {
		log.Fatal("Укажите -in <pattern|files>")
		os.Exit(1)
	}

	return cfg
}

func resolveFiles(pattern string, extra []string) []string {
	var files []string

	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		var err error
		files, err = filepath.Glob(pattern)
		if err != nil {
			log.Fatalf("Ошибка паттерна: %v", err)
		}
	} else if pattern != "" {
		files = []string{pattern}
	}

	files = append(files, extra...)
	return files
}
