package rag

// Export unexported functions for testing.
var Tokenize = tokenize

// ChunkResult holds exported chunk data for testing.
type ChunkResult struct {
	Title   string
	Content string
}

// ExportChunkByHeading wraps the unexported chunkByHeading and returns exported types.
func ExportChunkByHeading(content string) []ChunkResult {
	chunks := chunkByHeading(content)
	results := make([]ChunkResult, len(chunks))
	for i, c := range chunks {
		results[i] = ChunkResult{Title: c.title, Content: c.content}
	}
	return results
}
