package handler

import "github.com/gin-gonic/gin"

type SettingHandlerInterface interface {
	// GetSettings 获取设置
	GetSettings() gin.HandlerFunc

	// UpdateSettings 更新设置
	UpdateSettings() gin.HandlerFunc

	// GetCommentSettings 获取评论设置
	GetCommentSettings() gin.HandlerFunc

	// UpdateCommentSettings 更新评论设置
	UpdateCommentSettings() gin.HandlerFunc

	// GetS3Settings 获取 S3 存储设置
	GetS3Settings() gin.HandlerFunc

	// UpdateS3Settings 更新 S3 存储设置
	UpdateS3Settings() gin.HandlerFunc

	// GetOAuth2Settings 获取 OAuth2 设置
	GetOAuth2Settings() gin.HandlerFunc

	// UpdateOAuth2Settings 更新 OAuth2 设置
	UpdateOAuth2Settings() gin.HandlerFunc

	// GetOAuth2Status 获取 OAuth2 状态
	GetOAuth2Status() gin.HandlerFunc

	// GetWebhook 获取所有 Webhook
	GetWebhook() gin.HandlerFunc

	// DeleteWebhook 删除 Webhook
	DeleteWebhook() gin.HandlerFunc

	// UpdateWebhook 更新 Webhook
	UpdateWebhook() gin.HandlerFunc

	// CreateWebhook 创建 Webhook
	CreateWebhook() gin.HandlerFunc

	// ListAccessTokens 列出访问令牌
	ListAccessTokens() gin.HandlerFunc

	// CreateAccessToken 创建访问令牌
	CreateAccessToken() gin.HandlerFunc

	// DeleteAccessToken 删除访问令牌
	DeleteAccessToken() gin.HandlerFunc

	// GetFediverseSettings 获取联邦网络设置
	GetFediverseSettings() gin.HandlerFunc

	// UpdateFediverseSettings 更新联邦网络设置
	UpdateFediverseSettings() gin.HandlerFunc

	// GetBackupScheduleSetting 获取备份计划
	GetBackupScheduleSetting() gin.HandlerFunc

	// UpdateBackupScheduleSetting 更新备份计划
	UpdateBackupScheduleSetting() gin.HandlerFunc

	// GetAgentSettings 获取 Agent 设置
	GetAgentSettings() gin.HandlerFunc

	// UpdateAgentSettings 更新 Agent 设置
	UpdateAgentSettings() gin.HandlerFunc

	// GetAgentInfo 获取 Agent 信息
	GetAgentInfo() gin.HandlerFunc
}
