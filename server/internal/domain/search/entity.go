package search

import "time"

type Kind string

const (
	KindMoment Kind = "moment"
)

type Hit struct {
	Kind      Kind
	ID        int64
	Title     string
	Summary   string
	Snippet   string
	ShortURL  *string
	CreatedAt time.Time
	Score     float64
}
