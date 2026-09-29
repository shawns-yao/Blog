// Recursive splitting and heading context are adapted from Tencent/WeKnora.
// See licenses/WeKnora-MIT.txt for the original license and source references.
package rag

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

const ChunkerVersion = "weknora-markdown-token-parent-v2"

type section struct {
	start  int
	header string
}

type span struct {
	start, end int
	kind       string
}

type unit struct {
	start, end int
	kind       string
	header     string
}

var protectedPatterns = []struct {
	pattern *regexp.Regexp
	kind    string
}{
	{regexp.MustCompile("(?ms)^[ \\t]{0,3}`{3,}[^\\n]*\\n.*?^[ \\t]{0,3}`{3,}[ \\t\\r]*(?:\\n|$)"), "code"},
	{regexp.MustCompile("(?ms)^[ \\t]{0,3}~{3,}[^\\n]*\\n.*?^[ \\t]{0,3}~{3,}[ \\t\\r]*(?:\\n|$)"), "code"},
	{regexp.MustCompile(`(?s)\$\$.*?\$\$`), "math"},
	{regexp.MustCompile(`!\[[^\]\n]{0,200}\]\([^)\n]{1,500}\)`), "image_alt"},
	{regexp.MustCompile(`\[[^\]\n]{1,200}\]\([^)\n]{1,500}\)`), "text"},
	{regexp.MustCompile("`[^`\\r\\n]+`"), "code"},
	{regexp.MustCompile(`(?m)^[^\n]*\|[^\n]*\r?\n[ \t]*\|?[ \t]*:?-{3,}[^\n]*\|[^\n]*(?:\r?\n[^\n]*\|[^\n]*)*`), "table"},
}

// SplitMarkdown retains source text verbatim; offsets count Unicode runes.
// Headings and repeated table headers stay outside Content so offsets remain valid.
func SplitMarkdown(title, markdown string, size, overlap int) []domain.Chunk {
	return SplitMarkdownWithTuning(title, markdown, domain.Tuning{ChunkTargetTokens: size,
		ChunkMaxTokens: max(800, size), ChunkMinTokens: 180, ChunkOverlapTokens: overlap, ParentMaxTokens: 1600})
}

func SplitMarkdownWithTuning(title, markdown string, tuning domain.Tuning) []domain.Chunk {
	size, overlap := tuning.ChunkTargetTokens, tuning.ChunkOverlapTokens
	if size < 100 {
		size = 500
	}
	maximum := max(size, tuning.ChunkMaxTokens)
	if overlap < 0 || overlap >= size {
		overlap = 0
	}
	source := []byte(markdown)
	doc := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(source))
	sections := []section{{header: strings.TrimSpace(title)}}
	var headings [6]string
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		heading, ok := node.(*ast.Heading)
		if !entering || !ok || heading.Parent() != doc || heading.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		level := heading.Level
		headings[level-1] = strings.TrimSpace(string(heading.Text(source)))
		for i := level; i < len(headings); i++ {
			headings[i] = ""
		}
		path := []string{strings.TrimSpace(title)}
		for _, item := range headings {
			if item != "" && (len(path) == 0 || path[len(path)-1] != item) {
				path = append(path, item)
			}
		}
		start := heading.Lines().At(0).Start
		for start > 0 && source[start-1] != '\n' {
			start--
		}
		header := strings.Join(path, " > ")
		if start == sections[len(sections)-1].start {
			sections[len(sections)-1].header = header
		} else {
			sections = append(sections, section{start: start, header: header})
		}
		return ast.WalkContinue, nil
	})

	runeOffsets := make([]int, len(source)+1)
	runeOffset := 0
	for pos, r := range markdown {
		for i := 0; i < utf8.RuneLen(r); i++ {
			runeOffsets[pos+i] = runeOffset
		}
		runeOffset++
	}
	runeOffsets[len(source)] = runeOffset
	chunks := make([]domain.Chunk, 0)
	for i, section := range sections {
		end := len(source)
		if i+1 < len(sections) {
			end = sections[i+1].start
		}
		units := splitSection(markdown[section.start:end], max(1, maximum-CountTokens(section.header)-2))
		sectionStart := len(chunks)
		var current []unit
		currentSize := 0
		headerFor := func(items []unit) string {
			header := section.header
			for _, item := range items {
				if item.header != "" && !strings.Contains(header, item.header) {
					header += "\n" + item.header
				}
			}
			return header
		}
		flush := func() {
			if len(current) == 0 {
				return
			}
			start, end := current[0].start+section.start, current[len(current)-1].end+section.start
			if strings.TrimSpace(markdown[start:end]) == "" {
				return
			}
			kind, header := current[0].kind, headerFor(current)
			for _, item := range current {
				if item.kind != kind {
					kind = "text"
				}
			}
			chunks = append(chunks, domain.Chunk{
				Seq: len(chunks), Content: markdown[start:end], ContextHeader: header,
				Kind: kind, Start: runeOffsets[start], End: runeOffsets[end],
				HeadingPath: strings.Split(section.header, " > "),
				Tokens:      CountTokens(header + "\n\n" + markdown[start:end]),
			})
		}
		for _, next := range units {
			length := CountTokens(markdown[section.start+next.start : section.start+next.end])
			proposal := append(append([]unit{}, current...), next)
			overLimit := len(current) > 0 && CountTokens(headerFor(proposal)+"\n\n"+
				markdown[section.start+current[0].start:section.start+next.end]) > maximum
			// Target is soft: keep an intact paragraph/list if it fits the hard limit.
			if len(current) > 0 && (currentSize >= size || overLimit) {
				flush()
				// Keep only whole units within the overlap budget and this section.
				tail, tailSize := len(current), 0
				for tail > 0 {
					item := current[tail-1]
					itemSize := CountTokens(markdown[section.start+item.start : section.start+item.end])
					tailUnits := append(append([]unit{}, current[tail-1:]...), next)
					if tailSize+itemSize > overlap || CountTokens(headerFor(tailUnits)+"\n\n"+
						markdown[section.start+item.start:section.start+next.end]) > maximum {
						break
					}
					tail--
					tailSize += itemSize
				}
				current = append([]unit(nil), current[tail:]...)
				currentSize = tailSize
			}
			current = append(current, next)
			currentSize += length
		}
		flush()
		// A short final block can join its predecessor without crossing a heading.
		if len(chunks)-sectionStart > 1 {
			last, previous := &chunks[len(chunks)-1], &chunks[len(chunks)-2]
			start := byteOffset(markdown, previous.Start)
			merged := markdown[start:byteOffset(markdown, last.End)]
			if last.Tokens < tuning.ChunkMinTokens && CountTokens(previous.ContextHeader+"\n\n"+merged) <= maximum {
				previous.Content, previous.End = merged, last.End
				previous.Tokens = CountTokens(previous.ContextHeader + "\n\n" + merged)
				if previous.Kind != last.Kind {
					previous.Kind = "text"
				}
				chunks = chunks[:len(chunks)-1]
			}
		}
	}
	attachParents(markdown, chunks, max(maximum, tuning.ParentMaxTokens))
	return chunks
}

func splitSection(source string, size int) []unit {
	var spans []span
	// Goldmark retains paragraphs and complete list/blockquote structure.
	doc := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader([]byte(source)))
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if node.Kind() != ast.KindList && node.Kind() != ast.KindBlockquote && node.Kind() != ast.KindParagraph {
			continue
		}
		start, end := nodeRange(node, source)
		if start >= 0 && end > start {
			kind := "list"
			if node.Kind() == ast.KindParagraph {
				if CountTokens(source[start:end]) > size {
					continue
				}
				kind = "text"
			}
			spans = append(spans, span{start, end, kind})
		}
	}
	for _, protected := range protectedPatterns {
		for _, match := range protected.pattern.FindAllStringIndex(source, -1) {
			spans = append(spans, span{start: match[0], end: match[1], kind: protected.kind})
		}
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].start == spans[j].start {
			return spans[i].end > spans[j].end
		}
		return spans[i].start < spans[j].start
	})
	var units []unit
	appendRange := func(start, end int, kind, header string, limit int) {
		for _, paragraph := range strings.SplitAfter(source[start:end], "\n\n") {
			for _, piece := range splitBySeparators(paragraph, []string{"\n", "。", "！", "？", ". ", " "}, limit) {
				// Plain text without separators uses rune-safe windows.
				for len(piece) > 0 {
					bytes := tokenPrefixBytes(piece, limit)
					units = append(units, unit{start: start, end: start + bytes, kind: kind, header: header})
					start += bytes
					piece = piece[bytes:]
				}
			}
		}
	}
	pos := 0
	for _, protected := range spans {
		if protected.start < pos {
			continue
		}
		if protected.start > pos {
			appendRange(pos, protected.start, "text", "", size)
		}
		header := ""
		if protected.kind == "table" {
			lines := strings.SplitN(source[protected.start:protected.end], "\n", 3)
			if len(lines) >= 2 {
				header = lines[0] + "\n" + lines[1]
			}
		}
		block := source[protected.start:protected.end]
		if CountTokens(block)+CountTokens(header) <= size ||
			(protected.kind != "code" && protected.kind != "table" && protected.kind != "list") {
			units = append(units, unit{start: protected.start, end: protected.end, kind: protected.kind, header: header})
		} else {
			// Large code and tables split only between complete lines. An oversized
			// row/formula remains intact; the worker rejects it instead of truncating.
			if protected.kind == "code" {
				header = strings.SplitN(block, "\n", 2)[0]
			}
			pieces := strings.SplitAfter(block, "\n")
			if protected.kind == "list" {
				list := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader([]byte(block))).FirstChild()
				// Oversized lists split only between complete top-level items.
				var starts []int
				if list != nil && list.Kind() == ast.KindList {
					for item := list.FirstChild(); item != nil; item = item.NextSibling() {
						start, _ := nodeRange(item, block)
						if start >= 0 {
							starts = append(starts, start)
						}
					}
				}
				pieces = nil
				if len(starts) == 0 {
					pieces = []string{block}
				} else {
					starts[0] = 0
					for i, start := range starts {
						end := len(block)
						if i+1 < len(starts) {
							end = starts[i+1]
						}
						pieces = append(pieces, block[start:end])
					}
				}
			}
			start := protected.start
			for _, line := range pieces {
				if line != "" {
					units = append(units, unit{start: start, end: start + len(line), kind: protected.kind, header: header})
					start += len(line)
				}
			}
		}
		pos = protected.end
	}
	if pos < len(source) {
		appendRange(pos, len(source), "text", "", size)
	}
	return units
}

// Adapted from WeKnora splitBySeparators: preserve separators and recurse only
// within an oversized unit, instead of flattening all paragraph boundaries.
func splitBySeparators(source string, separators []string, size int) []string {
	if source == "" || len(separators) == 0 || CountTokens(source) <= size {
		return []string{source}
	}
	for i, separator := range separators {
		if !strings.Contains(source, separator) {
			continue
		}
		var pieces []string
		for _, piece := range strings.SplitAfter(source, separator) {
			if piece != "" {
				pieces = append(pieces, splitBySeparators(piece, separators[i+1:], size)...)
			}
		}
		return pieces
	}
	return []string{source}
}
