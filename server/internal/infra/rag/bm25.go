package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

// LexicalTokens keeps identifiers as words and indexes Chinese runs as bigrams.
// The same analysis is used for documents and queries, including repeated terms.
func LexicalTokens(source string) []string {
	var result []string
	var word, han []rune
	flush := func() {
		if len(word) > 0 {
			result = append(result, string(word))
			word = nil
		}
		if len(han) == 1 {
			result = append(result, string(han))
		}
		for i := 1; i < len(han); i++ {
			result = append(result, string(han[i-1:i+1]))
		}
		han = nil
	}
	for _, r := range strings.ToLower(source) {
		if unicode.Is(unicode.Han, r) {
			if len(word) > 0 {
				flush()
			}
			han = append(han, r)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_+#.-", r) {
			if len(han) > 0 {
				flush()
			}
			word = append(word, r)
		} else {
			flush()
		}
	}
	flush()
	return result
}

// BM25 ranks one live corpus for all queries, using actual TF/DF and average length.
// No external search service or persistent cache can bypass source visibility.
func BM25(corpus []domain.Evidence, questions []string, tuning domain.Tuning) [][]domain.Evidence {
	type document struct {
		frequencies map[string]int
		length      int
	}
	docs := make([]document, len(corpus))
	df := map[string]int{}
	totalLength := 0
	for i, item := range corpus {
		terms := LexicalTokens(item.Title + "\n" + item.ContextHeader + "\n" + item.Content)
		frequencies := map[string]int{}
		for _, term := range terms {
			frequencies[term]++
		}
		for term := range frequencies {
			df[term]++
		}
		docs[i] = document{frequencies, len(terms)}
		totalLength += len(terms)
	}
	results := make([][]domain.Evidence, len(questions))
	if totalLength == 0 {
		return results
	}
	average := float64(totalLength) / float64(len(corpus))
	for q, question := range questions {
		query := map[string]bool{}
		for _, term := range LexicalTokens(question) {
			query[term] = true
		}
		for i, item := range corpus {
			score := 0.0
			for term := range query {
				tf := float64(docs[i].frequencies[term])
				if tf == 0 {
					continue
				}
				idf := math.Log(1 + (float64(len(corpus)-df[term])+0.5)/(float64(df[term])+0.5))
				score += idf * tf * (tuning.BM25K1 + 1) / (tf + tuning.BM25K1*(1-tuning.BM25B+tuning.BM25B*float64(docs[i].length)/average))
			}
			if score > 0 {
				item.Score = score
				results[q] = append(results[q], item)
			}
		}
		sort.Slice(results[q], func(i, j int) bool {
			if results[q][i].Score == results[q][j].Score {
				return results[q][i].ID < results[q][j].ID
			}
			return results[q][i].Score > results[q][j].Score
		})
		results[q] = results[q][:min(len(results[q]), tuning.KeywordTopK)]
	}
	return results
}
