package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/identity"
	"github.com/shawns-yao/shawn-blog/server/internal/http/contract"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

type FederationTimelineHandler struct {
	contentRepo content.Repository
	userRepo    identity.Repository
	cfgSvc      *sysconfig.Service
}

func NewFederationTimelineHandler(contentRepo content.Repository, userRepo identity.Repository, cfgSvc *sysconfig.Service) *FederationTimelineHandler {
	return &FederationTimelineHandler{contentRepo: contentRepo, userRepo: userRepo, cfgSvc: cfgSvc}
}

// ListTimelinePosts returns published moments for federation timeline.
// @Summary 联合时间线
// @Tags Federation
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param per_page query int false "每页数量"
// @Param since query string false "起始时间 RFC3339"
// @Param until query string false "结束时间 RFC3339"
// @Success 200 {object} contract.FederationTimelineResp
// @Router /api/federation/timeline/posts [get]
func (h *FederationTimelineHandler) ListTimelinePosts(c *fiber.Ctx) error {
	if h.cfgSvc != nil {
		if settings, err := h.cfgSvc.FederationSettings(c.Context()); err == nil {
			if !settings.Enabled {
				return response.NewBizError(response.NotFound)
			}
		}
	}
	page := parseIntQuery(c, "page", 1)
	if page < 1 {
		page = 1
	}
	size := parseIntQuery(c, "per_page", 20)
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	since := parseTimeQuery(c, "since")
	until := parseTimeQuery(c, "until")

	moments, total, err := h.contentRepo.ListPublicMomentsForFederation(c.Context(), since, until, page, size)
	if err != nil {
		return err
	}

	baseURL := resolveFederationBaseURL(c, h.cfgSvc)
	items := make([]contract.FederationPostResp, len(moments))
	userCache := make(map[int64]*identity.User)
	for i, momentItem := range moments {
		author, ok := userCache[momentItem.AuthorID]
		if !ok {
			user, err := h.userRepo.FindByID(c.Context(), momentItem.AuthorID)
			if err == nil {
				author = user
				userCache[momentItem.AuthorID] = user
			}
		}
		authorName := ""
		var avatar *string
		if author != nil {
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
		items[i] = contract.FederationPostResp{
			ID:             momentItem.ShortURL,
			URL:            federationMomentURL(h.cfgSvc, c.Context(), baseURL, momentItem),
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

	resp := contract.FederationTimelineResp{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}
	return response.Success(c, resp)
}

func parseIntQuery(c *fiber.Ctx, key string, fallback int) int {
	if raw := c.Query(key); raw != "" {
		if val, err := strconv.Atoi(raw); err == nil {
			return val
		}
	}
	return fallback
}

func parseTimeQuery(c *fiber.Ctx, key string) *time.Time {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &parsed
}

func resolveFederationBaseURL(c *fiber.Ctx, svc *sysconfig.Service) string {
	if svc != nil {
		if settings, err := svc.FederationSettings(c.Context()); err == nil && strings.TrimSpace(settings.InstanceURL) != "" {
			return strings.TrimRight(settings.InstanceURL, "/")
		}
	}
	scheme := "https"
	if c.Protocol() != "" {
		scheme = c.Protocol()
	}
	return scheme + "://" + c.Hostname()
}
