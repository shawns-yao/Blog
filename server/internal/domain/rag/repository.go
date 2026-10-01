package rag

import (
	"context"
	"errors"
)

var ErrStaleSource = errors.New("rag source changed")

type Repository interface {
	Reconcile(ctx context.Context, profile string, force bool) error
	Claim(ctx context.Context, profile, leaseToken string) (*Source, error)
	CurrentSource(ctx context.Context, source Source, profile string) (bool, error)
	Complete(ctx context.Context, source Source, profile string, chunks []Chunk, durationMs int64) error
	Fail(ctx context.Context, source Source, reason string) error
	Stats(ctx context.Context, profile string) (IndexStats, error)
	Retrieve(ctx context.Context, profile string, questions []string, contentKind string, vectors [][]float64, tuning Tuning) ([][]Evidence, [][]Evidence, error)
	ContextSources(ctx context.Context, profile string, momentIDs []int64) ([]Source, error)
	Validate(ctx context.Context, profile string, evidence []Evidence) (bool, error)
	Documents(ctx context.Context, profile string, filter DocumentFilter) (DocumentPage, error)
	DocumentChunks(ctx context.Context, profile string, momentID int64, page, pageSize int) ([]Chunk, int64, error)
	ReindexDocument(ctx context.Context, profile string, momentID int64) (bool, error)
	RecordQuery(ctx context.Context, run QueryRun) error
	QueryMetrics(ctx context.Context, days int) (QueryMetrics, error)
}
