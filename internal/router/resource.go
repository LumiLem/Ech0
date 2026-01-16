package router

import (
	"github.com/lin-snow/ech0/internal/di"
	_ "github.com/lin-snow/ech0/internal/swagger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// setupResourceRoutes 设置资源路由
func setupResourceRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Swagger UI
	appRouterGroup.ResourceGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	appRouterGroup.ResourceGroup.GET("/rss", h.CommonHandler.GetRss)
	appRouterGroup.ResourceGroup.GET("/healthz", h.CommonHandler.Healthz())
}
