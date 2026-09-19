package like

import "time"

type TargetType string

const (
	TargetMoment TargetType = "moment"
	TargetAlbum  TargetType = "album"
)

type ContentLike struct {
	ID         int64
	TargetType TargetType
	TargetID   int64
	UserID     *int64
	VisitorID  *string
	ClientFP   string
	CreatedAt  time.Time
}
