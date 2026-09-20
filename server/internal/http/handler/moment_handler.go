package handler

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"

	"github.com/gofiber/fiber/v2"
	domaincomment "github.com/shawns-yao/shawn-blog/server/internal/domain/comment"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
	domainfed "github.com/shawns-yao/shawn-blog/server/internal/domain/federation"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/identity"

	"github.com/shawns-yao/shawn-blog/server/internal/app/moment"
	"github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	"github.com/shawns-yao/shawn-blog/server/internal/http/contract"
	"github.com/shawns-yao/shawn-blog/server/internal/http/middleware"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

type MomentHandler struct {
	svc           *moment.Service
	contentRepo   content.Repository
	commentRepo   domaincomment.CommentRepository
	userRepo      identity.Repository
	sysCfg        *sysconfig.Service
	deliveryRepo  domainfed.OutboundDeliveryRepository
	postCacheRepo domainfed.FederatedPostCacheRepository
	instanceRepo  domainfed.FederationInstanceRepository
}

func NewMomentHandler(svc *moment.Service, contentRepo content.Repository, commentRepo domaincomment.CommentRepository, userRepo identity.Repository, sysCfg *sysconfig.Service, opts ...MomentHandlerOption) *MomentHandler {
	h := &MomentHandler{
		svc:         svc,
		contentRepo: contentRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
		sysCfg:      sysCfg,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// MomentHandlerOption 配置手记处理器的可选依赖。
type MomentHandlerOption func(*MomentHandler)

// WithMomentFederationRepos 注入联合仓库，用于联合条目内容展开。
func WithMomentFederationRepos(deliveryRepo domainfed.OutboundDeliveryRepository, postCacheRepo domainfed.FederatedPostCacheRepository, instanceRepo domainfed.FederationInstanceRepository) MomentHandlerOption {
	return func(h *MomentHandler) {
		h.deliveryRepo = deliveryRepo
		h.postCacheRepo = postCacheRepo
		h.instanceRepo = instanceRepo
	}
}

// CreateMoment godoc
// @Summary 创建手记
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.CreateMomentReq true "创建手记参数"
// @Success 200 {object} contract.MomentResp
// @Security BearerAuth
// @Router /moments [post]
// @Security JWTAuth
func (h *MomentHandler) CreateMoment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.CreateMomentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if req.Views != nil && *req.Views < 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "views 不能为负数")
	}
	extInfo, err := parseMomentExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	cmd := moment.CreateMomentCmd{
		Title:        req.Title,
		Summary:      req.Summary,
		AISummary:    req.AISummary,
		Content:      req.Content,
		Cover:        req.Cover,
		ColumnID:     req.ColumnID,
		TopicIDs:     req.TopicIDs,
		ShortURL:     req.ShortURL,
		IsPublished:  req.IsPublished,
		IsTop:        req.IsTop,
		AllowComment: req.AllowComment,
		IsOriginal:   req.IsOriginal,
		ExtInfo:      extInfo,
		CreatedAt:    req.CreatedAt,
		Views:        req.Views,
	}
	if cmd.AllowComment == nil {
		defaultAllow := true
		cmd.AllowComment = &defaultAllow
	}

	createdMoment, err := h.svc.CreateMoment(c.Context(), claims.UserID, cmd)
	if err != nil {
		if errors.Is(err, moment.ErrArticleTitleRequired) || errors.Is(err, moment.ErrContentRequired) {
			return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
		}
		if errors.Is(err, content.ErrMomentShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, content.ErrColumnNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "分区不存在")
		}
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "话题不存在")
		}
		return err
	}

	momentResponse, err := h.toMomentResp(c.Context(), createdMoment)
	if err != nil {
		return err
	}

	Audit(c, "moment.create", map[string]any{
		"momentId": createdMoment.ID,
		"title":    createdMoment.Title,
		"userId":   claims.UserID,
	})

	return response.SuccessWithMessage(c, momentResponse, "手记创建成功")
}

// UpdateMoment godoc
// @Summary 更新手记
// @Tags Moment
// @Accept json
// @Produce json
// @Param id path int true "手记ID"
// @Param request body contract.UpdateMomentReq true "更新手记参数"
// @Success 200 {object} contract.MomentResp
// @Security BearerAuth
// @Router /moments/{id} [put]
// @Security JWTAuth
func (h *MomentHandler) UpdateMoment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	var req contract.UpdateMomentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	extInfo, err := parseMomentExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	cmd := moment.UpdateMomentCmd{
		Title:        req.Title,
		Summary:      req.Summary,
		AISummary:    req.AISummary,
		Content:      req.Content,
		Cover:        req.Cover,
		ColumnID:     req.ColumnID,
		TopicIDs:     req.TopicIDs,
		ShortURL:     req.ShortURL,
		IsPublished:  req.IsPublished,
		IsTop:        req.IsTop,
		AllowComment: req.AllowComment,
		IsOriginal:   req.IsOriginal,
		ExtInfo:      extInfo,
	}
	cmd.ID = id

	updatedMoment, err := h.svc.UpdateMoment(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, moment.ErrArticleTitleRequired) || errors.Is(err, moment.ErrContentRequired) {
			return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
		}
		if errors.Is(err, content.ErrMomentShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, content.ErrColumnNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "分区不存在")
		}
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "话题不存在")
		}
		return err
	}

	momentResponse, err := h.toMomentResp(c.Context(), updatedMoment)
	if err != nil {
		return err
	}

	Audit(c, "moment.update", map[string]any{
		"momentId": updatedMoment.ID,
		"title":    updatedMoment.Title,
		"userId":   claims.UserID,
	})

	return response.SuccessWithMessage(c, momentResponse, "手记更新成功")
}

// BatchSetMomentPublished godoc
// @Summary 批量设置手记发布状态（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.BatchSetMomentPublishedReq true "批量发布状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/moments/published [put]
// @Security JWTAuth
func (h *MomentHandler) BatchSetMomentPublished(c *fiber.Ctx) error {
	var req contract.BatchSetMomentPublishedReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	for _, id := range req.IDs {
		if id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 必须为正整数")
		}
	}

	if err := h.svc.BatchSetPublished(c.Context(), moment.BatchSetPublishedCmd{
		IDs:         req.IDs,
		IsPublished: req.IsPublished,
	}); err != nil {
		return err
	}

	if req.IsPublished {
		return response.SuccessWithMessage[any](c, nil, "手记发布状态已批量更新为已发布")
	}
	return response.SuccessWithMessage[any](c, nil, "手记发布状态已批量更新为未发布")
}

// BatchSetMomentTop godoc
// @Summary 批量设置手记置顶状态（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.BatchSetMomentTopReq true "批量置顶状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/moments/top [put]
// @Security JWTAuth
func (h *MomentHandler) BatchSetMomentTop(c *fiber.Ctx) error {
	var req contract.BatchSetMomentTopReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	for _, id := range req.IDs {
		if id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 必须为正整数")
		}
	}

	if err := h.svc.BatchSetTop(c.Context(), moment.BatchSetTopCmd{
		IDs:   req.IDs,
		IsTop: req.IsTop,
	}); err != nil {
		return err
	}

	if req.IsTop {
		return response.SuccessWithMessage[any](c, nil, "手记置顶状态已批量更新为置顶")
	}
	return response.SuccessWithMessage[any](c, nil, "手记置顶状态已批量更新为取消置顶")
}

// GetMoment godoc
// @Summary 获取手记详情
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Security BearerAuth
// @Success 200 {object} contract.MomentResp
// @Router /moments/{id} [get]
func (h *MomentHandler) GetMoment(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, content.ErrMomentNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
		}
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	momentResponse, err := h.toMomentResp(c.Context(), momentItem)
	if err != nil {
		return err
	}

	return response.Success(c, momentResponse)
}

// GetMomentAdmin godoc
// @Summary 获取手记详情（管理员）
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Security BearerAuth
// @Success 200 {object} contract.MomentResp
// @Router /admin/moments/{id} [get]
// @Security JWTAuth
func (h *MomentHandler) GetMomentAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}
	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, content.ErrMomentNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
		}
		return err
	}
	momentResponse, err := h.toMomentResp(c.Context(), momentItem)
	if err != nil {
		return err
	}
	return response.Success(c, momentResponse)
}

// ListSamePeriodMoments godoc
// @Summary 获取同一时期的手记（两周内）
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Success 200 {object} contract.SamePeriodMomentListResp
// @Router /moments/{id}/same-period-moments [get]
func (h *MomentHandler) ListSamePeriodMoments(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if err != nil {
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	const windowDays = 14
	// 多取一条，用于剔除自身后最多返回 2 条。
	const fetchLimit = 3
	const maxItems = 2
	start := momentItem.CreatedAt.AddDate(0, 0, -windowDays)
	end := momentItem.CreatedAt.AddDate(0, 0, windowDays)
	siteTZ := h.sysCfg.Timezone(c.Context())

	moments, err := h.contentRepo.ListPublishedMomentsByCreatedAtRange(c.Context(), start, end, fetchLimit)
	if err != nil {
		return err
	}

	items := make([]contract.SamePeriodMomentItemResp, 0, maxItems)
	for _, item := range moments {
		if item.ID == momentItem.ID {
			continue
		}
		items = append(items, contract.SamePeriodMomentItemResp{
			ID:        item.ID,
			Title:     item.Title,
			ShortURL:  item.ShortURL,
			Summary:   item.Summary,
			Cover:     item.Cover,
			CreatedAt: item.CreatedAt.In(siteTZ),
		})
		if len(items) >= maxItems {
			break
		}
	}

	return response.Success(c, contract.SamePeriodMomentListResp{
		Items: items,
	})
}

// GetMomentByShortURL godoc
// @Summary 根据短链接获取手记
// @Tags Moment
// @Produce json
// @Param shortUrl path string true "短链接"
// @Success 200 {object} contract.MomentResp
// @Router /moments/short/{shortUrl} [get]
func (h *MomentHandler) GetMomentByShortURL(c *fiber.Ctx) error {
	shortURL := c.Params("shortUrl")
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "短链接不能为空")
	}

	momentItem, err := h.svc.GetMomentByShortURL(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, content.ErrMomentNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
		}
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	momentResponse, err := h.toMomentResp(c.Context(), momentItem)
	if err != nil {
		return err
	}
	momentResponse.Content = h.expandFederationContent(c.Context(), momentItem, momentResponse.Content)

	return response.Success(c, momentResponse)
}

// ListMoments godoc
// @Summary 获取手记列表
// @Tags Moment
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param columnId query int false "分区ID"
// @Param topicId query int false "话题ID"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.MomentListResp
// @Router /moments [get]
func (h *MomentHandler) ListMoments(c *fiber.Ctx) error {
	query := buildMomentListQuery(c)
	return h.listPublicMomentsWithQuery(c, query)
}

// ListMomentsByColumnShortURL godoc
// @Summary 根据专栏短链接获取手记列表
// @Tags Moment
// @Produce json
// @Param shortUrl path string true "专栏短链接"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} contract.MomentListResp
// @Router /columns/short/{shortUrl}/moments [get]
func (h *MomentHandler) ListMomentsByColumnShortURL(c *fiber.Ctx) error {
	shortURL := strings.TrimSpace(c.Params("shortUrl"))
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "专栏短链接不能为空")
	}

	column, err := h.contentRepo.GetColumnByShortURL(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, content.ErrColumnNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "专栏不存在")
		}
		return err
	}

	query := buildMomentListQuery(c)
	query.ColumnID = &column.ID

	return h.listPublicMomentsWithQuery(c, query)
}

// ListMomentsAdmin godoc
// @Summary 获取手记列表（管理员）
// @Tags Moment
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param columnId query int false "分区ID"
// @Param topicId query int false "话题ID"
// @Param authorId query int false "作者ID"
// @Param published query bool false "是否发布"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.MomentListResp
// @Security BearerAuth
// @Router /admin/moments [get]
func (h *MomentHandler) ListMomentsAdmin(c *fiber.Ctx) error {
	query := buildMomentListQuery(c)
	return h.listMomentsWithQuery(c, query)
}

func (h *MomentHandler) listMomentsWithQuery(c *fiber.Ctx, query contract.ListMomentsReq) error {
	if !content.ValidContentKindFilter(query.ContentKind) {
		return response.NewBizErrorWithMsg(response.ParamsError, "内容类型无效")
	}
	moments, total, err := h.svc.ListMoments(c.Context(), content.MomentListOptionsInternal(query))
	if err != nil {
		return err
	}

	momentResponses := make([]contract.MomentListItemResp, len(moments))
	for i, item := range moments {
		resp, err := h.toMomentListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		momentResponses[i] = *resp
	}

	listResponse := contract.MomentListResp{
		Items: momentResponses,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}

	return response.Success(c, listResponse)
}

func (h *MomentHandler) listPublicMomentsWithQuery(c *fiber.Ctx, query contract.ListMomentsReq) error {
	if !content.ValidContentKindFilter(query.ContentKind) {
		return response.NewBizErrorWithMsg(response.ParamsError, "内容类型无效")
	}
	moments, total, err := h.svc.ListPublicMoments(c.Context(), content.MomentListOptions{
		Page:        query.Page,
		PageSize:    query.PageSize,
		ColumnID:    query.ColumnID,
		TopicID:     query.TopicID,
		AuthorID:    query.AuthorID,
		Search:      query.Search,
		ContentKind: query.ContentKind,
	})
	if err != nil {
		return err
	}

	momentResponses := make([]contract.MomentListItemResp, len(moments))
	for i, item := range moments {
		resp, err := h.toMomentListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		momentResponses[i] = *resp
	}

	listResponse := contract.MomentListResp{
		Items: momentResponses,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}

	return response.Success(c, listResponse)
}

func buildMomentListQuery(c *fiber.Ctx) contract.ListMomentsReq {
	query := contract.ListMomentsReq{
		Page:        1,
		PageSize:    10,
		ContentKind: c.Query("contentKind"),
	}

	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil && page > 0 {
		query.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("pageSize", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		query.PageSize = pageSize
	}
	if columnID, err := strconv.ParseInt(c.Query("columnId"), 10, 64); err == nil {
		query.ColumnID = &columnID
	}
	if topicID, err := strconv.ParseInt(c.Query("topicId"), 10, 64); err == nil {
		query.TopicID = &topicID
	}
	if authorID, err := strconv.ParseInt(c.Query("authorId"), 10, 64); err == nil {
		query.AuthorID = &authorID
	}
	if publishedStr := c.Query("published"); publishedStr != "" {
		if published, err := strconv.ParseBool(publishedStr); err == nil {
			query.Published = &published
		}
	}
	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	return query
}

// ListRecentPublicMoments godoc
// @Summary 获取最近公开手记
// @Tags Public
// @Produce json
// @Success 200 {object} contract.MomentListResp
// @Router /public/moments/recent [get]
func (h *MomentHandler) ListRecentPublicMoments(c *fiber.Ctx) error {
	const page = 1
	const size = 3

	moments, total, err := h.svc.ListPublicMoments(c.Context(), content.MomentListOptions{
		Page:        page,
		PageSize:    size,
		NewestFirst: true,
		ContentKind: content.KindNote,
	})
	if err != nil {
		return err
	}

	momentResponses := make([]contract.MomentListItemResp, len(moments))
	for i, item := range moments {
		resp, err := h.toMomentListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		momentResponses[i] = *resp
	}

	return response.Success(c, contract.MomentListResp{
		Items: momentResponses,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// CheckMomentLatest godoc
// @Summary 校验手记是否最新
// @Tags Moment
// @Accept json
// @Produce json
// @Param id path int true "手记ID"
// @Param request body contract.CheckMomentLatestReq true "手记版本校验参数"
// @Success 200 {object} contract.CheckMomentLatestResp
// @Router /moments/{id}/latest [post]
func (h *MomentHandler) CheckMomentLatest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	var req contract.CheckMomentLatestReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if errors.Is(err, content.ErrMomentNotFound) {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	} else if err != nil {
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	if req.Hash == momentItem.ContentHash {
		return response.Success(c, contract.CheckMomentLatestResp{
			Latest: true,
			MomentContentPayload: contract.MomentContentPayload{
				ContentHash: momentItem.ContentHash,
			},
		})
	}

	return response.Success(c, contract.CheckMomentLatestResp{
		Latest: false,
		MomentContentPayload: contract.MomentContentPayload{
			ContentHash: momentItem.ContentHash,
			Title:       momentItem.Title,
			Summary:     momentItem.Summary,
			TOC:         mapMomentTOCNodes(momentItem.TOC),
			Content:     momentItem.Content,
		},
	})
}

// ResetMomentFederationSignals godoc
// @Summary 重置手记联合条目状态（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param id path int true "手记ID"
// @Param request body contract.ResetMomentFederationSignalsReq false "重置参数"
// @Success 200 {object} contract.ResetMomentFederationSignalsResp
// @Security BearerAuth
// @Router /admin/moments/{id}/federation/signals/reset [post]
// @Security JWTAuth
func (h *MomentHandler) ResetMomentFederationSignals(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	var req contract.ResetMomentFederationSignalsReq
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
		}
	}

	retrigger := true
	if req.Retrigger != nil {
		retrigger = *req.Retrigger
	}
	if retrigger && h.sysCfg != nil {
		settings, err := h.sysCfg.FederationSettings(c.Context())
		if err != nil {
			return response.NewBizErrorWithCause(response.ServerError, "联合配置读取失败", err)
		}
		if !settings.Enabled || !settings.AllowOutbound {
			retrigger = false
		}
	}

	updatedMoment, retriggered, err := h.svc.ResetFederationSignals(c.Context(), moment.ResetFederationSignalsCmd{
		ID:        id,
		Mentions:  req.Mentions,
		Citations: req.Citations,
		Retrigger: retrigger,
	})
	if err != nil {
		if errors.Is(err, content.ErrMomentNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
		}
		return err
	}

	Audit(c, "moment.federation.signal.reset", map[string]any{
		"momentId":    id,
		"mentions":    req.Mentions,
		"citations":   req.Citations,
		"retrigger":   retrigger,
		"retriggered": retriggered,
		"userId":      claims.UserID,
	})

	resp := contract.ResetMomentFederationSignalsResp{
		MomentID:    updatedMoment.ID,
		Retriggered: retriggered,
		ExtInfo:     jsonRawFromBytes(updatedMoment.ExtInfo),
	}
	if retriggered {
		return response.SuccessWithMessage(c, resp, "已重置联合条目并重新触发")
	}
	return response.SuccessWithMessage(c, resp, "已重置联合条目")
}

// DeleteMoment godoc
// @Summary 删除手记
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /moments/{id} [delete]
// @Security JWTAuth
func (h *MomentHandler) DeleteMoment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	if err := h.svc.DeleteMoment(c.Context(), id); err != nil {
		return err
	}

	Audit(c, "moment.delete", map[string]any{
		"momentId": id,
		"userId":   claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "手记删除成功")
}

// BatchDeleteMoments godoc
// @Summary 批量删除手记（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.BatchDeleteMomentReq true "批量删除参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/moments/batch-delete [post]
// @Security JWTAuth
func (h *MomentHandler) BatchDeleteMoments(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.BatchDeleteMomentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	for _, id := range req.IDs {
		if id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 必须为正整数")
		}
	}

	if err := h.svc.BatchDelete(c.Context(), moment.BatchDeleteCmd{IDs: req.IDs}); err != nil {
		return err
	}

	Audit(c, "moment.batch_delete", map[string]any{
		"momentIds": req.IDs,
		"userId":    claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "手记批量删除成功")
}

// GetMomentMetrics godoc
// @Summary 获取手记指标
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Success 200 {object} contract.MetricsResp
// @Router /moments/{id}/metrics [get]
func (h *MomentHandler) GetMomentMetrics(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	metrics, err := h.svc.GetMomentMetrics(c.Context(), id)
	if err != nil {
		if errors.Is(err, content.ErrMomentNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
		}
		return err
	}

	resp := contract.MetricsResp{}
	if metrics != nil {
		resp.Views = metrics.Views
		resp.Likes = metrics.Likes
		resp.Comments = metrics.Comments
	}

	return response.Success(c, resp)
}

func (h *MomentHandler) toMomentResp(ctx context.Context, momentItem *content.Moment) (*contract.MomentResp, error) {
	topics, err := h.svc.GetMomentTopics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	metrics, err := h.svc.GetMomentMetrics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	siteTZ := h.sysCfg.Timezone(ctx)
	resp := contract.MomentResp{
		ContentKind:                content.ContentKind(momentItem.ExtInfo),
		ID:                         momentItem.ID,
		Title:                      momentItem.Title,
		Summary:                    momentItem.Summary,
		AISummary:                  momentItem.AISummary,
		TOC:                        mapMomentTOCNodes(momentItem.TOC),
		Content:                    momentItem.Content,
		ContentHash:                momentItem.ContentHash,
		AuthorID:                   momentItem.AuthorID,
		Cover:                      momentItem.Cover,
		ActivityPubObjectID:        momentItem.ActivityPubObjectID,
		ActivityPubLastPublishedAt: momentItem.ActivityPubLastPublishedAt,
		ColumnID:                   momentItem.ColumnID,
		CommentID:                  momentItem.CommentID,
		ShortURL:                   momentItem.ShortURL,
		FediverseObjectURL:         h.buildFediverseObjectURL(ctx, momentItem),
		IsPublished:                momentItem.IsPublished,
		IsTop:                      momentItem.IsTop,
		IsHot:                      momentItem.IsHot,
		AllowComment:               h.allowCommentByAreaID(ctx, momentItem.CommentID),
		IsOriginal:                 momentItem.IsOriginal,
		ExtInfo:                    jsonRawFromBytes(momentItem.ExtInfo),
		ContentUpdatedAt:           momentItem.ContentUpdatedAt,
		CreatedAt:                  momentItem.CreatedAt.In(siteTZ),
		UpdatedAt:                  momentItem.UpdatedAt,
	}

	if momentItem.ColumnID != nil {
		column, colErr := h.contentRepo.GetColumnByID(ctx, *momentItem.ColumnID)
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

func (h *MomentHandler) toMomentListItemResp(ctx context.Context, momentItem *content.Moment) (*contract.MomentListItemResp, error) {
	topics, err := h.svc.GetMomentTopics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	metrics, err := h.svc.GetMomentMetrics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	siteTZ := h.sysCfg.Timezone(ctx)
	resp := contract.MomentListItemResp{
		ContentKind:      content.ContentKind(momentItem.ExtInfo),
		ID:               momentItem.ID,
		Title:            momentItem.Title,
		ShortURL:         momentItem.ShortURL,
		Summary:          momentItem.Summary,
		IsTop:            momentItem.IsTop,
		IsHot:            momentItem.IsHot,
		AllowComment:     h.allowCommentByAreaID(ctx, momentItem.CommentID),
		IsOriginal:       momentItem.IsOriginal,
		IsPublished:      momentItem.IsPublished,
		ContentUpdatedAt: momentItem.ContentUpdatedAt,
		CreatedAt:        momentItem.CreatedAt.In(siteTZ),
		UpdatedAt:        momentItem.UpdatedAt,
		Topics:           []string{},
		Cover:            momentItem.Cover,
	}
	if resp.ContentKind == content.KindNote && (resp.Cover == nil || *resp.Cover == "") {
		resp.Cover = firstExtInfoImage(momentItem.ExtInfo)
	}
	resp.CommentID = momentItem.CommentID

	if metrics != nil {
		resp.Views = metrics.Views
		resp.Likes = metrics.Likes
		resp.Comments = metrics.Comments
	}

	if len(topics) > 0 {
		topicNames := make([]string, len(topics))
		for i, topic := range topics {
			topicNames[i] = topic.Name
		}
		resp.Topics = topicNames
	}

	if momentItem.ColumnID != nil {
		column, err := h.contentRepo.GetColumnByID(ctx, *momentItem.ColumnID)
		if err != nil {
			if !errors.Is(err, content.ErrColumnNotFound) {
				return nil, err
			}
		} else if column != nil {
			resp.ColumnName = column.Name
			if column.ShortURL != nil {
				resp.ColumnShortURL = *column.ShortURL
			}
		}
	}

	if h.userRepo != nil {
		user, err := h.userRepo.FindByID(ctx, momentItem.AuthorID)
		if err != nil {
			if !errors.Is(err, identity.ErrUserNotFound) {
				return nil, err
			}
		} else if user != nil {
			resp.AuthorName = user.Nickname
			resp.Avatar = user.Avatar
		}
	}

	return &resp, nil
}

func mapMomentTOCNodes(nodes []content.TOCNode) []contract.TOCNode {
	result := make([]contract.TOCNode, len(nodes))
	for i, node := range nodes {
		result[i] = contract.TOCNode{
			Name:     node.Name,
			Anchor:   node.Anchor,
			Children: mapMomentTOCNodes(node.Children),
		}
	}
	return result
}

func (h *MomentHandler) allowCommentByAreaID(ctx context.Context, areaID *int64) bool {
	if h.commentRepo == nil || areaID == nil || *areaID <= 0 {
		return true
	}
	area, err := h.commentRepo.GetAreaByID(ctx, *areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}

// buildFediverseObjectURL 返回手记在联邦网络中的对象地址（未启用 ActivityPub 时为空）。
func (h *MomentHandler) buildFediverseObjectURL(ctx context.Context, momentItem *content.Moment) *string {
	if h.sysCfg == nil || momentItem == nil {
		return nil
	}
	settings, err := h.sysCfg.ActivityPubSettings(ctx)
	if err != nil || !settings.Enabled {
		return nil
	}
	if momentItem.ActivityPubObjectID == nil {
		return nil
	}
	objectURL := strings.TrimSpace(*momentItem.ActivityPubObjectID)
	if objectURL == "" {
		return nil
	}
	return &objectURL
}

// expandFederationContent 将手记正文中的联合信号标记替换为可渲染的富内容块。
func (h *MomentHandler) expandFederationContent(ctx context.Context, momentItem *content.Moment, contentStr string) string {
	if contentStr == "" || momentItem == nil {
		return contentStr
	}
	if h.deliveryRepo == nil || h.postCacheRepo == nil || h.instanceRepo == nil {
		return contentStr
	}
	deliveries, err := h.deliveryRepo.ListBySourceMoment(ctx, momentItem.ID, 100)
	if err != nil {
		return contentStr
	}
	// 汇总投递目标实例，用于批量取回实例与缓存帖文。
	instanceURLs := make(map[string]struct{})
	for _, d := range deliveries {
		instanceURLs[d.TargetInstanceURL] = struct{}{}
	}
	var allInstances []domainfed.FederationInstance
	var allPosts []domainfed.FederatedPostCache
	seenInstances := make(map[int64]struct{})
	for u := range instanceURLs {
		inst := h.resolveFederationInstanceByURL(ctx, u)
		if inst == nil {
			continue
		}
		if _, ok := seenInstances[inst.ID]; ok {
			continue
		}
		seenInstances[inst.ID] = struct{}{}
		allInstances = append(allInstances, *inst)
		posts, err := h.postCacheRepo.ListByInstance(ctx, inst.ID, nil, 100)
		if err == nil {
			allPosts = append(allPosts, posts...)
		}
	}
	return moment.ExpandFederationSignals(contentStr, deliveries, allPosts, allInstances)
}

// resolveFederationInstanceByURL 按实例基址查询联合实例。投递目标可能存的是裸主机名
// （来自 <cite:host|id> 标记），而实例表存的是带协议头的完整地址，因此精确匹配失败时
// 回退到补齐协议的候选地址。
func (h *MomentHandler) resolveFederationInstanceByURL(ctx context.Context, raw string) *domainfed.FederationInstance {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return nil
	}
	candidates := []string{trimmed}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		candidates = append(candidates, "https://"+trimmed, "http://"+trimmed)
	}
	for _, candidate := range candidates {
		inst, err := h.instanceRepo.GetByBaseURL(ctx, candidate)
		if err == nil && inst != nil {
			return inst
		}
	}
	return nil
}
