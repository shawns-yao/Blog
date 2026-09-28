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

const ChunkerVersion = "weknora-markdown-v1"

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
	if size < 200 {
		size = 1200
	}
	if overlap < 0 || overlap >= size {
		overlap = 0
	}
	source := []byte(markdown)
	doc := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(source))
	sections := []section{{header: strings.TrimSpace(title)}}
	var headings [6]string
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		heading, ok := node.(*ast.Heading)
		if !entering || !ok || heading.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		level := heading.Level
		headings[level-1] = strings.TrimSpace(string(heading.Text(source)))
		for i := level; i < len(headings); i++ {
			headings[i] = ""
		}
		path := []string{strings.TrimSpace(title)}
		for _, item := range headings {
			if item != "" {
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
		units := splitSection(markdown[section.start:end], size)
		var current []unit
		currentSize := 0
		flush := func() {
			if len(current) == 0 {
				return
			}
			start, end := current[0].start+section.start, current[len(current)-1].end+section.start
			if strings.TrimSpace(markdown[start:end]) == "" {
				return
			}
			kind, header := current[0].kind, section.header
			for _, item := range current {
				if item.kind != kind {
					kind = "text"
				}
				if item.header != "" && !strings.Contains(header, item.header) {
					header += "\n" + item.header
				}
			}
			chunks = append(chunks, domain.Chunk{
				Seq: len(chunks), Content: markdown[start:end], ContextHeader: header,
				Kind: kind, Start: runeOffsets[start], End: runeOffsets[end],
			})
		}
		for _, next := range units {
			length := utf8.RuneCountInString(markdown[section.start+next.start : section.start+next.end])
			if len(current) > 0 && currentSize+length > size {
				flush()
				// Keep only whole units within the overlap budget and this section.
				tail, tailSize := len(current), 0
				for tail > 0 {
					item := current[tail-1]
					itemSize := utf8.RuneCountInString(markdown[section.start+item.start : section.start+item.end])
					if tailSize+itemSize > overlap || tailSize+itemSize+length > size {
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
	}
	return chunks
}

func splitSection(source string, size int) []unit {
	var spans []span
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
		for _, piece := range splitBySeparators(source[start:end], []string{"\n\n", "\n", "。", "！", "？", ". ", " "}, limit) {
			// Plain text without separators uses rune-safe windows.
			for len(piece) > 0 {
				bytes, runes := 0, 0
				for bytes < len(piece) && runes < limit {
					_, width := utf8.DecodeRuneInString(piece[bytes:])
					bytes += width
					runes++
				}
				units = append(units, unit{start: start, end: start + bytes, kind: kind, header: header})
				start += bytes
				piece = piece[bytes:]
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
		if utf8.RuneCountInString(block) <= size*2 ||
			(protected.kind != "code" && protected.kind != "table") {
			units = append(units, unit{start: protected.start, end: protected.end, kind: protected.kind, header: header})
		} else {
			// Large code and tables split only between complete lines. An oversized
			// row/formula remains intact; the worker rejects it instead of truncating.
			if protected.kind == "code" {
				header = strings.SplitN(block, "\n", 2)[0]
			}
			start := protected.start
			for _, line := range strings.SplitAfter(block, "\n") {
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
	if source == "" || len(separators) == 0 || utf8.RuneCountInString(source) <= size {
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
