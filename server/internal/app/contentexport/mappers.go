package contentexport

// 本文件复刻了 handler 私有的详情映射器，使导出 meta.json 与 admin API
// 详情响应（即旧 node 导出脚本的 meta）完全一致：
//   - toMomentResp   -> internal/http/handler/moment_handler.go
// 若上述映射器发生变更，请同步本文件（mappers_test.go 的 golden 测试可防漂移）。

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/jinzhu/copier"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
)

// Mapper 持有复刻映射器所需的全部公开依赖。
type Mapper struct {
	momentSvc   *moment.Service
	contentRepo content.Repository
	commentRepo comment.CommentRepository
	userRepo    identity.Repository
	sysCfg      *sysconfig.Service
}

func NewMapper(
	momentSvc *moment.Service,
	contentRepo content.Repository,
	commentRepo comment.CommentRepository,
	userRepo identity.Repository,
	sysCfg *sysconfig.Service,
) *Mapper {
	return &Mapper{
		momentSvc:   momentSvc,
		contentRepo: contentRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
		sysCfg:      sysCfg,
	}
}

// MomentResp 复刻 MomentHandler.toMomentResp；tz 为站点时区（每个任务只取一次）。
func (m *Mapper) MomentResp(ctx context.Context, tz *time.Location, momentItem *content.Moment) (*contract.MomentResp, error) {
	topics, err := m.momentSvc.GetMomentTopics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}
	metrics, err := m.momentSvc.GetMomentMetrics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	resp := contract.MomentResp{
		ID:                         momentItem.ID,
		Title:                      momentItem.Title,
		Summary:                    momentItem.Summary,
		AISummary:                  momentItem.AISummary,
		TOC:                        mapTOCNodes(momentItem.TOC),
		Content:                    momentItem.Content,
		ContentHash:                momentItem.ContentHash,
		AuthorID:                   momentItem.AuthorID,
		Cover:                      momentItem.Cover,
		ActivityPubObjectID:        momentItem.ActivityPubObjectID,
		ActivityPubLastPublishedAt: momentItem.ActivityPubLastPublishedAt,
		ColumnID:                   momentItem.ColumnID,
		CommentID:                  momentItem.CommentID,
		ShortURL:                   momentItem.ShortURL,
		IsPublished:                momentItem.IsPublished,
		IsTop:                      momentItem.IsTop,
		IsHot:                      momentItem.IsHot,
		AllowComment:               m.allowCommentPtr(ctx, momentItem.CommentID),
		IsOriginal:                 momentItem.IsOriginal,
		ExtInfo:                    jsonRawFromBytes(momentItem.ExtInfo),
		ContentUpdatedAt:           momentItem.ContentUpdatedAt,
		CreatedAt:                  momentItem.CreatedAt.In(tz),
		UpdatedAt:                  momentItem.UpdatedAt,
	}

	if momentItem.ColumnID != nil {
		column, colErr := m.contentRepo.GetColumnByID(ctx, *momentItem.ColumnID)
		if colErr == nil && column != nil {
			resp.ColumnName = column.Name
			if column.ShortURL != nil {
				resp.ColumnShortURL = *column.ShortURL
			}
		}
	}

	if len(topics) > 0 {
		resp.Topics = make([]contract.TagResp, len(topics))
		for i, topic := range topics {
			if err := copier.Copy(&resp.Topics[i], topic); err != nil {
				return nil, err
			}
		}
	}

	if metrics != nil {
		var metricsResp contract.MetricsResp
		if err := copier.Copy(&metricsResp, metrics); err != nil {
			return nil, err
		}
		resp.Metrics = &metricsResp
	}

	return &resp, nil
}

func (m *Mapper) allowCommentPtr(ctx context.Context, areaID *int64) bool {
	if m.commentRepo == nil || areaID == nil || *areaID <= 0 {
		return true
	}
	area, err := m.commentRepo.GetAreaByID(ctx, *areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}

func (m *Mapper) allowCommentID(ctx context.Context, areaID int64) bool {
	if m.commentRepo == nil || areaID <= 0 {
		return true
	}
	area, err := m.commentRepo.GetAreaByID(ctx, areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}

func mapTOCNodes(nodes []content.TOCNode) []contract.TOCNode {
	result := make([]contract.TOCNode, len(nodes))
	for i, node := range nodes {
		result[i] = contract.TOCNode{
			Name:     node.Name,
			Anchor:   node.Anchor,
			Children: mapTOCNodes(node.Children),
		}
	}
	return result
}

func jsonRawFromBytes(value []byte) *contract.JSONRaw {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	copied := append([]byte(nil), trimmed...)
	raw := contract.JSONRaw(copied)
	return &raw
}

// MetaNode 是写入 meta.json / flatten 文档头部的包装结构，
// 与旧 node 导出脚本的字段保持一致。
type MetaNode struct {
	Kind       string    `json:"kind"`
	ID         int64     `json:"id"`
	RoutePath  string    `json:"routePath"`
	SourcePath string    `json:"sourcePath"`
	ExportedAt time.Time `json:"exportedAt"`
	Metadata   any       `json:"metadata"`
}

// 以下 *Meta 类型利用"浅层同名字段遮蔽"把 content 从 JSON 中剔除：
// 外层 Content 声明为 json:"content,omitempty" 的 nil 指针，占据 content 这个
// 键名（深度 0 优先于内嵌结构体深度 1 的同名字段），nil + omitempty 即被省略，
// 其余字段保持原结构体顺序输出。（注意 `json:"-"` 不占键名，无法遮蔽深层字段。）

type momentMeta struct {
	*contract.MomentResp
	Content *string `json:"content,omitempty"`
}

func marshalIndent(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}
