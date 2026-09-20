package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/federation"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/identity"
	"github.com/shawns-yao/shawn-blog/server/internal/http/contract"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

type FederationPostHandler struct {
	contentRepo   content.Repository
	userRepo      identity.Repository
	postCacheRepo federation.FederatedPostCacheRepository
	cfgSvc        *sysconfig.Service
}

func NewFederationPostHandler(contentRepo content.Repository, userRepo identity.Repository, postCacheRepo federation.FederatedPostCacheRepository, cfgSvc *sysconfig.Service) *FederationPostHandler {
	return &FederationPostHandler{
		contentRepo:   contentRepo,
		userRepo:      userRepo,
		postCacheRepo: postCacheRepo,
		cfgSvc:        cfgSvc,
	}
}

// GetPostDetail returns a single post with optional related posts.
// @Summary 联合手记详情
// @Tags Federation
// @Accept json
// @Produce json
// @Param id path string true "手记 ID 或短链接"
// @Success 200 {object} contract.FederationPostDetailResp
// @Router /api/federation/posts/{id} [get]
func (h *FederationPostHandler) GetPostDetail(c *fiber.Ctx) error {
	if h.cfgSvc != nil {
		if settings, err := h.cfgSvc.FederationSettings(c.Context()); err == nil {
			if !settings.Enabled {
				return response.NewBizError(response.NotFound)
			}
		}
	}
	rawID := strings.TrimSpace(c.Params("id"))
	if rawID == "" {
		return response.NewBizError(response.ParamsError)
	}

	momentItem, err := h.resolveMoment(c, rawID)
	if err != nil {
		if errors.Is(err, content.ErrMomentNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return response.NewBizErrorWithCause(response.ServerError, "手记获取失败", err)
	}
	if !momentItem.IsPublished {
		return response.NewBizError(response.NotFound)
	}

	baseURL := resolveFederationBaseURL(c, h.cfgSvc)
	post := h.buildPostResp(c.Context(), baseURL, momentItem)

	related := h.relatedPosts(c, baseURL, momentItem)

	resp := contract.FederationPostDetailResp{
		Post:         post,
		RelatedPosts: related,
	}
	return response.Success(c, resp)
}

func (h *FederationPostHandler) resolveMoment(c *fiber.Ctx, rawID string) (*content.Moment, error) {
	if numericID, err := strconv.ParseInt(rawID, 10, 64); err == nil {
		return h.contentRepo.GetMomentByID(c.Context(), numericID)
	}
	return h.contentRepo.GetMomentByShortURL(c.Context(), rawID)
}

func (h *FederationPostHandler) buildPostResp(ctx context.Context, baseURL string, momentItem *content.Moment) contract.FederationPostResp {
	var authorName string
	var avatar *string
	if author, err := h.userRepo.FindByID(ctx, momentItem.AuthorID); err == nil && author != nil {
		authorName = author.Nickname
		if authorName == "" {
			authorName = author.Username
		}
		if author.Avatar != "" {
			avatar = &author.Avatar
		}
	}
	var preview *string
	if trimmed := strings.TrimSpace(momentItem.Summary); trimmed != "" {
		preview = &trimmed
	}
	return contract.FederationPostResp{
		ID:             momentItem.ShortURL,
		URL:            federationMomentURL(h.cfgSvc, ctx, baseURL, momentItem),
		Title:          momentItem.Title,
		Summary:        momentItem.Summary,
		ContentPreview: preview,
		Author: contract.FederationPostAuthorResp{
			Name:   authorName,
			Avatar: avatar,
		},
		PublishedAt:   momentItem.CreatedAt,
		UpdatedAt:     &momentItem.UpdatedAt,
		CoverImage:    momentItem.Cover,
		Language:      nil,
		AllowCitation: true,
		AllowComment:  true,
	}
}

// federationMomentURL 构造手记的公开地址（含发布日期段，与前台路由保持一致）。
func federationMomentURL(cfgSvc *sysconfig.Service, ctx context.Context, baseURL string, momentItem *content.Moment) string {
	tz := time.Local
	if cfgSvc != nil {
		tz = cfgSvc.Timezone(ctx)
	}
	local := momentItem.CreatedAt.In(tz)
	return fmt.Sprintf("%s/moments/%s/%s/%s/%s",
		baseURL,
		local.Format("2006"),
		local.Format("01"),
		local.Format("02"),
		momentItem.ShortURL,
	)
}

func (h *FederationPostHandler) relatedPosts(c *fiber.Ctx, baseURL string, momentItem *content.Moment) []contract.FederationPostResp {
	const limit = 6

	local := make([]contract.FederationPostResp, 0, limit)
	topics, err := h.contentRepo.GetTopicsByMomentID(c.Context(), momentItem.ID)
	if err == nil && len(topics) > 0 {
		topicID := topics[0].ID
		items, _, err := h.contentRepo.ListPublicMoments(c.Context(), content.MomentListOptions{
			Page:     1,
			PageSize: limit + 1,
			TopicID:  &topicID,
		})
		if err == nil {
			for _, item := range items {
				if item.ID == momentItem.ID {
					continue
				}
				local = append(local, h.buildPostResp(c.Context(), baseURL, item))
				if len(local) >= limit {
					break
				}
			}
		}
	}

	if len(local) >= limit {
		return local
	}

	remote, err := h.postCacheRepo.ListRecent(c.Context(), limit)
	if err != nil {
		return local
	}
	for _, item := range remote {
		if len(local) >= limit {
			break
		}
		local = append(local, mapRemotePostToResp(item))
	}
	return local
}

func mapRemotePostToResp(item federation.FederatedPostCache) contract.FederationPostResp {
	author := contract.FederationPostAuthorResp{Name: ""}
	var payload struct {
		Name   string  `json:"name"`
		URL    *string `json:"url,omitempty"`
		Avatar *string `json:"avatar,omitempty"`
	}
	if err := json.Unmarshal(item.Author, &payload); err == nil {
		author.Name = payload.Name
		author.URL = payload.URL
		author.Avatar = payload.Avatar
	}
	id := item.URL
	if item.RemotePostID != nil && *item.RemotePostID != "" {
		id = *item.RemotePostID
	}
	return contract.FederationPostResp{
		ID:             id,
		URL:            item.URL,
		Title:          item.Title,
		Summary:        item.Summary,
		ContentPreview: item.ContentPreview,
		Author:         author,
		PublishedAt:    item.PublishedAt,
		UpdatedAt:      item.UpdatedAt,
		CoverImage:     item.CoverImage,
		Language:       item.Language,
		AllowCitation:  item.AllowCitation,
		AllowComment:   item.AllowComment,
	}
}
