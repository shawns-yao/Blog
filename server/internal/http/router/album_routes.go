package router

import (
	"github.com/gofiber/fiber/v2"

	appalbum "github.com/shawns-yao/shawn-blog/server/internal/app/album"
	mediaapp "github.com/shawns-yao/shawn-blog/server/internal/app/media"
	"github.com/shawns-yao/shawn-blog/server/internal/http/handler"
	"github.com/shawns-yao/shawn-blog/server/internal/http/middleware"
	"github.com/shawns-yao/shawn-blog/server/internal/infra/persistence"
)

func registerAlbumPublicRoutes(v2 fiber.Router, deps Dependencies) {
	albumHandler := newAlbumHandler(deps)

	publicGroup := v2.Group("/albums")
	publicGroup.Get("/", albumHandler.ListAlbums)                        // GET /api/v2/albums
	publicGroup.Get("/:id", albumHandler.GetAlbum)                       // GET /api/v2/albums/:id
	publicGroup.Get("/short/:shortUrl", albumHandler.GetAlbumByShortURL) // GET /api/v2/albums/short/:shortUrl
	publicGroup.Get("/:id/metrics", albumHandler.GetAlbumMetrics)        // GET /api/v2/albums/:id/metrics
}

func registerAlbumAuthRoutes(v2 fiber.Router, deps Dependencies) {
	albumHandler := newAlbumHandler(deps)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)

	// Historical albums are read-only; new image content is published as notes.
	adminGroup := v2.Group("/admin", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	adminGroup.Get("/albums/:id", albumHandler.GetAlbumAdmin) // GET /api/v2/admin/albums/:id
	adminGroup.Get("/albums", albumHandler.ListAlbumsAdmin)   // GET /api/v2/admin/albums
}

func newAlbumHandler(deps Dependencies) *handler.AlbumHandler {
	albumRepo := persistence.NewAlbumRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	uploadRepo := persistence.NewUploadFileRepository(deps.DB)
	albumSvc := appalbum.NewService(albumRepo, commentRepo, deps.EventBus)
	mediaSvc := mediaapp.NewService(uploadRepo, deps.Config.Backup.UploadDir, deps.EventBus, deps.MediaGate)
	return handler.NewAlbumHandler(albumSvc, albumRepo, commentRepo, mediaSvc)
}
