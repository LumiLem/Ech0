package repository

import (
	"context"

	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type PwaRepository struct {
	db func() *gorm.DB
}

func NewPwaRepository(dbProvider func() *gorm.DB) PwaRepositoryInterface {
	return &PwaRepository{
		db: dbProvider,
	}
}

func (r *PwaRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return r.db()
}

func (r *PwaRepository) AddOrUpdateSubscription(ctx context.Context, sub *pwaModel.PushSubscription) error {
	var existing pwaModel.PushSubscription
	err := r.getDB(ctx).Where("endpoint = ?", sub.Endpoint).First(&existing).Error
	if err == nil {
		// 更新现有订阅
		sub.ID = existing.ID
		sub.CreatedAt = existing.CreatedAt
		return r.getDB(ctx).Save(sub).Error
	}
	return r.getDB(ctx).Create(sub).Error
}

func (r *PwaRepository) GetSubscriptionsByUserId(ctx context.Context, userID uint) ([]*pwaModel.PushSubscription, error) {
	var subs []*pwaModel.PushSubscription
	err := r.getDB(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func (r *PwaRepository) GetAllSubscriptions(ctx context.Context) ([]*pwaModel.PushSubscription, error) {
	var subs []*pwaModel.PushSubscription
	err := r.getDB(ctx).Find(&subs).Error
	return subs, err
}

func (r *PwaRepository) DeleteSubscription(ctx context.Context, endpoint string) error {
	return r.getDB(ctx).Where("endpoint = ?", endpoint).Delete(&pwaModel.PushSubscription{}).Error
}
