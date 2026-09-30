package rag

import "time"

type Chunk struct {
	ChunkID       string    `json:"chunkId"`
	DocumentID    int64     `json:"documentId"`
	HeadingPath   []string  `json:"headingPath"`
	ParentID      string    `json:"parentId"`
	ParentStart   int       `json:"parentStart"`
	ParentEnd     int       `json:"parentEnd"`
	PreviousID    string    `json:"previousId,omitempty"`
	NextID        string    `json:"nextId,omitempty"`
	Tokens        int       `json:"tokens"`
	Seq           int       `json:"seq"`
	Content       string    `json:"content"`
	ContextHeader string    `json:"contextHeader"`
	Kind          string    `json:"kind"`
	Start         int       `json:"start"`
	End           int       `json:"end"`
	Vector        []float64 `json:"vector,omitempty"`
	ChunkPolicy   string    `json:"chunkPolicy,omitempty"`
	TargetTokens  int       `json:"targetTokens,omitempty"`
	OverlapTokens int       `json:"overlapTokens,omitempty"`
}

type Source struct {
	MomentID   int64
	Title      string
	Summary    string
	Content    string
	SourceHash string
	Revision   int64
	Attempts   int
	LeaseToken string
}

type Evidence struct {
	ID            int64     `json:"chunkId"`
	MomentID      int64     `json:"momentId"`
	Title         string    `json:"title"`
	ShortURL      string    `json:"-"`
	URL           string    `json:"url"`
	Content       string    `json:"content"`
	ContextHeader string    `json:"contextHeader"`
	Kind          string    `json:"kind"`
	ContentKind   string    `json:"contentKind"`
	Start         int       `json:"start"`
	End           int       `json:"end"`
	SourceHash    string    `json:"-"`
	IndexVersion  string    `json:"indexVersion"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Score         float64   `json:"-"`
	Tokens        int       `json:"tokens,omitempty"`
	Expanded      bool      `json:"expanded,omitempty"`
}

type Citation struct {
	Number int `json:"number"`
	Evidence
}

type Answer struct {
	Status       string      `json:"status"`
	Mode         string      `json:"mode,omitempty"`
	Answer       string      `json:"answer"`
	Reason       string      `json:"reason,omitempty"`
	Citations    []Citation  `json:"citations"`
	IndexVersion string      `json:"indexVersion,omitempty"`
	Trace        *QueryTrace `json:"trace,omitempty"`
}

type QueryIntent string

const (
	IntentChat           QueryIntent = "chat"
	IntentDocumentSearch QueryIntent = "document_search"
	IntentKnowledgeQuery QueryIntent = "knowledge_query"
	IntentClarify        QueryIntent = "clarify"
)

// QueryTrace explains this request only; it is not stored as conversation history.
type QueryTrace struct {
	UnderstandingSource   string           `json:"understandingSource,omitempty"`
	ProtectedTerms        []string         `json:"protectedTerms,omitempty"`
	RetrievalPolicy       *RetrievalPolicy `json:"retrievalPolicy,omitempty"`
	EvidenceSelection     string           `json:"evidenceSelection,omitempty"`
	OriginalQuery         string           `json:"originalQuery"`
	Strategy              string           `json:"strategy,omitempty"`
	Queries               []string         `json:"queries,omitempty"`
	NeedRewrite           bool             `json:"needRewrite"`
	NeedHistory           bool             `json:"needHistory"`
	NeedMultiQuery        bool             `json:"needMultiQuery"`
	RewriteDegraded       bool             `json:"rewriteDegraded"`
	ContextTokens         int              `json:"contextTokens"`
	ContextMs             int64            `json:"contextMs"`
	UnderstandingMs       int64            `json:"understandingMs"`
	HistoryTokens         int              `json:"historyTokens"`
	TokenEncoding         string           `json:"tokenEncoding"`
	UnderstandingProvider string           `json:"understandingProvider,omitempty"`
	AnswerProvider        string           `json:"answerProvider,omitempty"`
	AnswerAssessment      string           `json:"answerAssessment,omitempty"`
	AnswerFindings        []AnswerFinding  `json:"answerFindings,omitempty"`
	AnswerAttempts        []string         `json:"answerAttempts,omitempty"`
	AnswerFailures        []ChannelFailure `json:"answerFailures,omitempty"`
	Intent                QueryIntent      `json:"intent"`
	Query                 string           `json:"query"`
	UnderstandingDegraded bool             `json:"understandingDegraded"`
	EmbeddingDegraded     bool             `json:"embeddingDegraded"`
	EmbeddingProvider     string           `json:"embeddingProvider,omitempty"`
	EmbeddingAttempts     []string         `json:"embeddingAttempts,omitempty"`
	EmbeddingFailures     []ChannelFailure `json:"embeddingFailures,omitempty"`
	EmbeddingFallbackUsed bool             `json:"embeddingFallbackUsed"`
	RerankProvider        string           `json:"rerankProvider,omitempty"`
	RerankAttempts        []string         `json:"rerankAttempts,omitempty"`
	RerankFailures        []ChannelFailure `json:"rerankFailures,omitempty"`
	RerankFallbackUsed    bool             `json:"rerankFallbackUsed"`
	VectorCandidates      int              `json:"vectorCandidates"`
	KeywordCandidates     int              `json:"keywordCandidates"`
	FusedCandidates       int              `json:"fusedCandidates"`
	RerankedCandidates    int              `json:"rerankedCandidates"`
	EvidenceCount         int              `json:"evidenceCount"`
	DynamicTopK           bool             `json:"dynamicTopK"`
	TopKMinimum           int              `json:"topKMinimum,omitempty"`
	TopKMaximum           int              `json:"topKMaximum,omitempty"`
	TopKTarget            int              `json:"topKTarget,omitempty"`
	TopKReason            string           `json:"topKReason,omitempty"`
	TopKStoppedBy         string           `json:"topKStoppedBy,omitempty"`
	SubqueryAnchors       int              `json:"subqueryAnchors,omitempty"`
	EvidenceSourceIDs     []int64          `json:"evidenceSourceIds,omitempty"`
}

// Findings are model judgments; the server validates reference identities and structure.
type AnswerFinding struct {
	QuestionPart string `json:"questionPart"`
	Relation     string `json:"relation"`
	Citations    []int  `json:"citations"`
}

// Effective request policy is distinct from the configured resource ceilings.
type RetrievalPolicy struct {
	Version          string  `json:"version"`
	Mode             string  `json:"mode"`
	Reason           string  `json:"reason"`
	VectorTopK       int     `json:"vectorTopK"`
	KeywordTopK      int     `json:"keywordTopK"`
	RerankTopK       int     `json:"rerankTopK"`
	VectorWeight     float64 `json:"vectorWeight"`
	KeywordWeight    float64 `json:"keywordWeight"`
	ContextMaxTokens int     `json:"contextMaxTokens"`
}

type ChannelFailure struct {
	Provider     string `json:"provider"`
	Reason       string `json:"reason"`
	Detail       string `json:"detail,omitempty"`
	DurationMs   int64  `json:"durationMs,omitempty"`
	FinishReason string `json:"finishReason,omitempty"`
}

type HistoryPolicy struct {
	MaxRounds     int `json:"maxRounds"`
	MaxCharacters int `json:"maxCharacters"`
}

// Conversation messages supply context, never independently verified evidence.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type IndexStats struct {
	Unindexed          int64 `json:"unindexed"`
	Outdated           int64 `json:"outdated"`
	Pending            int64 `json:"pending"`
	Running            int64 `json:"running"`
	Ready              int64 `json:"ready"`
	Failed             int64 `json:"failed"`
	Excluded           int64 `json:"excluded"`
	Chunks             int64 `json:"chunks"`
	EmbeddingDimension int   `json:"embeddingDimension"`
}

type Availability struct {
	Available  bool           `json:"available"`
	Reason     string         `json:"reason"`
	IndexReady bool           `json:"indexReady"`
	History    *HistoryPolicy `json:"history,omitempty"`
}
