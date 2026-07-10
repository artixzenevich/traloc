package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ollama/ollama/api"
	"github.com/schollz/progressbar/v3"

	"github.com/artixzenevich/traloc/internal/config"
	"github.com/artixzenevich/traloc/internal/ollama"
	"github.com/artixzenevich/traloc/internal/styles"
	"github.com/artixzenevich/traloc/internal/translator"
)

func Run(cfg *config.Config, files []string) {
	cfg.Resolve()

	if len(files) == 0 {
		log.Fatalf("Файлы не найдены: %s", cfg.InputPattern)
	}

	log.Infof("Найдено %d файлов для перевода", len(files))
	for i, f := range files {
		log.Infof("  [%d] %s", i+1, f)
	}

	ollamaCmd, err := ollama.StartIfNeeded()
	if err != nil {
		log.Fatalf("Ошибка запуска Ollama: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Warn("Получен сигнал завершения, очищаем ресурсы...")
		cancel()
	}()

	defer func() {
		if !cfg.KeepOllama {
			ollama.UnloadModel(cfg.Model)
			ollama.Stop(ollamaCmd)
		} else {
			log.Info("Режим -keep-ollama: Ollama и модель остаются в памяти")
		}
	}()

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("Ollama client: %v", err)
	}

	totalCacheHits := 0
	totalCacheMisses := 0
	totalChunks := 0
	successCount := 0
	startTime := time.Now()

	for fileIdx, inputFile := range files {
		if ctx.Err() != nil {
			log.Warn("Обработка прервана")
			break
		}

		var outputFile string
		if cfg.OutputDir != "" {
			if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
				log.Warnf("Не удалось создать директорию %s: %v", cfg.OutputDir, err)
			}
			baseName := filepath.Base(inputFile)
			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			outputFile = filepath.Join(cfg.OutputDir, fmt.Sprintf("%s.%s%s", nameWithoutExt, cfg.LangCode(cfg.TargetLang), ext))
		} else {
			ext := filepath.Ext(inputFile)
			base := strings.TrimSuffix(inputFile, ext)
			outputFile = fmt.Sprintf("%s.%s%s", base, cfg.LangCode(cfg.TargetLang), ext)
		}

		// Пропускаем уже переведённые файлы (*.ru.md, *.de.md и т.д.)
		if isAlreadyTranslated(inputFile, cfg.TargetLang, cfg) {
			log.Infof("Пропускаем (уже переведён): %s", inputFile)
			continue
		}

		fmt.Println(styles.Separator)
		fmt.Println(styles.FileHeader(fileIdx+1, len(files), inputFile, outputFile))

		data, err := os.ReadFile(inputFile)
		if err != nil {
			log.Errorf("Не могу прочитать %s: %v", inputFile, err)
			continue
		}
		text := string(data)

		if len(text) == 0 {
			log.Warnf("Файл пустой, пропускаем: %s", inputFile)
			continue
		}

		log.Infof("Файл: %d символов, ~%d токенов", len(text), translator.EstimateTokens(text))

		chunks := translator.SplitIntoChunks(text, cfg.ChunkSize)
		log.Infof("Разбито на %d чанков", len(chunks))

		var translated []string
		fileCacheHits := 0
		fileCacheMisses := 0
		fileStartTime := time.Now()

		chunkBar := progressbar.NewOptions(len(chunks),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionShowIts(),
			progressbar.OptionSetDescription("Чанки"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionClearOnFinish(),
		)

		for i, chunk := range chunks {
			if ctx.Err() != nil {
				break
			}

			key := translator.CacheKey(cfg.Model, cfg.SourceLang, cfg.TargetLang, chunk)
			if cached, ok := translator.GetCache(cfg.CacheDir, key); ok {
				translated = append(translated, cached)
				fileCacheHits++
				chunkBar.Add(1)
				continue
			}

			res, err := translator.TranslateChunk(ctx, client, cfg.Model, chunk, cfg.SourceLang, cfg.TargetLang)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				log.Errorf("Ошибка в чанке %d: %v", i+1, err)
				res = fmt.Sprintf("[ERROR chunk %d: %v]", i+1, err)
			} else {
				translator.SetCache(cfg.CacheDir, key, res)
			}
			translated = append(translated, res)
			fileCacheMisses++
			chunkBar.Add(1)
		}

		chunkBar.Finish()
		fmt.Println()

		if len(translated) > 0 {
			final := strings.Join(translated, "\n\n")
			if err := os.WriteFile(outputFile, []byte(final), 0644); err != nil {
				log.Errorf("Ошибка записи %s: %v", outputFile, err)
			} else {
				log.Infof("Сохранено: %s", outputFile)
				successCount++
			}
		}

		fileDuration := time.Since(fileStartTime)
		totalCacheHits += fileCacheHits
		totalCacheMisses += fileCacheMisses
		totalChunks += len(chunks)

		infoLines := []string{
			fmt.Sprintf("чанков: %d  кэш: %d  перевод: %d  время: %s",
				len(chunks), fileCacheHits, fileCacheMisses, fileDuration.Round(time.Second)),
		}
		fmt.Println(styles.StatBox("Статистика файла", infoLines...))
	}

	totalDuration := time.Since(startTime)
	hitRate := 0.0
	if totalChunks > 0 {
		hitRate = float64(totalCacheHits) / float64(totalChunks) * 100
	}
	estimatedSaved := time.Duration(totalCacheHits) * 2 * time.Minute

	fmt.Println(styles.Separator)
	log.Infof("Готово! Успешно переведено %d из %d файлов", successCount, len(files))
	log.Infof("Всего чанков: %d", totalChunks)
	log.Infof("Из кэша: %d", totalCacheHits)
	log.Infof("Переведено: %d", totalCacheMisses)
	log.Infof("Процент попаданий: %.1f%%", hitRate)
	log.Infof("Общее время: %s", totalDuration.Round(time.Second))
	if estimatedSaved > 0 {
		log.Infof("Экономия (примерно): %s", estimatedSaved.Round(time.Minute))
	}
}

func isAlreadyTranslated(filePath, targetLang string, cfg *config.Config) bool {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), ext)
	return strings.HasSuffix(base, "."+cfg.LangCode(targetLang))
}
