package model

import "time"

// PushSubscription 存储 PWA Web Push 订阅信息
type PushSubscription struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	UserID    uint      `json:"user_id"    gorm:"index"` // 所属用户ID
	Endpoint  string    `json:"endpoint"   gorm:"uniqueIndex:idx_endpoint;size:512"`
	P256dh    string    `json:"p256dh"`     // 公钥 base64
	Auth      string    `json:"auth"`       // 验证密钥 base64
	UserAgent string    `json:"user_agent"` // 设备信息
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PwaPushSnapshot 后端观察者快照结构
type PwaPushSnapshot struct {
	LastInboxId       uint           `json:"lastInboxId"`
	LastTodoId        uint           `json:"lastTodoId"`
	LastTodoRemindAt  int64          `json:"lastTodoRemindAt"`  // 上次待办提醒的 Unix 时间戳
	ReadHubCounts     map[string]int `json:"readHubCounts"`     // 用户已读的水位线（用于跨端同步红点）
	NotifiedHubCounts map[string]int `json:"notifiedHubCounts"` // 系统已推的水位线（用于防止重复通知）
}
