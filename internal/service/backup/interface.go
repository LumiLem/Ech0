package service

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type BackupServiceInterface interface {
	// Backup 执行备份
	Backup(userid uint) error

	// ExportBackup 导出备份
	ExportBackup(ctx *gin.Context, userid uint) error

	// 恢复备份
	ImportBackup(ctx *gin.Context, userid uint, file *multipart.FileHeader) error

	// 数据库维护兼容性 (原版)
	SyncToLegacyTable(ctx context.Context, userid uint) error
	CleanLegacyTable(ctx context.Context, userid uint) error
}
