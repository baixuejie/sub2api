package tutorials

import "github.com/gin-gonic/gin"

// RegisterAdminRoutes registers tutorial administration endpoints on the existing
// authenticated admin route group.
func RegisterAdminRoutes(admin *gin.RouterGroup, dataDir string) {
	h := NewCoverHandler(dataDir)
	admin.POST("/tutorials/covers", h.Upload)
}

// RegisterPublicRoutes registers public tutorial cover reads.
func RegisterPublicRoutes(v1 *gin.RouterGroup, dataDir string) {
	h := NewCoverHandler(dataDir)
	v1.GET("/tutorials/covers/:filename", h.Serve)
}
