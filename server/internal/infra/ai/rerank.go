// Adapted from Tencent/WeKnora internal/models/rerank/remote_api.go (MIT).
// See licenses/WeKnora-MIT.txt.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

var ErrRerankUnavailable = errors.New("rerank unavailable")

type Reranker struct {
	baseURL, model, apiKey string
	client                 *http.Client
}

func NewReranker(baseURL, model, apiKey string, timeout time.Duration) (*Reranker, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		model == "" || apiKey == "" || timeout <= 0 || timeout > time.Minute {
		return nil, ErrRerankUnavailable
	}
	return &Reranker{baseURL: strings.TrimRight(baseURL, "/"), model: model, apiKey: apiKey,
		client: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func (r *Reranker) Rerank(ctx context.Context, query string, candidates []domain.Evidence, threshold float64) ([]domain.Evidence, error) {
	documents := make([]string, len(candidates))
	for i, candidate := range candidates {
		documents[i] = candidate.ContextHeader + "\n\n" + candidate.Content
	}
	// Do not send truncate_prompt_tokens: truncating the templated prompt can remove the query.
	payload, err := json.Marshal(struct {
		Model           string   `json:"model"`
		Query           string   `json:"query"`
		Documents       []string `json:"documents"`
		TopN            int      `json:"top_n"`
		ReturnDocuments bool     `json:"return_documents"`
	}{r.model, query, documents, len(documents), false})
	if err != nil {
		return nil, ErrRerankUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/rerank", bytes.NewReader(payload))
	if err != nil {
		return nil, ErrRerankUnavailable
	}
	request.Header.Set("Authorization", "Bearer "+r.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "shawns-blog-rag/1.0")
	response, err := r.client.Do(request)
	if err != nil {
		return nil, unavailable(ErrRerankUnavailable, "network_error")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, httpFailure(ErrRerankUnavailable, response.StatusCode)
	}
	var ranked struct {
		Results []struct {
			Index *int     `json:"index"`
			Score *float64 `json:"relevance_score"`
		} `json:"results"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&ranked) != nil || len(ranked.Results) != len(candidates) {
		return nil, unavailable(ErrRerankUnavailable, "invalid_response")
	}
	seen := make(map[int]bool, len(candidates))
	result := make([]domain.Evidence, 0, len(candidates))
	for _, row := range ranked.Results {
		if row.Index == nil || *row.Index < 0 || *row.Index >= len(candidates) || seen[*row.Index] || row.Score == nil || math.IsNaN(*row.Score) || math.IsInf(*row.Score, 0) {
			return nil, ErrRerankUnavailable
		}
		seen[*row.Index] = true
		if *row.Score >= threshold {
			item := candidates[*row.Index]
			item.Score = *row.Score
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].ID < result[j].ID
		}
		return result[i].Score > result[j].Score
	})
	return result, nil
}
