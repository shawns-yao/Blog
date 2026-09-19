package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/taxonomy"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerTaxonomyPublicRoutes(v2 fiber.Router, deps Dependencies) {
	taxHandler := newTaxonomyHandler(deps)
	tagContentHandler := newTagContentHandler(deps)

	v2.Get("/columns", taxHandler.ListColumns)
	v2.Get("/tags", taxHandler.ListTags)
	v2.Get("/tags/:id/contents", tagContentHandler.ListByTagID)
}

func registerTaxonomyAdminRoutes(v2 fiber.Router, deps Dependencies) {
	taxHandler := newTaxonomyHandler(deps)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	admin := v2.Group("/admin", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))

	admin.Post("/columns", taxHandler.CreateColumn)
	admin.Put("/columns/:id", taxHandler.UpdateColumn)
	admin.Delete("/columns/:id", taxHandler.DeleteColumn)

	admin.Post("/tags", taxHandler.CreateTag)
	admin.Put("/tags/:id", taxHandler.UpdateTag)
	admin.Delete("/tags/:id", taxHandler.DeleteTag)
}

func newTaxonomyHandler(deps Dependencies) *handler.TaxonomyHandler {
	columnRepo := persistence.NewMomentColumnRepository(deps.DB)
	tagRepo := persistence.NewTagRepository(deps.DB)

	columnSvc := taxonomy.NewColumnService(columnRepo)
	tagSvc := taxonomy.NewTagService(tagRepo)

	return handler.NewTaxonomyHandler(columnSvc, tagSvc)
}

func newTagContentHandler(deps Dependencies) *handler.TagContentHandler {
	return handler.NewTagContentHandler(
		newMomentHandler(deps),
		persistence.NewContentRepository(deps.DB),
		deps.Redis,
		deps.Config.Redis.Prefix,
	)
}
