package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadDir(dir string) ([]Document, error) {
	var docs []Document
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		content := string(data)
		if strings.TrimSpace(content) == "" {
			return nil
		}

		chunks := chunkByHeading(content)
		relPath, _ := filepath.Rel(dir, path)
		if relPath == "" {
			relPath = path
		}

		for i, chunk := range chunks {
			docs = append(docs, Document{
				ID:      fmt.Sprintf("%s#%d", relPath, i),
				Path:    relPath,
				Title:   chunk.title,
				Content: chunk.content,
			})
		}
		return nil
	})
	return docs, err
}

type chunk struct {
	title   string
	content string
}

func chunkByHeading(content string) []chunk {
	lines := strings.Split(content, "\n")
	var chunks []chunk
	var current chunk
	var buf []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
			if len(buf) > 0 {
				current.content = strings.Join(buf, "\n")
				chunks = append(chunks, current)
			}
			title := strings.TrimPrefix(trimmed, "## ")
			if title == trimmed {
				title = strings.TrimPrefix(trimmed, "# ")
			}
			current = chunk{title: title}
			buf = nil
		} else {
			buf = append(buf, line)
		}
	}
	if len(buf) > 0 {
		current.content = strings.Join(buf, "\n")
		chunks = append(chunks, current)
	}

	if len(chunks) == 0 {
		chunks = append(chunks, chunk{
			title:   "document",
			content: content,
		})
	}
	return chunks
}
