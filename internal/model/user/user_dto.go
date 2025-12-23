package model

// UserInfoDto 用户信息数据传输对象
//
// swagger:model UserInfoDto
type UserInfoDto struct {
	// 用户名
	// example: linsnow
	Username string `json:"username"`

	// 密码
	// example: 123456
	Password string `json:"password"`

	// 是否为管理员
	// example: false
	IsAdmin bool `json:"is_admin"`

	// 头像地址
	// example: https://example.com/avatar.png
	Avatar string `json:"avatar"`
}

// OAuthInfoDto OAuth2 信息数据传输对象
type OAuthInfoDto struct {
	Provider string `json:"provider"`
	UserID   uint   `json:"user_id"`
	OAuthID  string `json:"oauth_id"`
	Issuer   string `json:"issuer"`
	AuthType string `json:"auth_type"`
}
