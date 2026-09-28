package rag

import "time"

// Tuning contains non-secret retrieval settings shared by the worker and query path.
type Tuning struct {
	ChunkSize           int     `json:"chunkSize"`
	ChunkOverlap        int     `json:"chunkOverlap"`
	IndexVersion        string  `json:"indexVersion"`
	VectorTopK          int     `json:"vectorTopK"`
	KeywordTopK         int     `json:"keywordTopK"`
	TopK                int     `json:"topK"`
	MinSimilarity       float64 `json:"minSimilarity"`
	RRFK                int     `json:"rrfK"`
	RRFVectorWeight     float64 `json:"rrfVectorWeight"`
	RRFKeywordWeight    float64 `json:"rrfKeywordWeight"`
	RerankEnabled       bool    `json:"rerankEnabled"`
	RerankCandidateTopK int     `json:"rerankCandidateTopK"`
	RerankThreshold     float64 `json:"rerankThreshold"`
	RerankFallback      bool    `json:"rerankFallback"`
}

type AdminSettings struct {
	Tuning              Tuning `json:"tuning"`
	Enabled             bool   `json:"enabled"`
	PrimaryModel        string `json:"primaryModel"`
	FallbackModel       string `json:"fallbackModel"`
	EmbeddingModel      string `json:"embeddingModel"`
	RerankModel         string `json:"rerankModel"`
	PrimaryConfigured   bool   `json:"primaryConfigured"`
	FallbackConfigured  bool   `json:"fallbackConfigured"`
	EmbeddingConfigured bool   `json:"embeddingConfigured"`
	RerankConfigured    bool   `json:"rerankConfigured"`
}

type Document struct {
	MomentID        int64      `json:"momentId"`
	Title           string     `json:"title"`
	ContentKind     string     `json:"contentKind"`
	Status          string     `json:"status"`
	Published       bool       `json:"published"`
	Chunks          int64      `json:"chunks"`
	Attempts        int        `json:"attempts"`
	LastError       string     `json:"lastError"`
	UpdatedAt       *time.Time `json:"updatedAt"`
	IndexedAt       *time.Time `json:"indexedAt"`
	IndexDurationMs *int64     `json:"indexDurationMs"`
}

type DocumentFilter struct {
	Page, PageSize              int
	Search, Status, ContentKind string
}

type DocumentPage struct {
	Items    []Document `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// QueryRun stores timings and outcomes only; questions, sessions and content are absent.
type QueryRun struct {
	Status, Reason                                   string
	DurationMs                                       int64
	EmbeddingMs, RetrievalMs, RerankMs, GenerationMs *int64
	PrimaryFailed, UsedFallback, RerankDegraded      bool
}

type QueryMetrics struct {
	Days             int            `json:"days"`
	Requests         int64          `json:"requests"`
	Answered         int64          `json:"answered"`
	NoEvidence       int64          `json:"noEvidence"`
	Unavailable      int64          `json:"unavailable"`
	PrimaryFailures  int64          `json:"primaryFailures"`
	FallbackRequests int64          `json:"fallbackRequests"`
	RerankDegraded   int64          `json:"rerankDegraded"`
	AvgDurationMs    *float64       `json:"avgDurationMs"`
	AvgEmbeddingMs   *float64       `json:"avgEmbeddingMs"`
	AvgRetrievalMs   *float64       `json:"avgRetrievalMs"`
	AvgRerankMs      *float64       `json:"avgRerankMs"`
	AvgGenerationMs  *float64       `json:"avgGenerationMs"`
	Failures         []FailureCount `json:"failures"`
}

type FailureCount struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}
