package model

type Inbox struct {
	ID        uint   `gorm:"primaryKey"                json:"id"`             // 收件箱消息ID
	Source    string `gorm:"type:varchar(50);not null" json:"source"`         // 消息来源: system/user/agent
	Content   string `gorm:"type:text"                 json:"content"`        // 消息内容
	Type      string `gorm:"type:varchar(50);not null" json:"type"`           // 消息类型: echo/notification...
	Read      bool   `gorm:"default:false"             json:"read"`           // 是否已读
	Meta      string `gorm:"type:text"                 json:"meta,omitempty"` // 额外元数据 (JSON格式)
	CreatedAt int64  `                                 json:"created_at"`     // 创建时间 (Unix时间戳)
}
