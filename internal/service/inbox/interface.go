package service

import (
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
)

type InboxServiceInterface interface {
	// GetInboxList 获取收件箱消息列表
	GetInboxList(
		userid uint,
		pageQueryDto commonModel.PageQueryDto,
	) (commonModel.PageQueryResult[[]*inboxModel.Inbox], error)

	// GetUnreadInbox 获取所有未读消息
	GetUnreadInbox(userid uint) ([]*inboxModel.Inbox, error)

	// MarkAsRead 将消息标记为已读
	MarkAsRead(userid, inboxID uint) error

	// DeleteInbox 删除指定的收件箱消息
	DeleteInbox(userid, inboxID uint) error

	// ClearInbox 清空收件箱
	ClearInbox(userid uint) error
}
