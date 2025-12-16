// package task declaration to use task related functionalities
package task

import (
	"context"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lin-snow/ech0/internal/backup"
	"github.com/lin-snow/ech0/internal/event"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

type Tasker struct {
	scheduler      gocron.Scheduler
	commonService  commonService.CommonServiceInterface
	settingService settingService.SettingServiceInterface
	eventBus       event.IEventBus
	queueRepo      queueRepository.QueueRepositoryInterface
}

func NewTasker(
	commonService commonService.CommonServiceInterface,
	settingService settingService.SettingServiceInterface,
	eventBusProvider func() event.IEventBus,
	queueRepo queueRepository.QueueRepositoryInterface,
) *Tasker {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logUtil.GetLogger().Error("Failed to create scheduler", zapcore.Field{
			Key:    "error",
			String: err.Error(),
		})
	}

	return &Tasker{
		scheduler:      scheduler,
		commonService:  commonService,
		settingService: settingService,
		eventBus:       eventBusProvider(),
		queueRepo:      queueRepo,
	}
}

func (t *Tasker) Start() {
	t.CleanupTempFilesTask()  // 启动清理临时文件任务
	t.DeadLetterConsumeTask() // 启动死信任务消费任务

	// 读取自动备份cron设置
	var backupScheduleSetting settingModel.BackupSchedule
	if err := t.settingService.GetBackupScheduleSetting(&backupScheduleSetting); err != nil {
		logUtil.GetLogger().Error("Failed to get backup schedule setting", zap.String("error", err.Error()))
		// 默认启用定时备份任务
		backupScheduleSetting.Enable = false
		backupScheduleSetting.CronExpression = "0 2 * * 0" // 每周日2点执行一次
	}
	if backupScheduleSetting.Enable {
		t.ScheduleBackupTask(backupScheduleSetting.CronExpression) // 启动定时备份任务
	}

	t.scheduler.Start()
}

func (t *Tasker) Stop() {
	if err := t.scheduler.Shutdown(); err != nil {
		logUtil.GetLogger().Error("Failed to shutdown scheduler", zap.String("error", err.Error()))
	}
}

// CleanupTempFilesTask 清理过期的临时文件任务
func (t *Tasker) CleanupTempFilesTask() {
	// 每三天执行一次
	_, err := t.scheduler.NewJob(
		gocron.DurationJob(72*time.Hour),
		gocron.NewTask(
			func() {
				if err := t.commonService.CleanupTempFiles(); err != nil {
					logUtil.GetLogger().Error("Failed to clean up temporary files", zap.String("error", err.Error()))
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule CleanupTempFilesTask", zap.String("error", err.Error()))
	}
}

// DeadLetterConsumeTask 死信任务消费任务
func (t *Tasker) DeadLetterConsumeTask() {
	// 每天12点执行一次, 测试时为每30秒执行一次
	_, err := t.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(12, 0, 0))),
		// gocron.DurationJob(30*time.Second), // 测试时为每30秒执行一次
		gocron.NewTask(
			func() {
				// 取出死信队列中的任务，逐个重试
				deadLetters, err := t.queueRepo.ListDeadLetters(10)
				if err != nil {
					logUtil.GetLogger().Error("Failed To Get DeadLetters!", zap.String("error", err.Error()))
				}

				// 遍历死信任务，重新发送事件
				for _, dl := range deadLetters {
					// 发布事件到事件总线，触发重试
					if err := t.eventBus.Publish(context.Background(),
						event.NewEvent(
							event.EventTypeDeadLetterRetried,
							event.EventPayload{
								event.EventPayloadDeadLetter: dl,
							},
						),
					); err != nil {
						logUtil.GetLogger().Error("Failed to publish dead letter retried event", zap.String("error", err.Error()))
					}
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule WebhookRetryTask", zap.String("error", err.Error()))
	}
}

// ScheduleBackupTask 定时备份任务
func (t *Tasker) ScheduleBackupTask(cronExpression string) {
	// 判断 cron 表达式的字段数量来确定是否包含秒字段
	// 5 位表达式（分 时 日 月 周）：withSeconds = false
	// 6 位表达式（秒 分 时 日 月 周）：withSeconds = true
	withSeconds := false
	// 按空格分割 cron 表达式以准确判断字段数量
	fieldCount := len(strings.Fields(cronExpression))
	if fieldCount == 6 {
		withSeconds = true
	}

	_, err := t.scheduler.NewJob(
		gocron.CronJob(cronExpression, withSeconds),
		gocron.NewTask(
			func() {
				// 执行备份
				if path, fileName, err := backup.ExecuteBackup(); err != nil {
					logUtil.GetLogger().Error("Failed to execute scheduled backup",
						zap.String("path", path),
						zap.String("fileName", fileName),
						zap.String("error", err.Error()))
				}

				// 发布备份完成事件
				if err := t.eventBus.Publish(
					context.Background(),
					event.NewEvent(
						event.EventTypeSystemBackup,
						event.EventPayload{
							event.EventPayloadInfo: "System scheduled backup completed",
						},
					),
				); err != nil {
					logUtil.GetLogger().Error("Failed to publish backup completed event", zap.String("error", err.Error()))
				}
			},
		),
		gocron.WithTags("BackupSchedule"),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule ScheduleBackupTask", zap.String("error", err.Error()))
	}
}
