package deepseekharness

import (
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	coreservice "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterRoutes wires the authenticated browser flow and the ticket-authenticated helper flow.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	jwtAuth servermiddleware.JWTAuthMiddleware,
	auditLog servermiddleware.AuditLogMiddleware,
	apiKeyService *coreservice.APIKeyService,
	settingService *coreservice.SettingService,
	redisClient *redis.Client,
	panelRateLimiter *servermiddleware.PanelRateLimiter,
) {
	if v1 == nil || jwtAuth == nil || apiKeyService == nil || settingService == nil || redisClient == nil || panelRateLimiter == nil {
		return
	}
	store := newRedisInstallStore(redisClient)
	service := newInstallService(apiKeyService, settingService, store)
	handler := newInstallHandler(service)

	route := v1.Group("/deepseek-harness")

	authenticated := route.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(servermiddleware.BackendModeUserGuard(settingService))
	if panelRateLimiter != nil {
		authenticated.Use(panelRateLimiter.Global())
	}
	if auditLog != nil {
		authenticated.Use(gin.HandlerFunc(auditLog))
	}
	{
		authenticated.GET("/profile", handler.Profile)
		authenticated.POST("/sessions", handler.CreateSession)
		authenticated.GET("/sessions/:id", handler.GetSession)
	}

	publicHandlers := make([]gin.HandlerFunc, 0, 1)
	if panelRateLimiter != nil {
		publicHandlers = append(publicHandlers, panelRateLimiter.PublicIP())
	}
	route.POST("/exchange", append(publicHandlers, handler.Exchange)...)
	route.POST("/sessions/:id/events", append(publicHandlers, handler.UpdateSession)...)
}
