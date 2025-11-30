package router

import "github.com/lin-snow/ech0/internal/di"

func setupAgentRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/agent/recent", h.AgentHandler.GetRecent())

	// Auth
}
