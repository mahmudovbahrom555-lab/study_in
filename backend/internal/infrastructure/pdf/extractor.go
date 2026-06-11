// Package pdf extracts plain text from PDF files for AI processing.
package pdf

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

// ExtractText reads the PDF at path and returns concatenated page text.
// Returns a best-effort result — partial text is returned alongside an error
// when some pages fail.
func ExtractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("pdf.Open: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	numPages := r.NumPage()
	var lastErr error

	for i := 1; i <= numPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			lastErr = err
			continue
		}
		sb.WriteString(text)
		sb.WriteByte('\n')
	}

	result := cleanText(sb.String())
	if result == "" && lastErr != nil {
		return "", fmt.Errorf("pdf extraction failed: %w", lastErr)
	}
	return result, lastErr
}

// ExtractTextFromBytes writes data to a temp file, extracts, and removes the file.
func ExtractTextFromBytes(data []byte) (string, error) {
	tmp, err := os.CreateTemp("", "repetapp-pdf-*.pdf")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return "", fmt.Errorf("write temp pdf: %w", err)
	}
	tmp.Close()

	return ExtractText(tmp.Name())
}

// cleanText normalises whitespace while preserving paragraph breaks.
func cleanText(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		l = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) && r != '\n' && r != '\t' {
				return ' '
			}
			return r
		}, l)
		l = strings.Join(strings.Fields(l), " ")
		out = append(out, l)
	}
	// Collapse 3+ blank lines into 2.
	result := strings.Join(out, "\n")
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result)
}

// ChunkText splits text into overlapping word-based chunks.
// chunkSize and overlap are in words.
func ChunkText(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = 400
	}
	if overlap < 0 || overlap >= chunkSize {
		overlap = 50
	}

	var chunks []string
	step := chunkSize - overlap
	for i := 0; i < len(words); i += step {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
		if end == len(words) {
			break
		}
	}
	return chunks
}
