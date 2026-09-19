package federation

import "time"

type MentionDetected struct {
	RequestID       string
	MomentID        int64
	AuthorID        int64
	Title           string
	ShortURL        string
	MomentCreatedAt time.Time
	TargetUser      string
	TargetInstance  string
	Context         string
	MentionType     string
	At              time.Time
}

func (e MentionDetected) Name() string { return "federation.mention.detected" }
func (e MentionDetected) OccurredAt() time.Time {
	return e.At
}

type CitationDetected struct {
	RequestID       string
	MomentID        int64
	AuthorID        int64
	Title           string
	ShortURL        string
	MomentCreatedAt time.Time
	TargetInstance  string
	TargetPostID    string
	Context         string
	CitationType    string
	At              time.Time
}

func (e CitationDetected) Name() string { return "federation.citation.detected" }
func (e CitationDetected) OccurredAt() time.Time {
	return e.At
}

type DeliveryStatusChanged struct {
	DeliveryID     int64
	RequestID      string
	DeliveryType   string
	SourceMomentID *int64
	Status         string
	HTTPStatus     *int
	ErrorMessage   *string
	RemoteTicketID *string
	At             time.Time
}

func (e DeliveryStatusChanged) Name() string { return "federation.delivery.status.changed" }
func (e DeliveryStatusChanged) OccurredAt() time.Time {
	return e.At
}

type FederatedPostsCached struct {
	PostCount int
	At        time.Time
}

func (e FederatedPostsCached) Name() string          { return "federation.posts.cached" }
func (e FederatedPostsCached) OccurredAt() time.Time { return e.At }
