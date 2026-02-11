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
	LastInboxId      uint           `json:"lastInboxId"`
	LastTodoId       uint           `json:"lastTodoId"`
	LastTodoRemindAt int64          `json:"lastTodoRemindAt"` // 上次待办提醒的 Unix 时间戳
	HubCounts        map[string]int `json:"hubCounts"`        // server_url -> total_echos
}
