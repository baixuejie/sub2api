package tutorials

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers tutorial administration endpoints on the existing
// authenticated admin route group.
func RegisterAdminRoutes(admin *gin.RouterGroup, dataDir string) {
	h := NewCoverHandler(dataDir)
	admin.POST("/tutorials/covers", h.Upload)
}

// RegisterPublicRoutes registers public tutorial reads and play-count actions.
// The video reader callback keeps the editable settings owned by the core
// settings service while the counters remain in this extension's table.
func RegisterPublicRoutes(v1 *gin.RouterGroup, dataDir string, db *sql.DB, readVideos PublicVideosProvider) {
	h := NewCoverHandler(dataDir)
	v1.GET("/tutorials/covers/:filename", h.Serve)
	if db == nil || readVideos == nil {
		return
	}
	public := NewVideoHandler(readVideos, NewSQLPlayCountStore(db))
	v1.GET("/tutorials/videos", public.List)
	// Wildcard capture preserves video IDs that legitimately contain '/'.
	v1.POST("/tutorials/videos/*id", public.Play)
}
