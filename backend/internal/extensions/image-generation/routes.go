package imagegeneration

import (
	imagegenerationhandler "github.com/Wei-Shaw/sub2api/internal/extensions/image-generation/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires the authenticated user-facing image generation endpoints.
// All policy and upstream work remains in the extension handler/service.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	h *imagegenerationhandler.Handler,
	jwtAuth middleware.JWTAuthMiddleware,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	settingService *service.SettingService,
	panelRateLimiter *middleware.PanelRateLimiter,
) {
	if v1 == nil || h == nil {
		return
	}
	route := v1.Group("/image-generation")
	route.Use(gin.HandlerFunc(jwtAuth))
	route.Use(middleware.BackendModeUserGuard(settingService))
	if panelRateLimiter != nil {
		route.Use(panelRateLimiter.Global())
	}
	{
		route.GET("/options", h.Options)
		route.GET("/config", h.Config)
		route.PUT("/config", h.SaveConfig)

		generate := route.Group("")
		generate.Use(h.PrepareAPIKey())
		generate.Use(gin.HandlerFunc(apiKeyAuth))
		generate.POST("/generate", h.Generate)

		optimize := route.Group("")
		optimize.Use(h.PrepareOptimizationAPIKey())
		optimize.Use(gin.HandlerFunc(apiKeyAuth))
		optimize.POST("/optimize", h.Optimize)
	}
}
