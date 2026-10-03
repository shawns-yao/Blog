package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shawns-yao/shawn-blog/server/internal/app/music"
	"github.com/shawns-yao/shawn-blog/server/internal/http/handler"
	"github.com/shawns-yao/shawn-blog/server/internal/http/middleware"
	"github.com/shawns-yao/shawn-blog/server/internal/infra/persistence"
)

func registerMusicRoutes(v2 fiber.Router, deps Dependencies) {
	service := music.NewService(deps.Config.Music)
	h := handler.NewMusicHandler(service, deps.Config.App.Env == "production")
	identity := persistence.NewIdentityRepository(deps.DB)
	auth := middleware.RequireAuth(deps.JWTManager, identity, persistence.NewAdminTokenRepository(deps.DB))
	v2.Get("/public/music/status", h.Status)
	v2.Get("/public/music/browse", h.Browse)
	v2.Get("/public/music/browse/albums/:id", h.BrowseAlbum)
	v2.Get("/music/access", auth, h.Access)
	personal := handler.NewMusicLibraryHandler(music.NewLibraryService(persistence.NewMusicRepository(deps.DB), service))
	v2.Get("/music/favorites/check", auth, personal.FavoriteIDs)
	v2.Get("/music/favorites", auth, personal.Favorites)
	v2.Put("/music/favorites/:songID", auth, personal.SetFavorite)
	v2.Delete("/music/favorites/:songID", auth, personal.SetFavorite)
	v2.Get("/music/playlists", auth, personal.Playlists)
	v2.Post("/music/playlists", auth, personal.Playlists)
	v2.Get("/music/playlists/:id", auth, personal.Playlist)
	v2.Patch("/music/playlists/:id", auth, personal.Playlist)
	v2.Delete("/music/playlists/:id", auth, personal.Playlist)
	v2.Put("/music/playlists/:id/songs/:songID", auth, personal.PlaylistSong)
	v2.Delete("/music/playlists/:id/songs/:songID", auth, personal.PlaylistSong)
	v2.Put("/music/playlists/:id/order", auth, personal.Reorder)
	v2.Get("/public/music/catalog", h.PublicCatalog)
	v2.Get("/public/music/albums/:id", h.PublicAlbum)
	v2.Get("/public/music/lyrics/:id", h.PublicLyrics)
	v2.Get("/public/music/stream/:id", h.PublicStream)
	v2.Head("/public/music/stream/:id", h.PublicStream)
	v2.Get("/public/music/cover/:id", h.PublicCover)
	v2.Head("/public/music/cover/:id", h.PublicCover)
	v2.Delete("/music/session", h.ClearSession)
	// Cookie 只用于同源音频和封面；其他接口继续使用现有 Bearer 认证。
	cookieAuth := func(c *fiber.Ctx) error {
		if c.Get("Authorization") == "" {
			if token := c.Cookies(handler.MusicCookieName); token != "" {
				c.Request().Header.Set("Authorization", "Bearer "+token)
			}
		}
		return auth(c)
	}
	v2.Get("/music/stream/:id", cookieAuth, h.RequireAccess, h.Stream)
	v2.Head("/music/stream/:id", cookieAuth, h.RequireAccess, h.Stream)
	v2.Get("/music/cover/:id", cookieAuth, h.RequireAccess, h.Cover)
	private := v2.Group("/music", auth, h.RequireAccess)
	private.Post("/session", h.Session)
	private.Get("/catalog", h.Catalog)
	private.Get("/albums/:id", h.Album)
	private.Get("/lyrics/:id", h.Lyrics)
	admin := v2.Group("/admin/music", auth, middleware.RequireAdmin(identity))
	admin.Get("/status", h.AdminStatus)
	admin.Get("/catalog", h.Catalog)
	admin.Post("/upload", h.Upload)
	admin.Post("/scan", h.Scan)
	admin.Put("/visibility/:id", h.Visibility)
	admin.Post("/assets/:id/cover", h.UploadCover)
	admin.Post("/assets/:id/lyrics", h.UploadLyrics)
}
