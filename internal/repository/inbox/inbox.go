package repository

import (
	"context"

	"gorm.io/gorm"

	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	"github.com/lin-snow/ech0/internal/transaction"
)

type InboxRepository struct {
	db func() *gorm.DB
}

func NewInboxRepository(dbProvider func() *gorm.DB) InboxRepositoryInterface {
	return &InboxRepository{
		db: dbProvider,
	}
}

// getDB 从上下文中获取事务
func (inboxRepository *InboxRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return inboxRepository.db()
}

// PostInbox 创建收件箱消息
func (inboxRepository *InboxRepository) PostInbox(ctx context.Context, inbox *inboxModel.Inbox) error {
	return inboxRepository.getDB(ctx).Create(inbox).Error
}

// GetInboxList 获取收件箱消息列表，支持分页、搜索和倒序
func (inboxRepository *InboxRepository) GetInboxList(
	ctx context.Context,
	offset, limit int,
	search string,
) ([]*inboxModel.Inbox, int64, error) {
	var (
		inboxes []*inboxModel.Inbox
		total   int64
	)

	query := inboxRepository.getDB(ctx).
		Model(&inboxModel.Inbox{})

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where(
			"content LIKE ? OR source LIKE ? OR type LIKE ?",
			searchLike,
			searchLike,
			searchLike,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&inboxes).Error; err != nil {
		return nil, 0, err
	}

	return inboxes, total, nil
}

// MarkAsRead 标记消息为已读
func (inboxRepository *InboxRepository) MarkAsRead(ctx context.Context, inboxID uint) error {
	result := inboxRepository.getDB(ctx).
		Model(&inboxModel.Inbox{}).
		Where("id = ?", inboxID).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteInbox 删除指定收件箱消息
func (inboxRepository *InboxRepository) DeleteInbox(ctx context.Context, inboxID uint) error {
	result := inboxRepository.getDB(ctx).Delete(&inboxModel.Inbox{}, inboxID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ClearInbox 清空收件箱
func (inboxRepository *InboxRepository) ClearInbox(ctx context.Context) error {
	return inboxRepository.getDB(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&inboxModel.Inbox{}).
		Error
}

// GetUnreadInbox 获取所有未读消息
func (inboxRepository *InboxRepository) GetUnreadInbox(ctx context.Context) ([]*inboxModel.Inbox, error) {
	var inboxes []*inboxModel.Inbox
	if err := inboxRepository.getDB(ctx).
		Where("read = ?", false).
		Order("created_at DESC").
		Find(&inboxes).Error; err != nil {
		return nil, err
	}
	return inboxes, nil
}
