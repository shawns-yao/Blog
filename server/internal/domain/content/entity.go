package content

import "time"

type TOCNode struct {
	Name     string    `json:"name"`
	Anchor   string    `json:"anchor"`
	Children []TOCNode `json:"children"`
}

type MomentColumn struct {
	ID        int64
	Name      string
	ShortURL  *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Tag struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type TagPublicItem struct {
	ID          int64
	Name        string
	MomentCount int64
}

type MomentTopic struct {
	ID       int64
	MomentID int64
	TagID    int64
}

type HotMomentMarked struct {
	ID          int64
	Title       string
	ShortURL    string
	IsPublished bool
}

// Moment 是唯一的长文实体（对外称「手记」）。
// 原 Article 的 cover 已并入，lead_in 与 img 按方案删除——
// 图片与引用直接写在正文 Markdown 里。
type Moment struct {
	ID                         int64
	Title                      string
	Summary                    string
	AISummary                  *string
	Content                    string
	ContentHash                string
	AuthorID                   int64
	TOC                        []TOCNode
	Cover                      *string
	ColumnID                   *int64
	CommentID                  *int64
	ShortURL                   string
	ActivityPubObjectID        *string
	ActivityPubLastPublishedAt *time.Time
	IsPublished                bool
	IsTop                      bool
	IsHot                      bool
	IsOriginal                 bool
	ExtInfo                    []byte
	ContentUpdatedAt           time.Time
	InitialViews               int64
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	DeletedAt                  *time.Time
}

type MomentMetrics struct {
	MomentID  int64
	Views     int64
	Likes     int
	Comments  int
	UpdatedAt time.Time
}
