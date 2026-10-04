package handler

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shawns-yao/shawn-blog/server/internal/app/music"
	"github.com/shawns-yao/shawn-blog/server/internal/http/middleware"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

const MusicCookieName = "music_session"
const musicCookiePath = "/api/v2/music"

type MusicHandler struct {
	service *music.Service
	secure  bool
}

func NewMusicHandler(service *music.Service, secure bool) *MusicHandler {
	return &MusicHandler{service: service, secure: secure}
}

func (h *MusicHandler) RequireAccess(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.NewBizError(response.NotLogin)
	}
	if !h.service.Allowed(claims.UserID, claims.IsAdmin) {
		return response.NewBizErrorWithMsg(response.Unauthorized, "当前账号没有音乐访问权限")
	}
	return c.Next()
}

func (h *MusicHandler) Status(c *fiber.Ctx) error {
	return response.Success(c, h.service.Status(c.UserContext(), false))
}
func (h *MusicHandler) AdminStatus(c *fiber.Ctx) error {
	return response.Success(c, h.service.Status(c.UserContext(), true))
}

func (h *MusicHandler) Session(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	parts := strings.Fields(c.Get("Authorization"))
	if !ok || claims.ExpiresAt <= time.Now().Unix() || len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.HasPrefix(parts[1], "gt_") {
		return response.NewBizErrorWithMsg(response.NotLogin, "请使用账号登录后播放音乐")
	}
	c.Cookie(&fiber.Cookie{Name: MusicCookieName, Value: parts[1], Path: musicCookiePath, HTTPOnly: true, Secure: h.secure, SameSite: "Strict", Expires: time.Unix(claims.ExpiresAt, 0)})
	c.Set("Cache-Control", "no-store")
	return response.Success(c, fiber.Map{"ready": true})
}

func (h *MusicHandler) ClearSession(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{Name: MusicCookieName, Value: "", Path: musicCookiePath, HTTPOnly: true, Secure: h.secure, SameSite: "Strict", Expires: time.Unix(1, 0), MaxAge: -1})
	return response.Success(c, fiber.Map{"ready": false})
}

func (h *MusicHandler) Catalog(c *fiber.Ctx) error {
	return h.catalog(c, false)
}

func (h *MusicHandler) PublicCatalog(c *fiber.Ctx) error { return h.catalog(c, true) }

func (h *MusicHandler) catalog(c *fiber.Ctx, public bool) error {
	c.Set("Cache-Control", "private, no-store")
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "曲库分页参数无效")
	}
	limit, err := strconv.Atoi(c.Query("limit", "30"))
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "曲库分页参数无效")
	}
	var data music.Catalog
	query := strings.TrimSpace(c.Query("query"))
	if public {
		data, err = h.service.PublicCatalog(query, offset, limit)
	} else {
		data, err = h.service.Catalog(c.UserContext(), query, offset, limit)
	}
	if err != nil {
		return musicError(err)
	}
	c.Set("Cache-Control", "private, no-store")
	return response.Success(c, data)
}

func (h *MusicHandler) Album(c *fiber.Ctx) error {
	data, err := h.service.Album(c.UserContext(), c.Params("id"))
	if err != nil {
		return musicError(err)
	}
	c.Set("Cache-Control", "private, no-store")
	return response.Success(c, data)
}

func (h *MusicHandler) PublicAlbum(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	data, err := h.service.PublicAlbum(c.Params("id"))
	if err != nil {
		return musicError(err)
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, data)
}

func (h *MusicHandler) Visibility(c *fiber.Ctx) error {
	var input struct {
		Public *bool `json:"public"`
	}
	if c.BodyParser(&input) != nil || input.Public == nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请指定公开试听状态")
	}
	data, err := h.service.SetPublic(c.UserContext(), c.Params("id"), *input.Public)
	if err != nil {
		return musicError(err)
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, data)
}

func (h *MusicHandler) Scan(c *fiber.Ctx) error {
	data, err := h.service.Scan(c.UserContext(), true)
	if err != nil {
		return musicError(err)
	}
	return response.SuccessWithMessage(c, data, "已请求扫描，请等待扫描完成后在曲库确认歌曲")
}

func (h *MusicHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请选择音乐文件")
	}
	maxSize := h.service.StatusLimit()
	if file.Size > maxSize {
		return response.NewBizErrorWithMsg(response.ParamsError, "音乐文件超过单文件大小限制")
	}
	reader, err := file.Open()
	if err != nil {
		return response.NewBizErrorWithMsg(response.ServerError, "音乐文件无法读取")
	}
	defer reader.Close()
	data, err := h.service.Upload(c.UserContext(), file.Filename, reader)
	if err != nil {
		return musicError(err)
	}
	return response.SuccessWithMessage(c, data, "文件已接收，后台正在处理")
}

func (h *MusicHandler) UploadStatus(c *fiber.Ctx) error {
	data, err := h.service.UploadStatus(c.Params("id"))
	if err != nil {
		return musicError(err)
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, data)
}

func (h *MusicHandler) UploadCompanionLyrics(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil || file.Size > music.MaxLyricsBytes {
		return response.NewBizErrorWithMsg(response.ParamsError, "请选择不超过 1 MB 的 LRC 歌词文件")
	}
	reader, err := file.Open()
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "歌词文件无法读取")
	}
	defer reader.Close()
	if err := h.service.UploadCompanionLyrics(c.UserContext(), c.Params("id"), file.Filename, reader); err != nil {
		return musicError(err)
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, fiber.Map{"uploadId": c.Params("id"), "kind": "lyrics"})
}

func (h *MusicHandler) UploadStatuses(c *fiber.Ctx) error {
	ids := strings.Split(c.Query("ids"), ",")
	if len(ids) > 100 {
		return response.NewBizErrorWithMsg(response.ParamsError, "一次最多查询100个上传任务")
	}
	items := make([]music.UploadResult, 0, len(ids))
	for _, id := range ids {
		item, err := h.service.UploadStatus(id)
		if err != nil {
			return musicError(err)
		}
		items = append(items, item)
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, items)
}

func (h *MusicHandler) Lyrics(c *fiber.Ctx) error       { return h.lyrics(c, false) }
func (h *MusicHandler) PublicLyrics(c *fiber.Ctx) error { return h.lyrics(c, true) }

func (h *MusicHandler) lyrics(c *fiber.Ctx, public bool) error {
	c.Set("Cache-Control", "private, no-store")
	var data music.LyricsResult
	var err error
	if public {
		data, err = h.service.PublicLyrics(c.UserContext(), c.Params("id"))
	} else {
		data, err = h.service.Lyrics(c.UserContext(), c.Params("id"))
	}
	if err != nil {
		return musicError(err)
	}
	return response.Success(c, data)
}

func (h *MusicHandler) UploadCover(c *fiber.Ctx) error  { return h.uploadAsset(c, "cover") }
func (h *MusicHandler) UploadLyrics(c *fiber.Ctx) error { return h.uploadAsset(c, "lyrics") }

func (h *MusicHandler) uploadAsset(c *fiber.Ctx, kind string) error {
	c.Set("Cache-Control", "no-store")
	file, err := c.FormFile("file")
	if err != nil || file.Size > music.AssetLimit(kind) {
		return response.NewBizErrorWithMsg(response.ParamsError, "请选择未超过大小限制的封面或歌词文件")
	}
	reader, err := file.Open()
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "文件无法读取")
	}
	defer reader.Close()
	data, err := h.service.UploadAsset(c.UserContext(), c.Params("id"), kind, file.Filename, reader)
	if err != nil {
		return musicError(err)
	}
	return response.Success(c, data)
}

func (h *MusicHandler) Stream(c *fiber.Ctx) error       { return h.binary(c, false, false) }
func (h *MusicHandler) Cover(c *fiber.Ctx) error        { return h.binary(c, true, false) }
func (h *MusicHandler) PublicStream(c *fiber.Ctx) error { return h.binary(c, false, true) }
func (h *MusicHandler) PublicCover(c *fiber.Ctx) error  { return h.binary(c, true, true) }

func (h *MusicHandler) binary(c *fiber.Ctx, cover, public bool) error {
	c.Set("Cache-Control", "private, no-store")
	var upstream *http.Response
	var err error
	if public {
		upstream, err = h.service.PublicBinary(c.UserContext(), c.Params("id"), c.Get("Range"), c.Get("If-Range"), cover)
	} else {
		upstream, err = h.service.Binary(c.UserContext(), c.Params("id"), c.Get("Range"), c.Get("If-Range"), cover)
	}
	if err != nil {
		return musicError(err)
	}
	for _, header := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
		if value := upstream.Header.Get(header); value != "" {
			c.Set(header, value)
		}
	}
	c.Set("Cache-Control", "private, no-store")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Status(upstream.StatusCode)
	if c.Method() == fiber.MethodHead {
		upstream.Body.Close()
		return nil
	}
	return c.SendStream(upstream.Body, int(upstream.ContentLength))
}

func musicError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return response.NewBizErrorWithMsg(response.NotFound, "音乐上传任务不存在")
	}
	if errors.Is(err, music.ErrNotPublic) {
		return response.NewBizErrorWithMsg(response.NotFound, "音乐未公开")
	}
	var input *music.InputError
	if errors.As(err, &input) {
		return response.NewBizErrorWithMsg(response.ParamsError, input.Message)
	}
	return response.NewBizErrorWithMsg(response.ServerError, err.Error())
}
