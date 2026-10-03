package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shawns-yao/shawn-blog/server/internal/app/music"
	domainmusic "github.com/shawns-yao/shawn-blog/server/internal/domain/music"
	"github.com/shawns-yao/shawn-blog/server/internal/http/middleware"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

type MusicLibraryHandler struct{ service *music.LibraryService }

func NewMusicLibraryHandler(service *music.LibraryService) *MusicLibraryHandler {
	return &MusicLibraryHandler{service: service}
}

func musicPage(c *fiber.Ctx) (int, int, error) {
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 || offset > 100000 {
		return 0, 0, response.NewBizError(response.ParamsError)
	}
	limit, err := strconv.Atoi(c.Query("limit", "30"))
	if err != nil || limit < 1 || limit > 30 {
		return 0, 0, response.NewBizError(response.ParamsError)
	}
	return offset, limit, nil
}
func (h *MusicHandler) Browse(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	offset, limit, err := musicPage(c)
	if err != nil {
		return err
	}
	data, err := h.service.Browse(c.UserContext(), strings.TrimSpace(c.Query("query")), offset, limit)
	if err != nil {
		return musicError(err)
	}
	return response.Success(c, data)
}
func (h *MusicHandler) BrowseAlbum(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	data, err := h.service.BrowseAlbum(c.UserContext(), c.Params("id"))
	if err != nil {
		return musicError(err)
	}
	return response.Success(c, data)
}
func (h *MusicHandler) Access(c *fiber.Ctx) error {
	c.Set("Cache-Control", "private, no-store")
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.NewBizError(response.NotLogin)
	}
	return response.Success(c, fiber.Map{"privateAccess": h.service.Allowed(claims.UserID, claims.IsAdmin)})
}
func musicOwner(c *fiber.Ctx) (int64, error) {
	c.Set("Cache-Control", "private, no-store")
	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID <= 0 {
		return 0, response.NewBizError(response.NotLogin)
	}
	return claims.UserID, nil
}
func musicPlaylistID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, response.NewBizError(response.ParamsError)
	}
	return id, nil
}
func personalMusicError(err error) error {
	if errors.Is(err, domainmusic.ErrPlaylistNotFound) || errors.Is(err, music.ErrSongNotFound) {
		return response.NewBizErrorWithMsg(response.NotFound, err.Error())
	}
	if errors.Is(err, domainmusic.ErrInvalidOrder) || errors.Is(err, domainmusic.ErrLibraryLimit) {
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	var input *music.InputError
	if errors.As(err, &input) {
		return response.NewBizErrorWithMsg(response.ParamsError, input.Message)
	}
	// Do not expose SQL errors, connection details or persisted snapshots.
	return response.NewBizErrorWithCause(response.ServerError, "个人音乐数据暂时无法读取或保存，请稍后重试", err)
}
func (h *MusicLibraryHandler) Favorites(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	offset, limit, err := musicPage(c)
	if err != nil {
		return err
	}
	data, err := h.service.Favorites(c.UserContext(), user, offset, limit)
	if err != nil {
		return personalMusicError(err)
	}
	return response.Success(c, data)
}
func (h *MusicLibraryHandler) FavoriteIDs(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	ids := []string{}
	if raw := c.Query("ids"); raw != "" {
		ids = strings.Split(raw, ",")
	}
	data, err := h.service.FavoriteIDs(c.UserContext(), user, ids)
	if err != nil {
		return personalMusicError(err)
	}
	return response.Success(c, fiber.Map{"ids": data})
}
func (h *MusicLibraryHandler) SetFavorite(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	err = h.service.SetFavorite(c.UserContext(), user, c.Params("songID"), c.Method() == fiber.MethodPut)
	if err != nil {
		return personalMusicError(err)
	}
	return response.Success(c, fiber.Map{"saved": true})
}
func (h *MusicLibraryHandler) Playlists(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	if c.Method() == fiber.MethodPost {
		var input struct {
			Name string `json:"name"`
		}
		if c.BodyParser(&input) != nil {
			return response.NewBizError(response.ParamsError)
		}
		data, err := h.service.CreatePlaylist(c.UserContext(), user, input.Name)
		if err != nil {
			return personalMusicError(err)
		}
		return response.Success(c, data)
	}
	offset, limit, err := musicPage(c)
	if err != nil {
		return err
	}
	items, err := h.service.Playlists(c.UserContext(), user, offset, limit)
	if err != nil {
		return personalMusicError(err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return response.Success(c, fiber.Map{"items": items, "hasMore": hasMore})
}
func (h *MusicLibraryHandler) Playlist(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	id, err := musicPlaylistID(c)
	if err != nil {
		return err
	}
	switch c.Method() {
	case fiber.MethodPatch:
		var input struct {
			Name string `json:"name"`
		}
		if c.BodyParser(&input) != nil {
			return response.NewBizError(response.ParamsError)
		}
		err = h.service.RenamePlaylist(c.UserContext(), user, id, input.Name)
	case fiber.MethodDelete:
		err = h.service.DeletePlaylist(c.UserContext(), user, id)
	default:
		offset, limit, pageErr := musicPage(c)
		if pageErr != nil {
			return pageErr
		}
		list, page, getErr := h.service.Playlist(c.UserContext(), user, id, offset, limit)
		if getErr != nil {
			return personalMusicError(getErr)
		}
		return response.Success(c, fiber.Map{"playlist": list, "songs": page.Songs, "hasMore": page.HasMore})
	}
	if err != nil {
		return personalMusicError(err)
	}
	return response.Success(c, fiber.Map{"saved": true})
}
func (h *MusicLibraryHandler) PlaylistSong(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	id, err := musicPlaylistID(c)
	if err != nil {
		return err
	}
	err = h.service.SetPlaylistSong(c.UserContext(), user, id, c.Params("songID"), c.Method() == fiber.MethodPut)
	if err != nil {
		return personalMusicError(err)
	}
	return response.Success(c, fiber.Map{"saved": true})
}
func (h *MusicLibraryHandler) Reorder(c *fiber.Ctx) error {
	user, err := musicOwner(c)
	if err != nil {
		return err
	}
	id, err := musicPlaylistID(c)
	if err != nil {
		return err
	}
	var input struct {
		SongIDs []string `json:"songIds"`
	}
	if c.BodyParser(&input) != nil {
		return response.NewBizError(response.ParamsError)
	}
	err = h.service.ReorderPlaylist(c.UserContext(), user, id, input.SongIDs)
	if err != nil {
		return personalMusicError(err)
	}
	return response.Success(c, fiber.Map{"saved": true})
}
