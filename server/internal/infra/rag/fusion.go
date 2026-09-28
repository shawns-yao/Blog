// Adapted from Tencent/WeKnora knowledgebase_search_fusion.go (MIT).
// See licenses/WeKnora-MIT.txt.
package rag

import (
	"sort"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

// Fuse uses one-based ranks rather than mixing incompatible raw scores.
func Fuse(vector, keyword []domain.Evidence, tuning domain.Tuning) []domain.Evidence {
	vectorRanks, keywordRanks := make(map[int64]int), make(map[int64]int)
	items := make(map[int64]domain.Evidence)
	for i, item := range vector {
		if _, exists := vectorRanks[item.ID]; !exists {
			vectorRanks[item.ID] = i + 1
			items[item.ID] = item
		}
	}
	for i, item := range keyword {
		if _, exists := keywordRanks[item.ID]; !exists {
			keywordRanks[item.ID] = i + 1
		}
		if _, exists := items[item.ID]; !exists {
			items[item.ID] = item
		}
	}
	merged := make([]domain.Evidence, 0, len(items))
	for id, item := range items {
		item.Score = 0
		if rank, exists := vectorRanks[id]; exists {
			item.Score += tuning.RRFVectorWeight / float64(tuning.RRFK+rank)
		}
		if rank, exists := keywordRanks[id]; exists {
			item.Score += tuning.RRFKeywordWeight / float64(tuning.RRFK+rank)
		}
		if item.Score > 0 {
			merged = append(merged, item)
		}
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Score == merged[j].Score {
			return merged[i].ID < merged[j].ID
		}
		return merged[i].Score > merged[j].Score
	})
	return merged[:min(len(merged), tuning.RerankCandidateTopK)]
}

// SelectEvidence applies source diversity after reranking so the model sees
// the highest-ranked non-overlapping passages, rather than a pre-trimmed pool.
func SelectEvidence(candidates []domain.Evidence, limit int) []domain.Evidence {
	result := make([]domain.Evidence, 0, limit)
	perSource := make(map[int64]int)
	for _, candidate := range candidates {
		if perSource[candidate.MomentID] >= 2 {
			continue
		}
		duplicate := false
		for _, selected := range result {
			if selected.MomentID != candidate.MomentID {
				continue
			}
			overlap := min(selected.End, candidate.End) - max(selected.Start, candidate.Start)
			if overlap > 0 && overlap*2 >= min(selected.End-selected.Start, candidate.End-candidate.Start) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result = append(result, candidate)
			perSource[candidate.MomentID]++
			if len(result) == limit {
				break
			}
		}
	}
	return result
}
