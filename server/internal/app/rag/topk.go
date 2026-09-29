package rag

import (
	"math"
	"sort"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

type evidenceLimit struct {
	Minimum, Maximum, Target int
	Reason                   string
	ScoreCutoff              float64
	HasScoreCutoff           bool
}

// Select from the actual reranked candidates; model scores are not probabilities.
// A dominant relative score drop may shorten context, after strategy coverage.
func adaptiveEvidenceLimit(t domain.Tuning, plan queryPlan, candidates []domain.Evidence, reranked bool) evidenceLimit {
	if !t.DynamicTopKEnabled {
		return evidenceLimit{Minimum: t.TopK, Maximum: t.TopK, Target: t.TopK, Reason: "fixed"}
	}
	limit := evidenceLimit{Minimum: t.DynamicTopKMin, Maximum: t.DynamicTopKMax, Reason: "fact_baseline"}
	clamp := func(n int) int { return min(limit.Maximum, max(limit.Minimum, n)) }
	limit.Target = clamp(t.TopK)
	parts := max(1, len(plan.Queries)-1)
	switch plan.Strategy {
	case "FOLLOW_UP":
		limit.Minimum = clamp(3)
		limit.Target, limit.Reason = clamp(t.TopK), "follow_up"
	case "COMPARE":
		parts = max(2, parts)
		limit.Minimum = clamp(2 * parts)
		limit.Target, limit.Reason = clamp(max(t.TopK, 3*parts)), "compare_coverage"
	case "MULTI_HOP":
		parts = max(2, parts)
		limit.Minimum = clamp(2 * parts)
		limit.Target, limit.Reason = clamp(max(t.TopK, 3*parts)), "multi_hop_coverage"
	case "GLOBAL":
		sources := map[int64]bool{}
		for _, item := range candidates {
			sources[item.MomentID] = true
		}
		limit.Minimum = clamp(4)
		limit.Target, limit.Reason = clamp(max(t.TopK, len(sources))), "global_coverage"
	}
	if !reranked || len(candidates) <= limit.Minimum || plan.Strategy == "GLOBAL" {
		return limit
	}
	// Sort a copy, since multi-query anchors can subsequently change presentation order.
	ranked := append([]domain.Evidence{}, candidates...)
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	window := min(len(ranked)-1, limit.Maximum)
	var totalDrop, largestDrop float64
	boundary := 0
	for i := 0; i < window; i++ {
		if math.IsNaN(ranked[i].Score) || math.IsInf(ranked[i].Score, 0) ||
			math.IsNaN(ranked[i+1].Score) || math.IsInf(ranked[i+1].Score, 0) {
			return limit
		}
		drop := ranked[i].Score - ranked[i+1].Score
		totalDrop += drop
		if drop > largestDrop {
			largestDrop, boundary = drop, i+1
		}
	}
	// Require a clearly dominant gap rather than an arbitrary absolute score threshold.
	if boundary > 0 && largestDrop > 0 && largestDrop >= totalDrop*0.35 &&
		largestDrop >= 2*totalDrop/float64(window) {
		// A leading gap can still lower the target to the strategy minimum.
		// Minimum coverage and protected subqueries remain enforced in buildContext.
		limit.Target = clamp(boundary)
		limit.ScoreCutoff = (ranked[boundary-1].Score + ranked[boundary].Score) / 2
		limit.HasScoreCutoff = true
		limit.Reason += "+score_gap"
	}
	return limit
}
