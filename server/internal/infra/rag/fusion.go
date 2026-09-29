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

// Each query/channel contributes its rank independently; raw scores never mix.
func FuseMany(vector, keyword [][]domain.Evidence, tuning domain.Tuning) []domain.Evidence {
	items := map[int64]domain.Evidence{}
	queries := max(len(vector), len(keyword))
	if queries == 0 {
		return nil
	}
	for q := 0; q < queries; q++ {
		lists := [][]domain.Evidence{nil, nil}
		if q < len(vector) {
			lists[0] = vector[q]
		}
		if q < len(keyword) {
			lists[1] = keyword[q]
		}
		for channel, list := range lists {
			weight := tuning.RRFVectorWeight
			if channel == 1 {
				weight = tuning.RRFKeywordWeight
			}
			seen := map[int64]bool{}
			for rank, item := range list {
				if seen[item.ID] {
					continue
				}
				seen[item.ID] = true
				contribution := weight / float64(queries) / float64(tuning.RRFK+rank+1)
				if existing, ok := items[item.ID]; ok {
					existing.Score += contribution
					items[item.ID] = existing
				} else {
					item.Score = contribution
					items[item.ID] = item
				}
			}
		}
	}
	result := make([]domain.Evidence, 0, len(items))
	for _, item := range items {
		if item.Score > 0 {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].ID < result[j].ID
		}
		return result[i].Score > result[j].Score
	})
	return result[:min(len(result), tuning.RerankCandidateTopK)]
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
