package router

import (
	"github.com/lin-snow/ech0/internal/di"
)

func setupPwaRoutes(app *AppRouterGroup, h *di.Handlers) {
	pwa := app.AuthRouterGroup.Group("/pwa")
	{
		pwa.GET("/vapid", h.PwaHandler.GetVapidPublicKey())
		pwa.POST("/subscribe", h.PwaHandler.Subscribe())
		pwa.POST("/unsubscribe", h.PwaHandler.Unsubscribe())
		pwa.GET("/snapshot", h.PwaHandler.GetSnapshot())
		pwa.POST("/snapshot", h.PwaHandler.UpdateSnapshot())
		pwa.GET("/aggregate", h.PwaHandler.GetAggregatedStatus())
	}
}
