// The batch embedding protocol is adapted from Tencent/WeKnora (MIT).
// See licenses/WeKnora-MIT.txt. Provider credentials remain on the server.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

var ErrEmbeddingUnavailable = errors.New("embedding service unavailable")

type Embedder struct {
	baseURL    string
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
}

func NewEmbedder(baseURL, apiKey, model string, dimensions int) *Embedder {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}
	return &Embedder{baseURL: baseURL, apiKey: apiKey, model: model, dimensions: dimensions,
		client: &http.Client{Timeout: 60 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (e *Embedder) BatchEmbed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	request := struct {
		Model          string   `json:"model"`
		Input          []string `json:"input"`
		EncodingFormat string   `json:"encoding_format"`
		Dimensions     int      `json:"dimensions,omitempty"`
	}{e.model, texts, "float", e.dimensions}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, ErrEmbeddingUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, ErrEmbeddingUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}
	response, err := e.client.Do(req)
	if err != nil {
		return nil, unavailable(ErrEmbeddingUnavailable, "network_error")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// Upstream errors can contain credentials or original text; never expose them.
		return nil, httpFailure(ErrEmbeddingUnavailable, response.StatusCode)
	}
	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16<<20)).Decode(&result); err != nil || len(result.Data) != len(texts) {
		return nil, unavailable(ErrEmbeddingUnavailable, "invalid_response")
	}
	vectors := make([][]float64, len(texts))
	dimension := e.dimensions
	for _, item := range result.Data {
		if item.Index < 0 || item.Index >= len(texts) || vectors[item.Index] != nil || len(item.Embedding) == 0 {
			return nil, ErrEmbeddingUnavailable
		}
		if dimension == 0 {
			dimension = len(item.Embedding)
		}
		if len(item.Embedding) != dimension || dimension > 16384 {
			return nil, unavailable(ErrEmbeddingUnavailable, "dimension_mismatch")
		}
		norm := 0.0
		for _, value := range item.Embedding {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, ErrEmbeddingUnavailable
			}
			norm += value * value
		}
		if norm <= 0 || math.IsInf(norm, 0) {
			return nil, ErrEmbeddingUnavailable
		}
		norm = math.Sqrt(norm)
		for i := range item.Embedding {
			item.Embedding[i] /= norm
		}
		vectors[item.Index] = item.Embedding
	}
	return vectors, nil
}
