package database

import (
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

// fixOldEchoLayoutData 为旧数据补充默认的布局值（layout 为 NULL 或空字符串时设为 'waterfall'）
func fixOldEchoLayoutData() error {
	db := GetDB()
	if db == nil {
		return errors.New(commonModel.DATABASE_NOT_INITED)
	}

	// 更新所有 layout 为 NULL 或空字符串的 echo 记录为 'waterfall'
	if err := db.Model(&echoModel.Echo{}).
		Where("layout IS NULL OR layout = ''").
		Update("layout", "waterfall").Error; err != nil {
		return err
	}

	return nil
}

// MigrateImageToMedia 将 images 表的数据增量同步到 media 表
func MigrateImageToMedia() error {
	db := GetDB()
	if db == nil {
		return errors.New(commonModel.DATABASE_NOT_INITED)
	}

	// 检查 images 表是否存在
	if !db.Migrator().HasTable("images") {
		return nil
	}

	// 执行增量同步：仅插入 media 表中不存在的 ID
	// 这样即使用户切回原版发布了新内容，再切回来时也会自动合入
	err := db.Exec(`
		INSERT INTO media (id, message_id, media_url, media_type, media_source, object_key, width, height)
		SELECT i.id, i.message_id, i.image_url, 'image', i.image_source, i.object_key, i.width, i.height
		FROM images i
		LEFT JOIN media m ON i.id = m.id
		WHERE m.id IS NULL
	`).Error

	if err != nil {
		return err
	}

	// 【重要】我们不再删除 images 表，以保持分之间的兼容性
	// 用户可以手动在设置页面清理该表
	// if err := db.Migrator().DropTable("images"); err != nil {
	// 	return err
	// }

	return nil
}

// UpdateMigration 执行旧数据库迁移和数据修复任务
func UpdateMigration() error {
	var err error

	err = fixOldEchoLayoutData()
	if err != nil {
		return err
	}

	err = MigrateImageToMedia()
	if err != nil {
		return err
	}

	return nil
}
