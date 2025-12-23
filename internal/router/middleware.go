package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/middleware"
)

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine) {
	// Recovery middleware to recover from any panics and write a 500 if there was one.
	r.Use(gin.Recovery())
	// Cors middleware
	r.Use(middleware.Cors())
	// Global write guard middleware
	r.Use(middleware.WriteGuard())
}
