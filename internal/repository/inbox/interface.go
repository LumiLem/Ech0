package repository

import (
	"context"

	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
)

type InboxRepositoryInterface interface {
	// 创建收件箱消息
	PostInbox(ctx context.Context, inbox *inboxModel.Inbox) error

	// 获取收件箱消息列表，支持分页与搜索
	GetInboxList(ctx context.Context, offset, limit int, search string) ([]*inboxModel.Inbox, int64, error)

	// 标记消息为已读
	MarkAsRead(ctx context.Context, inboxID uint) error

	// 删除收件箱消息
	DeleteInbox(ctx context.Context, inboxID uint) error

	// 清空收件箱
	ClearInbox(ctx context.Context) error

	// 获取所有未读消息
	GetUnreadInbox(ctx context.Context) ([]*inboxModel.Inbox, error)
}
