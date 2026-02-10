package service

import (
	"context"

	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
)

type PwaServiceInterface interface {
	Subscribe(ctx context.Context, userID uint, sub *pwaModel.PushSubscription) error
	Unsubscribe(ctx context.Context, endpoint string) error
	GetVapidPublicKey(ctx context.Context) (string, error)
	SendPushNotification(ctx context.Context, userID uint, payload interface{}) error
	// ObserverTaskLogic 供 Task 调用的核心逻辑
	ObserverTaskLogic(ctx context.Context) error
}
