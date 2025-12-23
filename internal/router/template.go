package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/di"
)

// setupTemplateRoutes 设置模板路由
func setupTemplateRoutes(r *gin.Engine, h *di.Handlers) {
	r.NoRoute(h.WebHandler.Templates())
}
