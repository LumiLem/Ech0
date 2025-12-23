package service

import (
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
)

type UserServiceInterface interface {
	// Login 用户登录
	Login(user *authModel.LoginDto) (string, error)

	// GetUserByID 根据用户ID获取用户信息
	GetUserByID(userId int) (model.User, error)

	// Register 用户注册
	Register(registerDto *authModel.RegisterDto) error

	// UpdateUser 更新用户信息
	UpdateUser(userid uint, userdto model.UserInfoDto) error

	// UpdateUserAdmin 更新用户的管理员权限
	UpdateUserAdmin(userid uint, id uint) error

	// GetAllUsers 获取所有用户
	GetAllUsers() ([]model.User, error)

	// GetSysAdmin 获取系统管理员
	GetSysAdmin() (model.User, error)

	// DeleteUser 删除用户
	DeleteUser(userid, id uint) error

	// BindOAuth 绑定 OAuth2 账号
	BindOAuth(userID uint, provider string, redirectURI string) (string, error)

	// GetOAuthLoginURL 获取 OAuth2 登录 URL
	GetOAuthLoginURL(provider string, redirectURI string) (string, error)

	// HandleOAuthCallback 处理 OAuth2 回调
	HandleOAuthCallback(provider string, code string, state string) string

	// GetOAuthInfo 获取 OAuth2 配置信息
	GetOAuthInfo(userId uint, provider string) (model.OAuthInfoDto, error)

	// Passkey / WebAuthn
	PasskeyRegisterBegin(
		userID uint,
		rpID, origin, deviceName string,
	) (authModel.PasskeyRegisterBeginResp, error)
	PasskeyRegisterFinish(userID uint, rpID, origin, nonce string, credential json.RawMessage) error
	PasskeyLoginBegin(rpID, origin string) (authModel.PasskeyLoginBeginResp, error)
	PasskeyLoginFinish(rpID, origin, nonce string, credential json.RawMessage) (string, error)
	ListPasskeys(userID uint) ([]authModel.PasskeyDeviceDto, error)
	DeletePasskey(userID, passkeyID uint) error
	UpdatePasskeyDeviceName(userID, passkeyID uint, deviceName string) error
}
