package handler

import "github.com/gin-gonic/gin"

type InboxHandlerInterface interface {
	// GetInboxList 获取收件箱消息列表
	GetInboxList() gin.HandlerFunc

	// GetUnreadInbox 获取所有未读消息
	GetUnreadInbox() gin.HandlerFunc

	// MarkInboxAsRead 将消息标记为已读
	MarkInboxAsRead() gin.HandlerFunc

	// DeleteInbox 删除指定的收件箱消息
	DeleteInbox() gin.HandlerFunc

	// ClearInbox 清空收件箱
	ClearInbox() gin.HandlerFunc
}
