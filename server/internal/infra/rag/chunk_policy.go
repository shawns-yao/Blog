package rag

import (
	"strings"
	"unicode"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	"github.com/yuin/goldmark/ast"
)

// Profile uses structural facts from the existing Markdown parse, not an LLM label.
type DocumentProfile struct {
	Tokens           int
	Headings         int
	Paragraphs       int
	CodeBlocks       int
	Tables           int
	StructuredTokens int
}

// A forced split may retain a bounded tail at a word/sentence boundary. Natural
// paragraphs, headings and protected structures never receive this text overlap.
func forcedOverlapTail(source string, item unit, budget int) (unit, bool) {
	if !item.forced || item.kind != "text" || budget <= 0 {
		return unit{}, false
	}
	fragment := source[item.start:item.end]
	runes := []rune(fragment)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi) / 2
		if CountTokens(string(runes[mid:])) > budget {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	for i := max(1, lo); i < len(runes); i++ {
		previous := runes[i-1]
		if !unicode.IsSpace(previous) && !strings.ContainsRune("。！？；", previous) {
			continue
		}
		tail := string(runes[i:])
		if strings.TrimSpace(tail) == "" || CountTokens(tail) > budget {
			continue
		}
		item.start += len(string(runes[:i]))
		return item, true
	}
	return unit{}, false
}

func profileMarkdown(markdown string, doc ast.Node) DocumentProfile {
	p := DocumentProfile{Tokens: CountTokens(markdown)}
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node.Kind() {
		case ast.KindHeading:
			p.Headings++
		case ast.KindParagraph:
			p.Paragraphs++
		case ast.KindFencedCodeBlock, ast.KindCodeBlock:
			p.CodeBlocks++
		}
		if node.Kind().String() == "Table" {
			p.Tables++
		}
		if node.Kind() == ast.KindFencedCodeBlock || node.Kind() == ast.KindCodeBlock || node.Kind().String() == "Table" {
			start, end := nodeRange(node, markdown)
			if start >= 0 && end > start {
				p.StructuredTokens += CountTokens(markdown[start:end])
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return p
}

func selectChunkPolicy(title string, p DocumentProfile, configured domain.Tuning) (domain.Tuning, string) {
	t := configured
	if !t.AdaptiveChunkingEnabled {
		return t, "configured"
	}
	switch {
	case p.Headings <= 1 && p.Tokens+CountTokens(title)+2 <= t.ChunkMaxTokens:
		t.ChunkTargetTokens, t.ChunkOverlapTokens = t.ChunkMaxTokens, 0
		return t, "short_document"
	case p.StructuredTokens*3 >= p.Tokens && p.StructuredTokens > 0:
		t.ChunkTargetTokens = max(t.ChunkMinTokens, t.ChunkTargetTokens*3/4)
		t.ChunkOverlapTokens = 0
		return t, "code_table"
	case p.Headings > 1:
		t.ChunkOverlapTokens = 0
		return t, "heading_structure"
	case p.Paragraphs > 1:
		t.ChunkTargetTokens = min(t.ChunkMaxTokens, t.ChunkTargetTokens*5/4)
		return t, "narrative"
	default:
		return t, "continuous_text"
	}
}

// Policies cannot hide a missing range, mutate source text or bypass the hard ceiling.
func validPolicyChunks(markdown string, chunks []domain.Chunk, tuning domain.Tuning) bool {
	source := []rune(markdown)
	covered := 0
	for _, chunk := range chunks {
		if chunk.Start < 0 || chunk.End > len(source) || chunk.End <= chunk.Start ||
			chunk.Content != string(source[chunk.Start:chunk.End]) || chunk.Tokens > ChunkTokenLimit(chunk.Kind, tuning.ChunkMaxTokens) {
			return false
		}
		if chunk.Start > covered && strings.TrimSpace(string(source[covered:chunk.Start])) != "" {
			return false
		}
		covered = max(covered, chunk.End)
	}
	return len(chunks) > 0 && strings.TrimSpace(string(source[covered:])) == ""
}
