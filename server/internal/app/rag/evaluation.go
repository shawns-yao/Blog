package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

type evaluationKey struct{}
type evaluationRequestKey struct{}

type evaluationHit struct {
	domain.Evidence
	Score float64 `json:"score"`
}

type evaluationStage struct {
	Queries []string          `json:"queries,omitempty"`
	Lists   [][]evaluationHit `json:"lists"`
}

type evaluationRecord struct {
	SessionID   string                     `json:"sessionId"`
	Question    string                     `json:"question"`
	ContentKind string                     `json:"contentKind"`
	StartedAt   time.Time                  `json:"startedAt"`
	Profile     string                     `json:"profile"`
	Tuning      domain.Tuning              `json:"tuning"`
	PromptHash  string                     `json:"answerPromptSha256"`
	Stages      map[string]evaluationStage `json:"stages"`
	Contexts    []passage                  `json:"contexts"`
	Answer      domain.Answer              `json:"answer"`
	Run         domain.QueryRun            `json:"run"`
}

// Evaluation capture is opt-in and only writes to an operator-configured local directory.
// It neither returns extra source text to callers nor changes retrieval or generation.
func WithEvaluationTrace(ctx context.Context, enabled bool) context.Context {
	if !enabled {
		return ctx
	}
	return context.WithValue(ctx, evaluationRequestKey{}, true)
}

func (s *Service) beginEvaluation(ctx context.Context, sessionID, question, kind string, settings settings) (context.Context, *evaluationRecord) {
	if enabled, _ := ctx.Value(evaluationRequestKey{}).(bool); !enabled || s.providers.EvaluationTraceDir == "" {
		return ctx, nil
	}
	if _, err := uuid.Parse(sessionID); err != nil {
		return ctx, nil
	}
	hash := sha256.Sum256([]byte(answerPrompt))
	record := &evaluationRecord{SessionID: sessionID, Question: question, ContentKind: kind,
		StartedAt: time.Now().UTC(), Profile: settings.profile, Tuning: settings.tuning,
		PromptHash: hex.EncodeToString(hash[:]), Stages: make(map[string]evaluationStage)}
	return context.WithValue(ctx, evaluationKey{}, record), record
}

func captureEvaluationStage(ctx context.Context, name string, queries []string, lists [][]domain.Evidence) {
	record, _ := ctx.Value(evaluationKey{}).(*evaluationRecord)
	if record == nil {
		return
	}
	stage := evaluationStage{Queries: append([]string(nil), queries...), Lists: make([][]evaluationHit, len(lists))}
	for i, list := range lists {
		stage.Lists[i] = make([]evaluationHit, len(list))
		for j, item := range list {
			stage.Lists[i][j] = evaluationHit{Evidence: item, Score: item.Score}
		}
	}
	record.Stages[name] = stage
}

func (s *Service) saveEvaluation(record *evaluationRecord) {
	data, err := json.Marshal(record)
	if err == nil {
		err = os.MkdirAll(s.providers.EvaluationTraceDir, 0700)
	}
	if err == nil {
		path := filepath.Join(s.providers.EvaluationTraceDir, record.SessionID+".json")
		err = os.WriteFile(path+".partial", data, 0600)
		if err == nil {
			err = os.Rename(path+".partial", path)
		}
	}
	if err != nil {
		log.Print("[rag] evaluation capture unavailable")
	}
}
