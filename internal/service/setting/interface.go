package service

import (
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
)

type SettingServiceInterface interface {
	// GetSetting 获取设置
	GetSetting(setting *model.SystemSetting) error

	// UpdateSetting 更新设置
	UpdateSetting(userid uint, newSetting *model.SystemSettingDto) error

	// GetCommentSetting 获取评论设置
	GetCommentSetting(setting *model.CommentSetting) error

	// UpdateCommentSetting 更新评论设置
	UpdateCommentSetting(userid uint, newSetting *model.CommentSettingDto) error

	// GetS3Setting 获取 S3 存储设置
	GetS3Setting(userid uint, setting *model.S3Setting) error

	// UpdateS3Setting 更新 S3 存储设置
	UpdateS3Setting(userid uint, newSetting *model.S3SettingDto) error

	// GetOAuth2Setting 获取 OAuth2 设置
	GetOAuth2Setting(userid uint, setting *model.OAuth2Setting, forInternal bool) error

	// UpdateOAuth2Setting 更新 OAuth2 设置
	UpdateOAuth2Setting(userid uint, newSetting *model.OAuth2SettingDto) error

	// GetOAuth2Status 获取 OAuth2 状态
	GetOAuth2Status(status *model.OAuth2Status) error

	// GetAllWebhooks 获取所有 Webhook
	GetAllWebhooks(userid uint) ([]webhookModel.Webhook, error)

	// DeleteWebhook 删除 Webhook
	DeleteWebhook(userid, id uint) error

	// UpdateWebhook 更新 Webhook
	UpdateWebhook(userid uint, id uint, newWebhook *model.WebhookDto) error

	// CreateWebhook 创建 Webhook
	CreateWebhook(userid uint, newWebhook *model.WebhookDto) error

	// ListAccessTokens 列出访问令牌
	ListAccessTokens(userid uint) ([]model.AccessTokenSetting, error)

	// CreateAccessToken 创建访问令牌
	CreateAccessToken(userid uint, newToken *model.AccessTokenSettingDto) (string, error)

	// DeleteAccessToken 删除访问令牌
	DeleteAccessToken(userid, id uint) error

	// GetFediverseSetting 获取联邦网络设置
	GetFediverseSetting(userid uint, setting *model.FediverseSetting) error

	// UpdateFediverseSetting 更新联邦网络设置
	UpdateFediverseSetting(userid uint, newSetting *model.FediverseSettingDto) error

	// GetBackupScheduleSetting 获取备份计划
	GetBackupScheduleSetting(setting *model.BackupSchedule) error

	// UpdateBackupScheduleSetting 更新备份计划
	UpdateBackupScheduleSetting(userid uint, newSetting *model.BackupScheduleDto) error

	// GetAgentSettings 获取 Agent 设置
	GetAgentSettings(setting *model.AgentSetting) error

	// UpdateAgentSettings 更新 Agent 设置
	UpdateAgentSettings(userid uint, newSetting *model.AgentSettingDto) error
}
