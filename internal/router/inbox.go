package router

import "github.com/lin-snow/ech0/internal/di"

// setupInboxRoutes 配置收件箱相关路由
func setupInboxRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	appRouterGroup.AuthRouterGroup.GET("/inbox", h.InboxHandler.GetInboxList())
	appRouterGroup.AuthRouterGroup.GET("/inbox/unread", h.InboxHandler.GetUnreadInbox())
	appRouterGroup.AuthRouterGroup.PUT("/inbox/:id/read", h.InboxHandler.MarkInboxAsRead())
	appRouterGroup.AuthRouterGroup.DELETE("/inbox/:id", h.InboxHandler.DeleteInbox())
	appRouterGroup.AuthRouterGroup.DELETE("/inbox", h.InboxHandler.ClearInbox())
}
