package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/shawns-yao/shawn-blog/server/internal/http/handler"
	"github.com/shawns-yao/shawn-blog/server/internal/http/middleware"
	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
	"github.com/shawns-yao/shawn-blog/server/internal/infra/persistence"
)

func registerRAGRoutes(v2 fiber.Router, deps Dependencies) {
	h := handler.NewRAGHandler(deps.RAG)
	v2.Get("/public/rag/status", h.Status)
	v2.Post("/public/ask", newRateLimiter(deps, limiter.Config{
		Max: 6, Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return response.NewBizErrorWithMsg(response.TooManyRequests, "")
		},
	}), h.Ask)
	identity := persistence.NewIdentityRepository(deps.DB)
	admin := v2.Group("/admin/rag",
		middleware.RequireAuth(deps.JWTManager, identity, persistence.NewAdminTokenRepository(deps.DB)),
		middleware.RequireAdmin(identity))
	admin.Get("/index", h.IndexStats)
	admin.Post("/reindex", h.Reindex)
	admin.Post("/preview", h.Preview)
	admin.Get("/settings", h.Settings)
	admin.Put("/settings", h.UpdateSettings)
	admin.Put("/chat-priority", h.UpdateChatPriority)
	admin.Get("/documents", h.Documents)
	admin.Get("/documents/:id/chunks", h.DocumentChunks)
	admin.Post("/documents/:id/reindex", h.ReindexDocument)
	admin.Get("/metrics", h.Metrics)
}
