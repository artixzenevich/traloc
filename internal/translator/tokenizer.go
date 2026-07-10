package translator

import (
	"strings"
	"unicode/utf8"
)

func EstimateTokens(text string) int {
	chars := utf8.RuneCountInString(text)
	nonASCII := 0
	for _, r := range text {
		if r > 127 {
			nonASCII++
		}
	}
	latinChars := chars - nonASCII
	tokens := float64(latinChars)/4.0 + float64(nonASCII)/2.0
	return int(tokens) + 1
}

func SplitIntoChunks(text string, maxTokens int) []string {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var current strings.Builder
	currentTokens := 0

	flush := func() {
		if current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
			currentTokens = 0
		}
	}

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		paraTokens := EstimateTokens(para)

		if paraTokens > maxTokens {
			flush()
			chunks = append(chunks, splitBySentences(para, maxTokens)...)
			continue
		}

		if currentTokens+paraTokens > maxTokens {
			flush()
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
			currentTokens += 2
		}
		current.WriteString(para)
		currentTokens += paraTokens
	}
	flush()
	return chunks
}

func splitBySentences(text string, maxTokens int) []string {
	var sentences []string
	var current strings.Builder
	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' || r == '\n' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}

	var chunks []string
	var buf strings.Builder
	bufTokens := 0
	for _, s := range sentences {
		t := EstimateTokens(s)
		if bufTokens+t > maxTokens && buf.Len() > 0 {
			chunks = append(chunks, buf.String())
			buf.Reset()
			bufTokens = 0
		}
		if buf.Len() > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(s)
		bufTokens += t
	}
	if buf.Len() > 0 {
		chunks = append(chunks, buf.String())
	}
	return chunks
}
