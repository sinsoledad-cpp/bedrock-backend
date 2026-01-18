package startup

import (
	"bedrock-backend/internal/web"
	"bedrock-backend/internal/web/middleware"
	jwtware "bedrock-backend/internal/web/middleware/jwt"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func InitGinServer(hdl *web.UserHandler, jwtHdl jwtware.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	server.Static("/uploads", filepath.Join(projectRoot(), "storage", "uploads"))
	m := middleware.NewJWTAuth(jwtHdl)
	server.Use(m.Middleware())
	hdl.RegisterRoutes(server)
	return server
}
