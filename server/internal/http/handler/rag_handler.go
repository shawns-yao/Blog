package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apprag "github.com/shawns-yao/shawn-blog/server/internal/app/rag"
	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

type RAGHandler struct{ service *apprag.Service }

func NewRAGHandler(service *apprag.Service) *RAGHandler { return &RAGHandler{service: service} }

func (h *RAGHandler) Status(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()
	return response.Success(c, h.service.Availability(ctx))
}

func (h *RAGHandler) Ask(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	var request struct {
		Question    string           `json:"question"`
		ContentKind string           `json:"contentKind,omitempty"`
		SessionID   string           `json:"sessionId,omitempty"`
		History     []domain.Message `json:"history,omitempty"`
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if len(c.Body()) > 128000 || decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF {
		return response.Success(c, domain.Answer{Status: "invalid_scope",
			Reason: "请求格式无效，请调整问题后重试。", Citations: []domain.Citation{}})
	}
	if request.SessionID == "" {
		request.SessionID = uuid.NewString()
	} else if id, err := uuid.Parse(request.SessionID); err != nil || id == uuid.Nil {
		return response.Success(c, domain.Answer{Status: "invalid_scope",
			Reason: "会话标识无效，请重新打开问答后重试。", Citations: []domain.Citation{}})
	} else {
		request.SessionID = id.String()
	}
	ctx, cancel := context.WithTimeout(apprag.WithEvaluationTrace(c.UserContext(), c.Get("X-RAG-Evaluation") == "1"), 120*time.Second)
	defer cancel()
	return response.Success(c, h.service.Ask(ctx, request.Question, request.ContentKind, request.SessionID, request.History))
}

func (h *RAGHandler) IndexStats(c *fiber.Ctx) error {
	stats, err := h.service.IndexStats(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无法读取索引状态，请检查检索配置与数据库迁移。")
	}
	return response.Success(c, stats)
}

func (h *RAGHandler) Reindex(c *fiber.Ctx) error {
	if err := h.service.Reindex(c.UserContext()); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无法提交重建任务，请检查问答配置与索引迁移。")
	}
	return response.Success(c, struct {
		Queued bool `json:"queued"`
	}{true})
}

func (h *RAGHandler) Preview(c *fiber.Ctx) error {
	var request struct {
		Title              string `json:"title"`
		Markdown           string `json:"markdown"`
		ChunkTargetTokens  *int   `json:"chunkTargetTokens,omitempty"`
		ChunkOverlapTokens *int   `json:"chunkOverlapTokens,omitempty"`
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF || strings.TrimSpace(request.Markdown) == "" ||
		utf8.RuneCountInString(request.Markdown) > 100000 || utf8.RuneCountInString(request.Title) > 255 {
		return response.NewBizErrorWithMsg(response.ParamsError, "请提供标题与 1–100000 个字符的 Markdown。")
	}
	chunks, err := h.service.Preview(c.UserContext(), request.Title, request.Markdown, request.ChunkTargetTokens, request.ChunkOverlapTokens)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "分块配置不可用，请检查站内问答设置。")
	}
	return response.Success(c, chunks)
}

func (h *RAGHandler) Settings(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	settings, err := h.service.AdminSettings(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无法读取检索配置。")
	}
	return response.Success(c, settings)
}

func (h *RAGHandler) UpdateSettings(c *fiber.Ctx) error {
	var request domain.Tuning
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if len(c.Body()) > 8000 || decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF {
		return response.NewBizErrorWithMsg(response.ParamsError, "检索配置格式无效。")
	}
	settings, err := h.service.UpdateTuning(c.UserContext(), request)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	return response.Success(c, settings)
}

func (h *RAGHandler) UpdateChatPriority(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	var request struct {
		Priority []string `json:"priority"`
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if len(c.Body()) > 1000 || decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF {
		return response.NewBizErrorWithMsg(response.ParamsError, "语言模型优先级格式无效。")
	}
	settings, err := h.service.UpdateChatPriority(c.UserContext(), request.Priority)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	return response.Success(c, settings)
}

func (h *RAGHandler) Documents(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	result, err := h.service.Documents(c.UserContext(), domain.DocumentFilter{Page: c.QueryInt("page", 1),
		PageSize: c.QueryInt("pageSize", 20), Search: c.Query("search"), Status: c.Query("status"), ContentKind: c.Query("contentKind")})
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "文档列表读取失败，请检查筛选参数与索引迁移。")
	}
	return response.Success(c, result)
}

func (h *RAGHandler) DocumentChunks(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	page, size := c.QueryInt("page", 1), c.QueryInt("pageSize", 10)
	chunks, total, err := h.service.DocumentChunks(c.UserContext(), id, page, size)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "分块读取失败，请检查文档标识与索引状态。")
	}
	return response.Success(c, struct {
		Items    []domain.Chunk `json:"items"`
		Total    int64          `json:"total"`
		Page     int            `json:"page"`
		PageSize int            `json:"pageSize"`
	}{chunks, total, page, size})
}

func (h *RAGHandler) ReindexDocument(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := h.service.ReindexDocument(c.UserContext(), id); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无法提交重建任务，仅支持已发布的文章与手记。")
	}
	return response.Success(c, struct {
		Queued bool `json:"queued"`
	}{true})
}

func (h *RAGHandler) Metrics(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	metrics, err := h.service.QueryMetrics(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "运行指标读取失败，请检查索引迁移。")
	}
	return response.Success(c, metrics)
}
