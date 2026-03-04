package rag

type Document struct {
	ID      string
	Path    string
	Title   string
	Content string
	tokens  []string
}

type SearchResult struct {
	Path    string  `json:"path"`
	Title   string  `json:"title"`
	Score   float64 `json:"score"`
	Excerpt string  `json:"excerpt"`
}
