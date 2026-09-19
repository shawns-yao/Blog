package content

import (
	"context"
	"time"
)

// Repository 定义手记（唯一长文实体）及其词表的持久化操作。
// 原 Article / ArticleCategory / ArticleTag 系列方法已随「文章并入手记」删除。
type Repository interface {
	// Moment 相关操作
	CreateMoment(ctx context.Context, moment *Moment) error
	GetMomentByID(ctx context.Context, id int64) (*Moment, error)
	GetMomentByShortURL(ctx context.Context, shortURL string) (*Moment, error)
	GetMomentByActivityPubObjectID(ctx context.Context, objectID string) (*Moment, error)
	UpdateMoment(ctx context.Context, moment *Moment) error
	DeleteMoment(ctx context.Context, id int64) error
	ListMoments(ctx context.Context, options MomentListOptionsInternal) ([]*Moment, int64, error)
	ListPublicMoments(ctx context.Context, options MomentListOptions) ([]*Moment, int64, error)
	ListPublishedMomentsByCreatedAtRange(ctx context.Context, start time.Time, end time.Time, limit int) ([]*Moment, error)
	// ListPublicMomentsForFederation 提供联合时间线的公开手记列表。
	ListPublicMomentsForFederation(ctx context.Context, since *time.Time, until *time.Time, page int, pageSize int) ([]*Moment, int64, error)
	// SyncHotMoments 根据指标同步热门手记状态
	SyncHotMoments(ctx context.Context, viewThreshold, likeThreshold, commentThreshold int64) ([]HotMomentMarked, error)

	// MomentColumn 相关操作
	CreateColumn(ctx context.Context, column *MomentColumn) error
	GetColumnByID(ctx context.Context, id int64) (*MomentColumn, error)
	GetColumnByShortURL(ctx context.Context, shortURL string) (*MomentColumn, error)
	ListColumns(ctx context.Context) ([]*MomentColumn, error)
	UpdateColumn(ctx context.Context, column *MomentColumn) error
	DeleteColumn(ctx context.Context, id int64) error

	// Tag 相关操作
	CreateTag(ctx context.Context, tag *Tag) error
	GetTagByID(ctx context.Context, id int64) (*Tag, error)
	GetTagByName(ctx context.Context, name string) (*Tag, error)
	ListTags(ctx context.Context) ([]*Tag, error)
	UpdateTag(ctx context.Context, tag *Tag) error
	DeleteTag(ctx context.Context, id int64) error
	TagIDsExist(ctx context.Context, ids []int64) (bool, error)

	// MomentTopic 关联操作
	AddTopicsToMoment(ctx context.Context, momentID int64, tagIDs []int64) error
	SyncTopicsToMoment(ctx context.Context, momentID int64, tagIDs []int64) error
	GetTopicsByMomentID(ctx context.Context, momentID int64) ([]*Tag, error)

	// Metrics 相关操作（这里用于统计交互信息，保证原子操作）
	UpdateMomentViews(ctx context.Context, momentID int64) error
	GetMomentMetrics(ctx context.Context, momentID int64) (*MomentMetrics, error)
}
