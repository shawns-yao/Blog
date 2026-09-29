package rag

import (
	"math"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

// Lexical diversity is an explicit heuristic, not semantic similarity or confidence.
// Recompute the best remaining candidate each step; one redundant candidate cannot
// terminate selection of a later candidate that contains different evidence.
func diversifyEvidence(candidates []domain.Evidence, anchors int, diversity float64) []domain.Evidence {
	if diversity <= 0 || len(candidates) < 2 {
		return candidates
	}
	anchors = min(anchors, len(candidates))
	sets := make([]map[string]bool, len(candidates))
	for i, candidate := range candidates {
		sets[i] = map[string]bool{}
		for _, term := range infrarag.LexicalTokens(candidate.Content) {
			sets[i][term] = true
		}
	}
	chosen := make([]int, 0, len(candidates))
	used := make([]bool, len(candidates))
	for i := 0; i < anchors; i++ {
		chosen = append(chosen, i)
		used[i] = true
	}
	for len(chosen) < len(candidates) {
		best, bestUtility := -1, math.Inf(-1)
		for i, candidate := range candidates {
			if used[i] {
				continue
			}
			redundancy := 0.0
			for _, previous := range chosen {
				redundancy = max(redundancy, lexicalOverlap(sets[i], sets[previous]))
				other := candidates[previous]
				if candidate.MomentID == other.MomentID {
					shared := max(0, min(candidate.End, other.End)-max(candidate.Start, other.Start))
					length := max(1, min(candidate.End-candidate.Start, other.End-other.Start))
					redundancy = max(redundancy, float64(shared)/float64(length))
				}
			}
			utility := (1-diversity)/float64(i+1) + diversity*(1-redundancy)
			if utility > bestUtility {
				best, bestUtility = i, utility
			}
		}
		chosen = append(chosen, best)
		used[best] = true
	}
	result := make([]domain.Evidence, len(chosen))
	for i, index := range chosen {
		result[i] = candidates[index]
	}
	return result
}

func lexicalOverlap(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	shared := 0
	for term := range a {
		if b[term] {
			shared++
		}
	}
	return float64(shared) / float64(len(a)+len(b)-shared)
}
