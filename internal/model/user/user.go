package model

const (
	USER_NOT_EXISTS_ID = 0
)

// User 定义用户实体
type User struct {
	ID       uint   `gorm:"primaryKey"               json:"id"`
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null"        json:"password"`
	IsAdmin  bool   `gorm:"bool"                     json:"is_admin"`
	Avatar   string `gorm:"size:255"                 json:"avatar"`
}

type OAuthBinding struct {
	ID       uint   `gorm:"primaryKey"              json:"id"`
	UserID   uint   `gorm:"not null;index"          json:"user_id"`
	Provider string `gorm:"size:64;not null;index"  json:"provider"` // 例如 "github"
	OAuthID  string `gorm:"size:255;not null;index" json:"oauth_id"` // 第三方平台的用户ID
}
