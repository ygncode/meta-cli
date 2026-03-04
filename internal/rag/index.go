package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

type Index struct {
	docs  []Document
	tfidf map[int]map[string]float64
	idf   map[string]float64
}

func Build(docs []Document) *Index {
	idx := &Index{
		docs:  docs,
		tfidf: make(map[int]map[string]float64),
		idf:   make(map[string]float64),
	}

	for i := range idx.docs {
		idx.docs[i].tokens = tokenize(idx.docs[i].Content)
	}

	df := make(map[string]int)
	for _, doc := range idx.docs {
		seen := make(map[string]bool)
		for _, t := range doc.tokens {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}

	n := float64(len(idx.docs))
	for term, count := range df {
		idx.idf[term] = math.Log(1 + n/float64(count))
	}

	for i, doc := range idx.docs {
		tf := make(map[string]int)
		for _, t := range doc.tokens {
			tf[t]++
		}
		tfidf := make(map[string]float64)
		total := float64(len(doc.tokens))
		if total == 0 {
			total = 1
		}
		for term, count := range tf {
			tfidf[term] = (float64(count) / total) * idx.idf[term]
		}
		idx.tfidf[i] = tfidf
	}

	return idx
}

func (idx *Index) Search(query string, topK int) []SearchResult {
	if idx == nil || len(idx.docs) == 0 {
		return nil
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	type scored struct {
		index int
		score float64
	}
	var scores []scored

	for i, doc := range idx.docs {
		var score float64
		for _, qt := range queryTokens {
			if w, ok := idx.tfidf[i][qt]; ok {
				score += w
			}
		}
		if score > 0 {
			titleLower := strings.ToLower(doc.Title)
			for _, qt := range queryTokens {
				if strings.Contains(titleLower, qt) {
					score *= 1.5
				}
			}
			scores = append(scores, scored{index: i, score: score})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]SearchResult, 0, topK)
	for _, s := range scores[:topK] {
		doc := idx.docs[s.index]
		excerpt := doc.Content
		if len(excerpt) > 300 {
			excerpt = excerpt[:300] + "..."
		}
		results = append(results, SearchResult{
			Path:    doc.Path,
			Title:   doc.Title,
			Score:   s.score,
			Excerpt: excerpt,
		})
	}
	return results
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var buf strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else if buf.Len() > 0 {
			t := buf.String()
			if len(t) > 1 {
				tokens = append(tokens, t)
			}
			buf.Reset()
		}
	}
	if buf.Len() > 1 {
		tokens = append(tokens, buf.String())
	}
	return tokens
}
