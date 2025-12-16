//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	fediverse "github.com/lin-snow/ech0/internal/fediverse"
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	fediverseHandler "github.com/lin-snow/ech0/internal/handler/fediverse"
	inboxHandler "github.com/lin-snow/ech0/internal/handler/inbox"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	todoHandler "github.com/lin-snow/ech0/internal/handler/todo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
	"github.com/lin-snow/ech0/internal/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	connectRepository "github.com/lin-snow/ech0/internal/repository/connect"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	fediverseRepository "github.com/lin-snow/ech0/internal/repository/fediverse"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	settingRepository "github.com/lin-snow/ech0/internal/repository/setting"
	todoRepository "github.com/lin-snow/ech0/internal/repository/todo"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	backupService "github.com/lin-snow/ech0/internal/service/backup"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fediverseService "github.com/lin-snow/ech0/internal/service/fediverse"
	inboxService "github.com/lin-snow/ech0/internal/service/inbox"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
)

// BuildHandlers 使用wire生成的代码来构建Handlers实例
func BuildHandlers(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*Handlers, error) {
	wire.Build(
		CacheSet,
		TransactionManagerSet,
		WebSet,
		UserSet,
		EchoSet,
		CommonSet,
		WebhookSet,
		KeyValueSet,
		SettingSet,
		InboxSet,
		TodoSet,
		ConnectSet,
		MetricSet,
		MonitorSet,
		DashboardSet,
		AgentSet,
		BackupSet,
		FediverseCoreSet,
		FediverseSet,
		NewHandlers, // NewHandlers 聚合各个模块的 Handler
	)

	return &Handlers{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	wire.Build(
		CacheSet,
		KeyValueSet,
		TransactionManagerSet,
		WebhookSet,
		SettingSet,
		EchoSet,
		CommonSet,
		QueueSet,
		TaskSet,
	)
	return &task.Tasker{}, nil
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
) (*event.EventRegistrar, error) {
	wire.Build(
		EchoSet,
		UserSet,
		TodoSet,
		InboxSet,
		CacheSet,
		TransactionManagerSet,
		KeyValueSet,
		QueueSet,
		WebhookSet,
		FediverseCoreSet,
		FediverseSet,
		EventSet,
	)

	return &event.EventRegistrar{}, nil
}

// CacheSet 包含了构建缓存所需的所有 Provider
var CacheSet = wire.NewSet(
	ProvideCache,
)

// TransactionManagerSet 包含了构建事务管理器所需的所有 Provider
var TransactionManagerSet = wire.NewSet(
	ProvideTransactionManager,
)

// WebSet 包含了构建 WebHandler 所需的所有 Provider
var WebSet = wire.NewSet(
	webHandler.NewWebHandler,
)

// UserSet 包含了构建 UserHandler 所需的所有 Provider
var UserSet = wire.NewSet(
	userRepository.NewUserRepository,
	userService.NewUserService,
	userHandler.NewUserHandler,
)

// EchoSet 包含了构建 EchoHandler 所需的所有 Provider
var EchoSet = wire.NewSet(
	echoRepository.NewEchoRepository,
	echoService.NewEchoService,
	echoHandler.NewEchoHandler,
)

// CommonSet 包含了构建 CommonHandler 所需的所有 Provider
var CommonSet = wire.NewSet(
	commonRepository.NewCommonRepository,
	commonService.NewCommonService,
	commonHandler.NewCommonHandler,
)

// KeyValueSet 包含了构建 KeyValueRepository 所需的所有 Provider
var KeyValueSet = wire.NewSet(
	keyvalueRepository.NewKeyValueRepository,
)

// SettingSet 包含了构建 SettingHandler 所需的所有 Provider
var SettingSet = wire.NewSet(
	settingRepository.NewSettingRepository,
	settingService.NewSettingService,
	settingHandler.NewSettingHandler,
)

// TodoSet 包含了构建 TodoHandler 所需的所有 Provider
var TodoSet = wire.NewSet(
	todoRepository.NewTodoRepository,
	todoService.NewTodoService,
	todoHandler.NewTodoHandler,
)

// ConnectSet 包含了构建 ConnectHandler 所需的所有 Provider
var ConnectSet = wire.NewSet(
	connectRepository.NewConnectRepository,
	connectService.NewConnectService,
	connectHandler.NewConnectHandler,
)

// BackupSet 包含了构建 BackupHandler 所需的所有 Provider
var BackupSet = wire.NewSet(
	backupHandler.NewBackupHandler,
	backupService.NewBackupService,
)

// DashboardSet 包含了构建 DashboardHandler 所需的所有 Provider
var DashboardSet = wire.NewSet(
	dashboardService.NewDashboardService,
	dashboardHandler.NewDashboardHandler,
)

// AgentSet 包含了构建 AgentHandler 所需的所有 Provider
var AgentSet = wire.NewSet(
	agentService.NewAgentService,
	agentHandler.NewAgentHandler,
)

// WebhookSet 包含了构建 WebhookDispatcher 所需的所有 Provider
var WebhookSet = wire.NewSet(
	webhookRepository.NewWebhookRepository,
)

// InboxSet 包含了构建 InboxRepository 所需的所有 Provider
var InboxSet = wire.NewSet(
	inboxRepository.NewInboxRepository,
	inboxService.NewInboxService,
	inboxHandler.NewInboxHandler,
)

// TaskSet 包含了构建 Tasker 所需的所有 Provider
var TaskSet = wire.NewSet(
	task.NewTasker,
)

// QueueSet 包含了构建 Queue 所需的所有 Provider
var QueueSet = wire.NewSet(
	queueRepository.NewQueueRepository,
)

// FediverseCoreSet 包含了构建 FediverseCore 所需的所有 Provider
var FediverseCoreSet = wire.NewSet(
	fediverse.NewFediverseCore,
)

// FediverseSet 包含了构建 Fediverse 所需的所有 Provider
var FediverseSet = wire.NewSet(
	fediverseRepository.NewFediverseRepository,
	fediverseService.NewFediverseService,
	fediverseHandler.NewFediverseHandler,

	event.NewFediverseAgent,
)

// EventSet 包含了构建 Event 相关所需的所有 Provider
var EventSet = wire.NewSet(
	event.NewWebhookDispatcher,
	event.NewBackupScheduler,
	event.NewDeadLetterResolver,
	event.NewAgentProcessor,
	event.NewEventHandlers,
	event.NewEventRegistry,
)

// MetricSet 包含了构建 Metric 相关所需的所有 Provider
var MetricSet = wire.NewSet(
	metric.NewSystemCollector,
)

// MonitorSet 包含了构建 Monitor 相关所需的所有 Provider
var MonitorSet = wire.NewSet(
	monitor.NewMonitor,
)
