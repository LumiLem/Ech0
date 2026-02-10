package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SherClockHolmes/webpush-go"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	pwaRepository "github.com/lin-snow/ech0/internal/repository/pwa"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	inboxService "github.com/lin-snow/ech0/internal/service/inbox"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
)

type PwaService struct {
	pwaRepo        pwaRepository.PwaRepositoryInterface
	kvRepo         keyvalueRepository.KeyValueRepositoryInterface
	inboxService   inboxService.InboxServiceInterface
	todoService    todoService.TodoServiceInterface
	connectService connectService.ConnectServiceInterface
}

func NewPwaService(
	pwaRepo pwaRepository.PwaRepositoryInterface,
	kvRepo keyvalueRepository.KeyValueRepositoryInterface,
	inboxSvc inboxService.InboxServiceInterface,
	todoSvc todoService.TodoServiceInterface,
	connectSvc connectService.ConnectServiceInterface,
) PwaServiceInterface {
	return &PwaService{
		pwaRepo:        pwaRepo,
		kvRepo:         kvRepo,
		inboxService:   inboxSvc,
		todoService:    todoSvc,
		connectService: connectSvc,
	}
}

func (s *PwaService) Subscribe(ctx context.Context, userID uint, sub *pwaModel.PushSubscription) error {
	sub.UserID = userID
	return s.pwaRepo.AddOrUpdateSubscription(ctx, sub)
}

func (s *PwaService) Unsubscribe(ctx context.Context, endpoint string) error {
	return s.pwaRepo.DeleteSubscription(ctx, endpoint)
}

func (s *PwaService) GetVapidPublicKey(ctx context.Context) (string, error) {
	pub, err := s.getOrGenerateVapidKeys(ctx)
	if err != nil {
		return "", err
	}
	return pub, nil
}

func (s *PwaService) getOrGenerateVapidKeys(ctx context.Context) (pub string, err error) {
	pubVal, err := s.kvRepo.GetKeyValue(commonModel.VapidPublicKeyKey)
	if err == nil && pubVal != nil {
		return pubVal.(string), nil
	}

	// 生成新密钥
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return "", err
	}

	_ = s.kvRepo.AddOrUpdateKeyValue(ctx, commonModel.VapidPublicKeyKey, pub)
	_ = s.kvRepo.AddOrUpdateKeyValue(ctx, commonModel.VapidPrivateKeyKey, priv)

	return pub, nil
}

func (s *PwaService) SendPushNotification(ctx context.Context, userID uint, payload interface{}) error {
	subs, err := s.pwaRepo.GetSubscriptionsByUserId(ctx, userID)
	if err != nil || len(subs) == 0 {
		return err
	}

	privKeyVal, err := s.kvRepo.GetKeyValue(commonModel.VapidPrivateKeyKey)
	if err != nil {
		return err
	}
	privKey := privKeyVal.(string)

	pubKeyVal, err := s.kvRepo.GetKeyValue(commonModel.VapidPublicKeyKey)
	if err != nil {
		return err
	}
	pubKey := pubKeyVal.(string)

	payloadBytes, _ := json.Marshal(payload)

	for _, sub := range subs {
		pushSub := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		resp, err := webpush.SendNotification(payloadBytes, pushSub, &webpush.Options{
			Subscriber:      "pwa-notify@ech0.local",
			VAPIDPublicKey:  pubKey,
			VAPIDPrivateKey: privKey,
			TTL:             30,
		})

		if err != nil {
			fmt.Printf("[PWA Push] Error sending to user %d (%s): %v\n", userID, sub.Endpoint, err)
			continue
		}

		fmt.Printf("[PWA Push] Successfully sent notification to user %d, endpoint: %s, status: %d, payload: %s\n", userID, sub.Endpoint, resp.StatusCode, string(payloadBytes))

		// 如果返回 410 Gone 或 404 Not Found，说明订阅已失效，从数据库移除
		if resp.StatusCode == 410 || resp.StatusCode == 404 {
			_ = s.pwaRepo.DeleteSubscription(ctx, sub.Endpoint)
			fmt.Printf("WebPush subscription expired for %s, removed.\n", sub.Endpoint)
		}

		resp.Body.Close()
	}

	return nil
}

func (s *PwaService) ObserverTaskLogic(ctx context.Context) error {
	subs, err := s.pwaRepo.GetAllSubscriptions(ctx)
	if err != nil {
		return err
	}

	userSubMap := make(map[uint]bool)
	for _, sub := range subs {
		userSubMap[sub.UserID] = true
	}

	for userID := range userSubMap {
		s.checkAndNotify(ctx, userID)
	}

	return nil
}

func (s *PwaService) checkAndNotify(ctx context.Context, userID uint) {
	// 1. 获取当前状态 (调用现有的 Service 获取最新数据)
	unreadInboxes, _ := s.inboxService.GetUnreadInbox(userID)

	// 对于 Todo，我们需要获取该用户的全部待办并过滤未完成
	allTodos, _ := s.todoService.GetTodoList(userID)
	var incompleteTodos []struct {
		ID      uint
		Content string
	}
	for _, t := range allTodos {
		if t.Status == 0 {
			incompleteTodos = append(incompleteTodos, struct {
				ID      uint
				Content string
			}{ID: uint(t.ID), Content: t.Content})
		}
	}

	// 对于 Hub 更新 (跨站)
	connects, _ := s.connectService.GetConnectsInfo()

	lastInboxId := uint(0)
	for _, item := range unreadInboxes {
		if item.ID > lastInboxId {
			lastInboxId = item.ID
		}
	}

	lastTodoId := uint(0)
	for _, item := range incompleteTodos {
		if item.ID > lastTodoId {
			lastTodoId = item.ID
		}
	}

	hubCounts := make(map[string]int)
	for _, c := range connects {
		hubCounts[c.ServerURL] = c.TotalEchos
	}

	// 2. 获取快照
	snapshotKey := fmt.Sprintf("%s%d", commonModel.PwaPushSnapshotPrefix, userID)
	snapshotStr, err := s.kvRepo.GetKeyValue(snapshotKey)

	var snapshot pwaModel.PwaPushSnapshot
	isFirstTime := false
	if err == nil && snapshotStr != nil {
		_ = json.Unmarshal([]byte(snapshotStr.(string)), &snapshot)
	} else {
		isFirstTime = true
		snapshot.HubCounts = make(map[string]int)
	}

	// 3. 对齐逻辑并推送
	if !isFirstTime {
		// Inbox
		for _, item := range unreadInboxes {
			if item.ID > snapshot.LastInboxId {
				body := item.Content
				if len(body) > 100 {
					body = body[:97] + "..."
				}
				_ = s.SendPushNotification(ctx, userID, map[string]interface{}{
					"title": fmt.Sprintf("📩 来自 %s 的新消息", item.Source),
					"body":  body,
					"tag":   fmt.Sprintf("inbox-%d", item.ID),
					"data": map[string]interface{}{
						"url":     "/?mode=inbox",
						"type":    "inbox",
						"inboxId": item.ID,
					},
					"actions": []map[string]string{
						{"action": "inbox-read", "title": "设为已读"},
					},
				})
			}
		}

		// Todo
		for _, item := range incompleteTodos {
			if item.ID > snapshot.LastTodoId {
				_ = s.SendPushNotification(ctx, userID, map[string]interface{}{
					"title": "📋 待办事项提醒",
					"body":  item.Content,
					"tag":   fmt.Sprintf("todo-%d", item.ID),
					"data": map[string]interface{}{
						"url":    "/?mode=todo",
						"type":   "todo",
						"todoId": item.ID,
					},
					"actions": []map[string]string{
						{"action": "todo-done", "title": "完成任务"},
					},
				})
			}
		}

		// Hub
		updatedHubs := []struct {
			Name string
			URL  string
			New  int
		}{}
		for _, c := range connects {
			lastCount := snapshot.HubCounts[c.ServerURL]
			if c.TotalEchos > lastCount {
				updatedHubs = append(updatedHubs, struct {
					Name string
					URL  string
					New  int
				}{Name: c.ServerName, URL: c.ServerURL, New: c.TotalEchos - lastCount})
			}
		}

		if len(updatedHubs) > 0 {
			title := "✨ Hub 发现了新动态"
			if len(updatedHubs) == 1 {
				title = fmt.Sprintf("✨ %s 发布了新动态", updatedHubs[0].Name)
			}

			var body string
			totalNew := 0
			for _, h := range updatedHubs {
				totalNew += h.New
			}

			if len(updatedHubs) == 1 {
				body = fmt.Sprintf("发布了 %d 条新内容", updatedHubs[0].New)
			} else if len(updatedHubs) <= 3 {
				names := ""
				for i, h := range updatedHubs {
					if i > 0 {
						names += "、"
					}
					names += h.Name
				}
				body = fmt.Sprintf("%s 更新了 %d 条动态", names, totalNew)
			} else {
				body = fmt.Sprintf("%s 等 %d 个站点更新了 %d 条动态", updatedHubs[0].Name, len(updatedHubs), totalNew)
			}

			_ = s.SendPushNotification(ctx, userID, map[string]interface{}{
				"title": title,
				"body":  body,
				"tag":   "hub-update",
				"data": map[string]interface{}{
					"url":  "/hub",
					"type": "hub",
				},
				"renotify": true,
			})
		}
	}

	// 4. 保存新快照
	s.saveSnapshot(ctx, snapshotKey, lastInboxId, lastTodoId, hubCounts)
}

func (s *PwaService) saveSnapshot(ctx context.Context, key string, inboxId, todoId uint, hubCounts map[string]int) {
	snap := pwaModel.PwaPushSnapshot{
		LastInboxId: inboxId,
		LastTodoId:  todoId,
		HubCounts:   hubCounts,
	}
	bytes, _ := json.Marshal(snap)
	_ = s.kvRepo.AddOrUpdateKeyValue(ctx, key, string(bytes))
}
