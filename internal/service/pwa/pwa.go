package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	pwaRepository "github.com/lin-snow/ech0/internal/repository/pwa"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	inboxService "github.com/lin-snow/ech0/internal/service/inbox"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
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
			Subscriber:      "ech0@lumlime.cn", // webpush-go 库会自动添加 mailto: 前缀
			VAPIDPublicKey:  pubKey,
			VAPIDPrivateKey: privKey,
			TTL:             30,
		})

		if err != nil {
			fmt.Printf("[PWA Push] Error sending to user %d (%s): %v\n", userID, sub.Endpoint, err)
			continue
		}

		// 读取响应体（Apple 等推送服务会在非 2xx 时返回 {"reason":"..."} 的 JSON）
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("[PWA Push] ✅ 推送成功 user=%d endpoint=%s status=%d payload=%s\n",
				userID, sub.Endpoint, resp.StatusCode, string(payloadBytes))
		} else {
			fmt.Printf("[PWA Push] ❌ 推送失败 user=%d endpoint=%s status=%d reason=%s payload=%s\n",
				userID, sub.Endpoint, resp.StatusCode, string(respBody), string(payloadBytes))
		}

		// 如果返回 410 Gone 或 404 Not Found，说明订阅已失效，从数据库移除
		if resp.StatusCode == 410 || resp.StatusCode == 404 {
			_ = s.pwaRepo.DeleteSubscription(ctx, sub.Endpoint)
			fmt.Printf("[PWA Push] 🗑️ 已清除失效订阅: %s\n", sub.Endpoint)
		}
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

	// 1. 预先抓取所有 Hub 站点的当前信息（全量一份，避免用户循环内重抓）
	connects, _ := s.connectService.GetConnectsInfo()

	for userID := range userSubMap {
		s.checkAndNotify(ctx, userID, connects)
	}

	return nil
}

func (s *PwaService) checkAndNotify(ctx context.Context, userID uint, connects []connectModel.Connect) {
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

	// 2. 获取快照
	snapshotKey := fmt.Sprintf("%s%d", commonModel.PwaPushSnapshotPrefix, userID)
	snapshotStr, err := s.kvRepo.GetKeyValue(snapshotKey)

	var snapshot pwaModel.PwaPushSnapshot
	isFirstTime := false
	if err == nil && snapshotStr != nil {
		_ = json.Unmarshal([]byte(snapshotStr.(string)), &snapshot)
	} else {
		isFirstTime = true
	}

	// 确保 Map 已初始化
	if snapshot.ReadHubCounts == nil {
		snapshot.ReadHubCounts = make(map[string]int)
	}
	if snapshot.NotifiedHubCounts == nil {
		snapshot.NotifiedHubCounts = make(map[string]int)
	}

	// 计算最新的 ID 水位线 (单调递增，防止倒退)
	lastInboxId := snapshot.LastInboxId
	for _, item := range unreadInboxes {
		if item.ID > lastInboxId {
			lastInboxId = item.ID
		}
	}

	lastTodoId := snapshot.LastTodoId
	for _, item := range incompleteTodos {
		if item.ID > lastTodoId {
			lastTodoId = item.ID
		}
	}

	currentHubCounts := make(map[string]int)
	for _, c := range connects {
		currentHubCounts[c.ServerURL] = c.TotalEchos
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

		// Todo: 未完成待办每 4 小时提醒一次（新增待办由前台通知处理）
		const todoRemindInterval int64 = 4 * 60 * 60 // 4 小时（秒）
		now := time.Now().Unix()
		if len(incompleteTodos) > 0 && (now-snapshot.LastTodoRemindAt >= todoRemindInterval) {
			for _, item := range incompleteTodos {
				_ = s.SendPushNotification(ctx, userID, map[string]interface{}{
					"title":    "⏰ 待办事项未完成",
					"body":     item.Content,
					"tag":      fmt.Sprintf("todo-%d", item.ID),
					"renotify": true,
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
			snapshot.LastTodoRemindAt = now
		}

		// Hub
		updatedHubs := []struct {
			Name string
			URL  string
			Logo string
			New  int
		}{}
		for _, c := range connects {
			lastCount, exists := snapshot.NotifiedHubCounts[c.ServerURL]
			if !exists {
				// 新站点：仅记录基准，不触发通知
				continue
			}
			if c.TotalEchos > lastCount {
				updatedHubs = append(updatedHubs, struct {
					Name string
					URL  string
					Logo string
					New  int
				}{Name: c.ServerName, URL: c.ServerURL, Logo: c.Logo, New: c.TotalEchos - lastCount})
				// 更新通知水位线
				snapshot.NotifiedHubCounts[c.ServerURL] = c.TotalEchos
			} else if c.TotalEchos < lastCount {
				// 自动下调校准
				snapshot.NotifiedHubCounts[c.ServerURL] = c.TotalEchos
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
				// 与前端一致：单站更新时先尝试获取最新动态内容
				latestContent := fetchLatestEchoContent(updatedHubs[0].URL)
				if latestContent != "" {
					runes := []rune(latestContent)
					if len(runes) > 50 {
						body = string(runes[:50]) + "..."
					} else {
						body = latestContent
					}
				} else {
					body = fmt.Sprintf("发布了 %d 条新内容", updatedHubs[0].New)
				}
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
				// 与前端一致：超过3个站点时显示前两个名字
				firstTwo := updatedHubs[0].Name + "、" + updatedHubs[1].Name
				body = fmt.Sprintf("%s 等 %d 个站点更新了 %d 条动态", firstTwo, len(updatedHubs), totalNew)
			}

			notification := map[string]interface{}{
				"title": title,
				"body":  body,
				"tag":   "hub-update",
				"data": map[string]interface{}{
					"url":  "/hub",
					"type": "hub",
				},
				"renotify": true,
			}

			// 与前端一致：单站更新时使用该站的 Logo 作为图标
			if len(updatedHubs) == 1 && updatedHubs[0].Logo != "" && strings.HasPrefix(updatedHubs[0].Logo, "http") {
				notification["icon"] = updatedHubs[0].Logo
			}

			_ = s.SendPushNotification(ctx, userID, notification)
		}
	}

	// 4. 保存新快照 (更新后的 NotifiedHubCounts)
	s.saveSnapshot(ctx, snapshotKey, lastInboxId, lastTodoId, snapshot.LastTodoRemindAt, snapshot.ReadHubCounts, snapshot.NotifiedHubCounts)
}

func (s *PwaService) saveSnapshot(ctx context.Context, key string, inboxId, todoId uint, lastTodoRemindAt int64, readCounts, notifiedCounts map[string]int) {
	snap := pwaModel.PwaPushSnapshot{
		LastInboxId:       inboxId,
		LastTodoId:        todoId,
		LastTodoRemindAt:  lastTodoRemindAt,
		ReadHubCounts:     readCounts,
		NotifiedHubCounts: notifiedCounts,
	}
	bytes, _ := json.Marshal(snap)
	_ = s.kvRepo.AddOrUpdateKeyValue(ctx, key, string(bytes))
}

// GetSnapshot 获取用户的推送快照（供前端和 SW 读取）
func (s *PwaService) GetSnapshot(ctx context.Context, userID uint) (*pwaModel.PwaPushSnapshot, error) {
	snapshotKey := fmt.Sprintf("%s%d", commonModel.PwaPushSnapshotPrefix, userID)
	snapshotStr, err := s.kvRepo.GetKeyValue(snapshotKey)

	var snapshot pwaModel.PwaPushSnapshot
	if err == nil && snapshotStr != nil {
		_ = json.Unmarshal([]byte(snapshotStr.(string)), &snapshot)
	}

	if snapshot.ReadHubCounts == nil {
		snapshot.ReadHubCounts = make(map[string]int)
	}
	if snapshot.NotifiedHubCounts == nil {
		snapshot.NotifiedHubCounts = make(map[string]int)
	}

	return &snapshot, nil
}

// UpdateSnapshot 更新用户的推送快照（供前端和 SW 写入）
func (s *PwaService) UpdateSnapshot(ctx context.Context, userID uint, snapshot *pwaModel.PwaPushSnapshot) error {
	snapshotKey := fmt.Sprintf("%s%d", commonModel.PwaPushSnapshotPrefix, userID)
	bytes, _ := json.Marshal(snapshot)
	return s.kvRepo.AddOrUpdateKeyValue(ctx, snapshotKey, string(bytes))
}

// GetAggregatedStatus 获取聚合后的状态
func (s *PwaService) GetAggregatedStatus(ctx context.Context, userID uint) (*pwaModel.PwaAggregatedStatus, error) {
	// 1. 获取基础数据
	unreadInboxes, _ := s.inboxService.GetUnreadInbox(userID)
	allTodos, _ := s.todoService.GetTodoList(userID)
	connects, _ := s.connectService.GetConnectsInfo()
	snapshot, _ := s.GetSnapshot(ctx, userID)

	res := &pwaModel.PwaAggregatedStatus{
		HasUpdate:     false,
		Notifications: []pwaModel.PwaNotification{},
		Snapshot:      *snapshot,
		InboxCount:    len(unreadInboxes),
	}

	// 2. 统计未完成 Todo
	todoCount := 0
	incompleteTodos := []struct {
		ID      uint
		Content string
	}{}
	for _, t := range allTodos {
		if t.Status == 0 {
			todoCount++
			incompleteTodos = append(incompleteTodos, struct {
				ID      uint
				Content string
			}{ID: uint(t.ID), Content: t.Content})
		}
	}
	res.TodoCount = todoCount

	// 3. 计算 Hub 差异
	hubDiff := 0
	updatedHubs := []struct {
		Name string
		URL  string
		Logo string
		New  int
	}{}
	for _, c := range connects {
		lastRead := snapshot.ReadHubCounts[c.ServerURL]
		if c.TotalEchos > lastRead {
			hubDiff += (c.TotalEchos - lastRead)
		}

		lastNotified := snapshot.NotifiedHubCounts[c.ServerURL]
		if c.TotalEchos > lastNotified {
			updatedHubs = append(updatedHubs, struct {
				Name string
				URL  string
				Logo string
				New  int
			}{Name: c.ServerName, URL: c.ServerURL, Logo: c.Logo, New: c.TotalEchos - lastNotified})
		}
	}
	res.HubDiff = hubDiff

	// 4. 构建通知内容 (Inbox)
	for _, item := range unreadInboxes {
		if item.ID > snapshot.LastInboxId {
			res.HasUpdate = true
			body := item.Content
			if len(body) > 100 {
				body = body[:97] + "..."
			}
			res.Notifications = append(res.Notifications, pwaModel.PwaNotification{
				Title: fmt.Sprintf("📩 来自 %s 的新消息", item.Source),
				Body:  body,
				Tag:   fmt.Sprintf("inbox-%d", item.ID),
				Icon:  "/icons/notification-inbox.png",
				Data: map[string]interface{}{
					"url":     "/?mode=inbox",
					"type":    "inbox",
					"inboxId": item.ID,
				},
			})
			snapshot.LastInboxId = item.ID // 临时更新快照水位，供后续持久化
		}
	}

	// 5. 构建通知内容 (Todo)
	const todoRemindInterval int64 = 4 * 60 * 60
	now := time.Now().Unix()
	if len(incompleteTodos) > 0 && (now-snapshot.LastTodoRemindAt >= todoRemindInterval) {
		res.HasUpdate = true
		for _, item := range incompleteTodos {
			res.Notifications = append(res.Notifications, pwaModel.PwaNotification{
				Title: "⏰ 待办事项未完成",
				Body:  item.Content,
				Tag:   fmt.Sprintf("todo-%d", item.ID),
				Icon:  "/icons/notification-todo.png",
				Data: map[string]interface{}{
					"url":    "/?mode=todo",
					"type":   "todo",
					"todoId": item.ID,
				},
			})
		}
		snapshot.LastTodoRemindAt = now
	}

	// 6. 构建通知内容 (Hub)
	if len(updatedHubs) > 0 {
		res.HasUpdate = true
		title := "✨ Hub 发现了新动态"
		if len(updatedHubs) == 1 {
			title = fmt.Sprintf("✨ %s 发布了新动态", updatedHubs[0].Name)
		}

		var body string
		totalNew := 0
		for _, h := range updatedHubs {
			totalNew += h.New
		}

		icon := "/icons/notification-hub.png"
		if len(updatedHubs) == 1 {
			latestContent := fetchLatestEchoContent(updatedHubs[0].URL)
			if latestContent != "" {
				runes := []rune(latestContent)
				if len(runes) > 50 {
					body = string(runes[:50]) + "..."
				} else {
					body = latestContent
				}
			} else {
				body = fmt.Sprintf("发布了 %d 条新内容", updatedHubs[0].New)
			}
			if updatedHubs[0].Logo != "" && strings.HasPrefix(updatedHubs[0].Logo, "http") {
				icon = updatedHubs[0].Logo
			}
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
			firstTwo := updatedHubs[0].Name + "、" + updatedHubs[1].Name
			body = fmt.Sprintf("%s 等 %d 个站点更新了 %d 条动态", firstTwo, len(updatedHubs), totalNew)
		}

		res.Notifications = append(res.Notifications, pwaModel.PwaNotification{
			Title: title,
			Body:  body,
			Tag:   "hub-update",
			Icon:  icon,
			Data: map[string]interface{}{
				"url":  "/hub",
				"type": "hub",
			},
		})

		// 更新快照中的通知水位线
		for _, c := range connects {
			snapshot.NotifiedHubCounts[c.ServerURL] = c.TotalEchos
		}
	}

	// 7. 更新快照 (如果内容发生了变化，且是后端触发的通知)
	if res.HasUpdate {
		res.Snapshot = *snapshot // 同步返回内存中计算出的新快照
		// 注意：后端快照在此接口仅供 SW "预览" 并由回复的 updateSnapshot 逻辑决定是否正式持久化
		// 但为了保险，我们在这里根据业务逻辑可以选则直接保存一次
		s.UpdateSnapshot(ctx, userID, snapshot)
	}

	return res, nil
}

// fetchLatestEchoContent 获取指定站点的最新一条动态内容
// 与前端 fetchLatestEchoContent 逻辑一致
func fetchLatestEchoContent(serverURL string) string {
	apiURL := httpUtil.TrimURL(serverURL) + "/api/echo/page?page=1&pageSize=1"

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				Content string `json:"content"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	if result.Code == 1 && len(result.Data.Items) > 0 {
		return result.Data.Items[0].Content
	}

	return ""
}
