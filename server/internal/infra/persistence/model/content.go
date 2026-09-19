package model

import (
	"time"

	"gorm.io/gorm"
)

type MomentColumn struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	Name      string         `gorm:"column:name;size:45;not null"`
	ShortURL  string         `gorm:"column:short_url;size:255"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (MomentColumn) TableName() string { return "moment_column" }

type Tag struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	Name      string         `gorm:"column:name;size:45;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Tag) TableName() string { return "tag" }

type MomentTopic struct {
	ID       int64 `gorm:"column:id;primaryKey"`
	MomentID int64 `gorm:"column:moment_id;not null"`
	TagID    int64 `gorm:"column:tag_id;not null"`
}

func (MomentTopic) TableName() string { return "moment_topic" }

type Moment struct {
	ID                         int64          `gorm:"column:id;primaryKey"`
	Title                      string         `gorm:"column:title;size:255;not null"`
	Summary                    string         `gorm:"column:summary;type:text;not null"`
	AISummary                  *string        `gorm:"column:ai_summary;type:text"`
	Content                    string         `gorm:"column:content;type:text;not null"`
	ContentHash                string         `gorm:"column:content_hash;size:32;not null"`
	AuthorID                   int64          `gorm:"column:author_id;not null"`
	TOC                        []byte         `gorm:"column:toc;type:jsonb;not null"`
	Cover                      *string        `gorm:"column:cover;size:255"`
	ColumnID                   *int64         `gorm:"column:column_id"`
	CommentID                  *int64         `gorm:"column:comment_id"`
	ShortURL                   string         `gorm:"column:short_url;size:255;not null"`
	ActivityPubObjectID        *string        `gorm:"column:activitypub_object_id;size:500"`
	ActivityPubLastPublishedAt *time.Time     `gorm:"column:activitypub_last_published_at"`
	IsPublished                bool           `gorm:"column:is_published"`
	IsTop                      bool           `gorm:"column:is_top"`
	IsHot                      bool           `gorm:"column:is_hot"`
	IsOriginal                 bool           `gorm:"column:is_original"`
	ExtInfo                    []byte         `gorm:"column:ext_info;type:jsonb"`
	ContentUpdatedAt           time.Time      `gorm:"column:content_updated_at"`
	CreatedAt                  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt                  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Moment) TableName() string { return "moment" }

type MomentMetrics struct {
	MomentID  int64     `gorm:"column:moment_id;primaryKey"`
	Views     int64     `gorm:"column:views"`
	Likes     int       `gorm:"column:likes"`
	Comments  int       `gorm:"column:comments"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (MomentMetrics) TableName() string { return "moment_metrics" }
