package translator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	codeBlockRe  = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\n.*?\n```")
	inlineCodeRe = regexp.MustCompile("`[^`\n]+`")
)

type CodeStore struct {
	counter int
	blocks  map[string]string
}

func NewCodeStore() *CodeStore {
	return &CodeStore{
		blocks: make(map[string]string),
	}
}

func (s *CodeStore) Mask(text string) string {
	result := codeBlockRe.ReplaceAllStringFunc(text, func(match string) string {
		s.counter++
		key := fmt.Sprintf("\n@@KEEP_BLOCK_%d@@\n", s.counter)
		s.blocks[strings.TrimSpace(key)] = match
		return key
	})

	result = inlineCodeRe.ReplaceAllStringFunc(result, func(match string) string {
		s.counter++
		key := fmt.Sprintf("@@KEEP_INLINE_%d@@", s.counter)
		s.blocks[key] = match
		return key
	})

	return result
}

func (s *CodeStore) Restore(text string) string {
	keys := make([]string, 0, len(s.blocks))
	for k := range s.blocks {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, key := range keys {
		text = strings.ReplaceAll(text, key, s.blocks[key])
	}
	return text
}
