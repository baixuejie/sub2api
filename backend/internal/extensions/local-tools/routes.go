// Package localtools is the single server integration point for local tool
// Extensions. Tool-specific routes and services remain in their own packages.
package localtools

import (
	deepseekharness "github.com/Wei-Shaw/sub2api/internal/extensions/deepseek-harness"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	coreservice "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(
	v1 *gin.RouterGroup,
	jwtAuth servermiddleware.JWTAuthMiddleware,
	auditLog servermiddleware.AuditLogMiddleware,
	apiKeyService *coreservice.APIKeyService,
	settingService *coreservice.SettingService,
	redisClient *redis.Client,
	panelRateLimiter *servermiddleware.PanelRateLimiter,
) {
	deepseekharness.RegisterRoutes(
		v1,
		jwtAuth,
		auditLog,
		apiKeyService,
		settingService,
		redisClient,
		panelRateLimiter,
	)
}
