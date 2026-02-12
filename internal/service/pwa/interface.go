package service

import (
	"context"

	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
)

type PwaServiceInterface interface {
	Subscribe(ctx context.Context, userID uint, sub *pwaModel.PushSubscription) error
	Unsubscribe(ctx context.Context, endpoint string) error
	GetVapidPublicKey(ctx context.Context) (string, error)
	SendPushNotification(ctx context.Context, userID uint, payload interface{}) (bool, error)
	// ObserverTaskLogic 供 Task 调用的核心逻辑
	ObserverTaskLogic(ctx context.Context) error
	// GetSnapshot 获取用户的推送快照
	GetSnapshot(ctx context.Context, userID uint) (*pwaModel.PwaPushSnapshot, error)
	// UpdateSnapshot 更新用户的推送快照
	UpdateSnapshot(ctx context.Context, userID uint, snapshot *pwaModel.PwaPushSnapshot) error
	// GetAggregatedStatus 获取聚合后的状态（供 SW 调用，减少请求次数）
	GetAggregatedStatus(ctx context.Context, userID uint) (*pwaModel.PwaAggregatedStatus, error)
}
