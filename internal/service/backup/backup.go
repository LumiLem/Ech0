package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/backup"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/event"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// ImageLegacy 映射到原版的 images 表，用于数据兼容维护
type ImageLegacy struct {
	ID          uint   `gorm:"primaryKey"`
	MessageID   uint   `gorm:"index;not null"`
	ImageURL    string `gorm:"type:text"`
	ImageSource string `gorm:"type:varchar(20)"`
	ObjectKey   string `gorm:"type:text"`
	Width       int    `gorm:"default:0"`
	Height      int    `gorm:"default:0"`
}

func (ImageLegacy) TableName() string {
	return "images"
}

type BackupService struct {
	commonService commonService.CommonServiceInterface
	eventBus      event.IEventBus
}

func NewBackupService(
	commonService commonService.CommonServiceInterface,
	eventBusProvider func() event.IEventBus,
) BackupServiceInterface {
	return &BackupService{
		commonService: commonService,
		eventBus:      eventBusProvider(),
	}
}

// Backup 执行备份
func (backupService *BackupService) Backup(userid uint) error {
	user, err := backupService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 执行备份
	if _, _, err := backup.ExecuteBackup(); err != nil {
		return err
	}

	// 触发备份完成事件
	if err := backupService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeSystemBackup,
			event.EventPayload{
				event.EventPayloadInfo: "System backup completed",
			},
		),
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish system backup completed event", zap.String("error", err.Error()))
	}

	return nil
}

// ExportBackup 导出备份
func (backupService *BackupService) ExportBackup(ctx *gin.Context, userid uint) error {
	// 鉴权
	user, err := backupService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 导出备份
	// 1. 先备份
	var backupFilePath string // 备份文件路径

	backupFilePath, _, err = backup.ExecuteBackup()
	if err != nil {
		return err
	}

	// 2. 计算文件大小
	fileInfo, err := os.Stat(backupFilePath)
	if err != nil {
		return err
	}

	// 设置响应头
	filename := fmt.Sprintf("ech0-backup-%s.zip", time.Now().UTC().Format("2006-01-02-150405"))

	// 设置响应头的顺序很重要
	ctx.Writer.Header().Set("Content-Type", "application/zip")
	ctx.Writer.Header().
		Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	ctx.Writer.Header().Set("Accept-Ranges", "bytes")
	ctx.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// ✅ 立即刷新响应头到客户端
	ctx.Writer.WriteHeader(200)

	// 使用 Gin 的内置方法，支持 Range 请求
	ctx.File(backupFilePath)

	// 触发导出完成事件
	if err := backupService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeSystemExport,
			event.EventPayload{
				event.EventPayloadInfo: "System export completed",
				event.EventPayloadSize: fileInfo.Size(),
			},
		),
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish system export completed event", zap.String("error", err.Error()))
	}

	return nil
}

// ImportBackup 恢复备份
func (backupService *BackupService) ImportBackup(
	ctx *gin.Context,
	userid uint,
	file *multipart.FileHeader,
) error {
	user, err := backupService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 保存上传的文件到临时位置, (./temp/snapshot_时间戳.zip)
	timestamp := time.Now().UTC().Unix()
	tempFilePath := fmt.Sprintf("./temp/snapshot_%d.zip", timestamp)
	if err := ctx.SaveUploadedFile(file, tempFilePath); err != nil {
		return errors.New(commonModel.SNAPSHOT_UPLOAD_FAILED + ": " + err.Error())
	}

	// 执行恢复
	if err := backup.ExcuteRestoreOnline(tempFilePath, timestamp); err != nil {
		return errors.New(commonModel.SNAPSHOT_RESTORE_FAILED + ": " + err.Error())
	}

	// 触发恢复完成事件
	if err := backupService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeSystemRestore,
			event.EventPayload{
				event.EventPayloadInfo: "System restore completed",
			},
		),
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish system restore completed event", zap.String("error", err.Error()))
	}

	return nil
}

// SyncToLegacyTable 将 media 中的数据同步回 images 表，供回退到原版使用
func (backupService *BackupService) SyncToLegacyTable(ctx context.Context, userid uint) error {
	// 鉴权
	user, err := backupService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	db := database.GetDB()

	// 1. 自动迁移 images 表（ImageLegacy 已通过 TableName 映射）
	if err := db.AutoMigrate(&ImageLegacy{}); err != nil {
		return err
	}

	// 2. 将 media 表中的图片数据通过 SQL 迁移回 images
	// 使用 NOT EXISTS 避免重复插入
	err = db.Exec(`
		INSERT INTO images (id, message_id, image_url, image_source, object_key, width, height)
		SELECT m.id, m.message_id, m.media_url, m.media_source, m.object_key, m.width, m.height
		FROM media m
		WHERE m.media_type = 'image' 
		AND NOT EXISTS (SELECT 1 FROM images i WHERE i.id = m.id)
	`).Error

	return err
}

// CleanLegacyTable 清理原版兼容表 images
func (backupService *BackupService) CleanLegacyTable(ctx context.Context, userid uint) error {
	// 鉴权
	user, err := backupService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	db := database.GetDB()
	if db.Migrator().HasTable("images") {
		return db.Migrator().DropTable("images")
	}
	return nil
}
