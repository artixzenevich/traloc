package translator

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ollama/ollama/api"
	"github.com/schollz/progressbar/v3"
)

func TranslateChunk(ctx context.Context, client *api.Client, model, text, src, tgt string) (string, error) {
	store := NewCodeStore()
	maskedText := store.Mask(text)

	prompt := fmt.Sprintf(
		`Translate the following text from %s to %s.
Rules:
- Preserve Markdown formatting (headings, lists, links, bold, italic).
- The text contains placeholders like @@KEEP_BLOCK_N@@ and @@KEEP_INLINE_N@@.
- DO NOT translate or modify these placeholders — copy them exactly as-is into the output.
- Output ONLY the translation, no explanations.

%s`,
		src, tgt, maskedText,
	)

	var result strings.Builder
	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: ptr(true),
		Options: map[string]any{
			"num_predict": 4096,
			"num_ctx":     4096,
			"temperature": 0.3,
		},
	}

	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Генерация"),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("ток/с"),
		progressbar.OptionClearOnFinish(),
	)

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		result.WriteString(resp.Response)
		bar.Add(1)
		return nil
	})

	bar.Finish()
	fmt.Println()

	if err != nil {
		return "", err
	}

	translated := store.Restore(result.String())
	log.Infof("Переведено символов: %d -> %d", len(text), len(translated))
	return translated, nil
}

func ptr[T any](v T) *T {
	return &v
}
