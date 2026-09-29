package rag

import "time"

type Chunk struct {
	Seq           int       `json:"seq"`
	Content       string    `json:"content"`
	ContextHeader string    `json:"contextHeader"`
	Kind          string    `json:"kind"`
	Start         int       `json:"start"`
	End           int       `json:"end"`
	Vector        []float64 `json:"vector,omitempty"`
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
	UnderstandingProvider string           `json:"understandingProvider,omitempty"`
	AnswerProvider        string           `json:"answerProvider,omitempty"`
	AnswerAttempts        []string         `json:"answerAttempts,omitempty"`
	AnswerFailures        []ChannelFailure `json:"answerFailures,omitempty"`
	Intent                QueryIntent      `json:"intent"`
	Query                 string           `json:"query"`
	UnderstandingDegraded bool             `json:"understandingDegraded"`
	EmbeddingDegraded     bool             `json:"embeddingDegraded"`
	VectorCandidates      int              `json:"vectorCandidates"`
	KeywordCandidates     int              `json:"keywordCandidates"`
	FusedCandidates       int              `json:"fusedCandidates"`
	RerankedCandidates    int              `json:"rerankedCandidates"`
	EvidenceCount         int              `json:"evidenceCount"`
}

type ChannelFailure struct {
	Provider string `json:"provider"`
	Reason   string `json:"reason"`
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
