package rag

import (
	"context"
	"encoding/json"
	"strings"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

type passage struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Section   string `json:"section"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func passagesFor(evidence []domain.Evidence) []passage {
	result := make([]passage, len(evidence))
	for i, item := range evidence {
		result[i] = passage{i + 1, item.Title, item.ContextHeader, item.Content,
			item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")}
	}
	return result
}

func evidenceTokens(evidence []domain.Evidence) int {
	if len(evidence) == 0 {
		return 0
	}
	payload, _ := json.Marshal(passagesFor(evidence))
	return infrarag.CountTokens(string(payload))
}

// Keep whole recent turns. UI history remains intact when model history is trimmed.
func budgetHistory(history []domain.Message, limit int) []domain.Message {
	start := len(history)
	for start >= 2 {
		data, _ := json.Marshal(history[start-2:])
		if infrarag.CountTokens(string(data)) > limit {
			break
		}
		start -= 2
	}
	return history[start:]
}

func (s *Service) buildContext(ctx context.Context, settings settings, plan queryPlan, candidates []domain.Evidence) ([]domain.Evidence, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	var ids []int64
	seen := map[int64]bool{}
	for _, item := range candidates {
		if !seen[item.MomentID] {
			ids = append(ids, item.MomentID)
			seen[item.MomentID] = true
		}
	}
	sources, err := s.repo.ContextSources(ctx, settings.profile, ids)
	if err != nil {
		return nil, err
	}
	sourceByID := map[int64]domain.Source{}
	chunksByID := map[int64][]domain.Chunk{}
	for _, source := range sources {
		sourceByID[source.MomentID] = source
		chunksByID[source.MomentID] = infrarag.SplitMarkdownWithTuning(source.Title, source.Content, settings.tuning)
	}
	var result []domain.Evidence
	perSource := map[int64]int{}
	for _, candidate := range candidates {
		source, ok := sourceByID[candidate.MomentID]
		if !ok || source.SourceHash != candidate.SourceHash {
			return nil, domain.ErrStaleSource
		}
		maxPerSource := 2
		if plan.Strategy == "GLOBAL" {
			maxPerSource = 1
		}
		base := candidate
		base.Tokens = infrarag.CountTokens(base.ContextHeader + "\n\n" + base.Content)
		if plan.Strategy != "FACT" || base.Tokens < settings.tuning.ChunkMinTokens {
			candidate = infrarag.ExpandEvidence(candidate, source, chunksByID[candidate.MomentID], plan.Strategy != "FACT")
		} else {
			candidate = base
		}
		if candidate.Tokens > settings.tuning.ParentMaxTokens {
			candidate = base
		}
		duplicate := false
		for index, previous := range result {
			if previous.MomentID != candidate.MomentID {
				continue
			}
			if min(previous.End, candidate.End) >= max(previous.Start, candidate.Start) {
				if merged, ok := infrarag.MergeEvidence(previous, candidate, source, settings.tuning.ParentMaxTokens); ok {
					proposal := append([]domain.Evidence{}, result...)
					proposal[index] = merged
					if evidenceTokens(proposal) <= settings.tuning.ContextMaxTokens {
						result = proposal
					}
				}
				duplicate = true
				break
			}
			if strings.Join(strings.Fields(previous.Content), " ") == strings.Join(strings.Fields(candidate.Content), " ") {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		if perSource[candidate.MomentID] >= maxPerSource {
			continue
		}
		proposal := append(append([]domain.Evidence{}, result...), candidate)
		if evidenceTokens(proposal) > settings.tuning.ContextMaxTokens {
			proposal[len(proposal)-1] = base
			if evidenceTokens(proposal) > settings.tuning.ContextMaxTokens {
				continue
			}
			candidate = base
		}
		result = append(result, candidate)
		perSource[candidate.MomentID]++
		if len(result) == settings.tuning.TopK {
			break
		}
	}
	return result, nil
}
