package rag_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ygncode/meta-cli/internal/rag"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{"empty", "", nil},
		{"single char tokens dropped", "a b c", nil},
		{"basic words", "hello world", []string{"hello", "world"}},
		{"lowercased", "Hello World", []string{"hello", "world"}},
		{"digits kept", "go 1.21 release", []string{"go", "21", "release"}},
		{"punctuation splits", "foo-bar_baz.qux", []string{"foo", "bar", "baz", "qux"}},
		{"unicode letters", "café résumé", []string{"café", "résumé"}},
		{"mixed", "Hello, World! 123 test.", []string{"hello", "world", "123", "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rag.Tokenize(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("len: got %d, want %d\ngot:  %v\nwant: %v", len(got), len(tt.expect), got, tt.expect)
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("token[%d]: got %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestChunkByHeading(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantCount  int
		wantTitles []string
	}{
		{
			name:       "no headings",
			content:    "just some text\nmore text",
			wantCount:  1,
			wantTitles: []string{""}, // no heading prefix, title stays empty; "document" fallback only when buf is empty
		},
		{
			name:       "single h1",
			content:    "# Title\nsome body text",
			wantCount:  1,
			wantTitles: []string{"Title"},
		},
		{
			name:       "multiple headings",
			content:    "# First\nbody1\n## Second\nbody2\n# Third\nbody3",
			wantCount:  3,
			wantTitles: []string{"First", "Second", "Third"},
		},
		{
			name:       "heading without body ignored until next heading adds it",
			content:    "# A\n# B\ntext",
			wantCount:  1,
			wantTitles: []string{"B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := rag.ExportChunkByHeading(tt.content)
			if len(chunks) != tt.wantCount {
				t.Fatalf("chunk count: got %d, want %d", len(chunks), tt.wantCount)
			}
			for i, title := range tt.wantTitles {
				if chunks[i].Title != title {
					t.Errorf("chunk[%d].Title: got %q, want %q", i, chunks[i].Title, title)
				}
			}
		})
	}
}

func TestBuildAndSearch(t *testing.T) {
	docs := []rag.Document{
		{ID: "1", Path: "a.md", Title: "Go Programming", Content: "Go is a programming language designed at Google"},
		{ID: "2", Path: "b.md", Title: "Python Programming", Content: "Python is a programming language known for readability"},
		{ID: "3", Path: "c.md", Title: "Cooking Recipes", Content: "Learn how to make pasta and pizza from scratch"},
	}

	idx := rag.Build(docs)

	t.Run("relevant query", func(t *testing.T) {
		results := idx.Search("programming language", 2)
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		// Both programming docs should match
		for _, r := range results {
			if r.Score <= 0 {
				t.Errorf("expected positive score, got %f", r.Score)
			}
		}
	})

	t.Run("no match", func(t *testing.T) {
		results := idx.Search("quantum physics", 3)
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("empty query", func(t *testing.T) {
		results := idx.Search("", 3)
		if results != nil {
			t.Errorf("expected nil, got %v", results)
		}
	})

	t.Run("nil index", func(t *testing.T) {
		var nilIdx *rag.Index
		results := nilIdx.Search("test", 3)
		if results != nil {
			t.Errorf("expected nil, got %v", results)
		}
	})

	t.Run("title boost", func(t *testing.T) {
		// "programming" appears in content of both Go and Python docs.
		// "Go Programming" title contains "programming", so title boost should rank it higher.
		results := idx.Search("programming", 3)
		if len(results) < 2 {
			t.Fatal("expected at least 2 results for 'programming'")
		}
		// Both programming docs should have positive scores
		for _, r := range results[:2] {
			if r.Score <= 0 {
				t.Errorf("expected positive score, got %f for %s", r.Score, r.Path)
			}
		}
	})

	t.Run("topK capped", func(t *testing.T) {
		results := idx.Search("programming", 100)
		if len(results) > len(docs) {
			t.Errorf("results should not exceed doc count")
		}
	})

	t.Run("excerpt truncated", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 100; i++ {
			longContent += "programming language design "
		}
		longDocs := []rag.Document{
			{ID: "long", Path: "long.md", Title: "Long", Content: longContent},
		}
		longIdx := rag.Build(longDocs)
		results := longIdx.Search("programming", 1)
		if len(results) == 1 && len(results[0].Excerpt) > 310 {
			t.Errorf("excerpt should be truncated, got len %d", len(results[0].Excerpt))
		}
	})
}

func TestBuildEmptyDocs(t *testing.T) {
	idx := rag.Build(nil)
	results := idx.Search("anything", 5)
	if results != nil {
		t.Errorf("expected nil for empty index, got %v", results)
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(dir, "doc1.md"), []byte("# Hello\nworld"), 0o644)
	os.WriteFile(filepath.Join(dir, "doc2.txt"), []byte("plain text content"), 0o644)
	os.WriteFile(filepath.Join(dir, "ignored.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(dir, "empty.md"), []byte("   "), 0o644)

	// Create subdirectory with file
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "nested.md"), []byte("# Nested\ncontent"), 0o644)

	docs, err := rag.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir error: %v", err)
	}

	// Should load: doc1.md (1 chunk), doc2.txt (1 chunk, no heading), sub/nested.md (1 chunk)
	// Should skip: ignored.go, empty.md
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	// Verify paths are relative
	for _, d := range docs {
		if filepath.IsAbs(d.Path) {
			t.Errorf("expected relative path, got %s", d.Path)
		}
	}
}

func TestLoadDirNotExist(t *testing.T) {
	_, err := rag.LoadDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
