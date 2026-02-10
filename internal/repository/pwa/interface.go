package repository

import (
	"context"

	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
)

type PwaRepositoryInterface interface {
	AddOrUpdateSubscription(ctx context.Context, sub *pwaModel.PushSubscription) error
	GetSubscriptionsByUserId(ctx context.Context, userID uint) ([]*pwaModel.PushSubscription, error)
	GetAllSubscriptions(ctx context.Context) ([]*pwaModel.PushSubscription, error)
	DeleteSubscription(ctx context.Context, endpoint string) error
}
