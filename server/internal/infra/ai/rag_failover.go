package ai

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"sync"
	"time"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

type ModelRoute struct {
	Name, BaseURL, Model, APIKey string
	Timeout                      time.Duration
}

type ProviderTrace struct {
	Provider string
	Attempts []string
	Failures []domain.ChannelFailure
	Fallback bool
}

type modelError struct {
	cause  error
	reason string
}

func (e *modelError) Error() string                { return e.cause.Error() + ": " + e.reason }
func (e *modelError) Unwrap() error                { return e.cause }
func unavailable(cause error, reason string) error { return &modelError{cause, reason} }

// Health belongs to the long-lived pool, not per-request settings.
// A failing route is skipped for one minute; subsequent calls retry in configured order.
type providerHealth struct {
	mu      sync.Mutex
	retryAt map[string]time.Time
}

func (h *providerHealth) cooling(name string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return time.Now().Before(h.retryAt[name])
}
func (h *providerHealth) failed(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.retryAt == nil {
		h.retryAt = make(map[string]time.Time)
	}
	h.retryAt[name] = time.Now().Add(time.Minute)
}
func routeFailure(ctx context.Context, err error) string {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "timeout"
	}
	if ctx.Err() != nil {
		return "canceled"
	}
	var failure *modelError
	if errors.As(err, &failure) {
		return failure.reason
	}
	return "provider_unavailable"
}
func validRoute(route ModelRoute) bool {
	parsed, err := url.Parse(route.BaseURL)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && route.Model != "" &&
		route.APIKey != "" && route.Name != "" && route.Timeout > 0 && route.Timeout <= time.Minute
}

type embeddingRoute struct {
	ModelRoute
	client *Embedder
}
type EmbeddingPool struct {
	routes []embeddingRoute
	budget time.Duration
	health providerHealth
}

func NewEmbeddingPool(routes []ModelRoute, dimensions int, spaceID string, budget time.Duration) (*EmbeddingPool, error) {
	if len(routes) == 0 || dimensions < 0 || dimensions > 16384 || budget <= 0 || budget > 2*time.Minute {
		return nil, ErrEmbeddingUnavailable
	}
	pool := &EmbeddingPool{budget: budget}
	seen := map[string]bool{}
	for i, route := range routes {
		if !validRoute(route) || seen[route.Name] {
			return nil, ErrEmbeddingUnavailable
		}
		// Failover may not mix embedding models or discover dimensions independently per route.
		if i > 0 && (spaceID == "" || dimensions == 0 || route.Model != routes[0].Model) {
			return nil, ErrEmbeddingUnavailable
		}
		seen[route.Name] = true
		pool.routes = append(pool.routes, embeddingRoute{route, NewEmbedder(route.BaseURL, route.APIKey, route.Model, dimensions)})
	}
	return pool, nil
}
func (p *EmbeddingPool) BatchEmbed(ctx context.Context, texts []string) ([][]float64, error) {
	vectors, _, err := p.BatchEmbedWithTrace(ctx, texts)
	return vectors, err
}

// Do not consume a source retry merely because every route is still cooling.
func (p *EmbeddingPool) Ready() bool {
	for _, route := range p.routes {
		if !p.health.cooling(route.Name) {
			return true
		}
	}
	return false
}
func (p *EmbeddingPool) BatchEmbedWithTrace(ctx context.Context, texts []string) ([][]float64, ProviderTrace, error) {
	return p.BatchEmbedChecked(ctx, texts, nil)
}

func (p *EmbeddingPool) BatchEmbedChecked(ctx context.Context, texts []string, beforeAttempt func(context.Context) error) ([][]float64, ProviderTrace, error) {
	trace := ProviderTrace{}
	if len(texts) == 0 {
		return nil, trace, nil
	}
	stage, cancel := context.WithTimeout(ctx, p.budget)
	defer cancel()
	for i, route := range p.routes {
		if stage.Err() != nil {
			break
		}
		if p.health.cooling(route.Name) {
			trace.Failures = append(trace.Failures, domain.ChannelFailure{Provider: route.Name, Reason: "cooldown"})
			continue
		}
		trace.Attempts = append(trace.Attempts, route.Name)
		attempt, cancelAttempt := context.WithTimeout(stage, route.Timeout)
		if beforeAttempt != nil {
			if err := beforeAttempt(attempt); err != nil {
				cancelAttempt()
				return nil, trace, err
			}
		}
		vectors, err := route.client.BatchEmbed(attempt, texts)
		reason := routeFailure(attempt, err)
		cancelAttempt()
		if err == nil {
			trace.Provider, trace.Fallback = route.Name, i > 0
			return vectors, trace, nil
		}
		trace.Failures = append(trace.Failures, domain.ChannelFailure{Provider: route.Name, Reason: reason})
		if stage.Err() == nil {
			p.health.failed(route.Name)
		}
	}
	return nil, trace, ErrEmbeddingUnavailable
}

type rerankRoute struct {
	ModelRoute
	client *Reranker
}
type RerankPool struct {
	routes []rerankRoute
	budget time.Duration
	health providerHealth
}

func NewRerankPool(routes []ModelRoute, budget time.Duration) (*RerankPool, error) {
	if len(routes) == 0 || budget <= 0 || budget > time.Minute {
		return nil, ErrRerankUnavailable
	}
	pool := &RerankPool{budget: budget}
	seen := map[string]bool{}
	for _, route := range routes {
		if !validRoute(route) || seen[route.Name] {
			return nil, ErrRerankUnavailable
		}
		client, err := NewReranker(route.BaseURL, route.Model, route.APIKey, route.Timeout)
		if err != nil {
			return nil, err
		}
		seen[route.Name] = true
		pool.routes = append(pool.routes, rerankRoute{route, client})
	}
	return pool, nil
}
func (p *RerankPool) Rerank(ctx context.Context, query string, candidates []domain.Evidence, threshold float64) ([]domain.Evidence, ProviderTrace, error) {
	trace := ProviderTrace{}
	stage, cancel := context.WithTimeout(ctx, p.budget)
	defer cancel()
	for i, route := range p.routes {
		if stage.Err() != nil {
			break
		}
		if p.health.cooling(route.Name) {
			trace.Failures = append(trace.Failures, domain.ChannelFailure{Provider: route.Name, Reason: "cooldown"})
			continue
		}
		trace.Attempts = append(trace.Attempts, route.Name)
		attempt, cancelAttempt := context.WithTimeout(stage, route.Timeout)
		// A threshold calibrated for the primary model cannot filter another model's scores.
		currentThreshold := threshold
		if route.Model != p.routes[0].Model {
			currentThreshold = math.Inf(-1)
		}
		ranked, err := route.client.Rerank(attempt, query, candidates, currentThreshold)
		reason := routeFailure(attempt, err)
		cancelAttempt()
		if err == nil {
			trace.Provider, trace.Fallback = route.Name, i > 0
			return ranked, trace, nil
		}
		trace.Failures = append(trace.Failures, domain.ChannelFailure{Provider: route.Name, Reason: reason})
		if stage.Err() == nil {
			p.health.failed(route.Name)
		}
	}
	return nil, trace, ErrRerankUnavailable
}

func httpFailure(cause error, status int) error {
	return unavailable(cause, fmt.Sprintf("http_%d", status))
}

func EmbeddingQueries(queries []string, instruction string) []string {
	if strings.TrimSpace(instruction) == "" {
		return queries
	}
	result := make([]string, len(queries))
	for i, query := range queries {
		result[i] = "Instruct: " + instruction + "\nQuery: " + query
	}
	return result
}
