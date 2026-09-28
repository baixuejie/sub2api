package main

import (
	"database/sql"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pelican"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// This is the only adapter between the local gallery and upstream services.
func providePelican(router *gin.Engine, db *sql.DB, cfg *config.Config, encryptor service.SecretEncryptor,
	adminAuth middleware.AdminAuthMiddleware, jwtAuth middleware.JWTAuthMiddleware,
	audit middleware.AuditLogMiddleware, settings *service.SettingService, redis *redis.Client) *pelican.Module {
	m := pelican.NewModule(db, encryptor, cfg.Totp.EncryptionKeyConfigured, pelican.LocalEndpoint(cfg.Server.Host, cfg.Server.Port))
	limiter := middleware.NewPanelRateLimiter(redis, settings)
	admin := router.Group("/api/v1/admin", gin.HandlerFunc(adminAuth), limiter.Global(), gin.HandlerFunc(audit), middleware.AdminComplianceGuard(settings))
	user := router.Group("/api/v1", gin.HandlerFunc(jwtAuth), middleware.BackendModeUserGuard(settings), limiter.Global())
	m.RegisterRoutes(admin, user, func(c *gin.Context) int64 {
		subject, _ := middleware.GetAuthSubjectFromContext(c)
		return subject.UserID
	})
	return m
}
