package di

import (
	"github.com/lin-snow/ech0/internal/cache"
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
	"github.com/lin-snow/ech0/internal/transaction"
)

// Handlers 聚合各个模块的Handler
type Handlers struct {
	WebHandler       *webHandler.WebHandler
	UserHandler      *userHandler.UserHandler
	EchoHandler      *echoHandler.EchoHandler
	CommonHandler    *commonHandler.CommonHandler
	SettingHandler   *settingHandler.SettingHandler
	InboxHandler     *inboxHandler.InboxHandler
	TodoHandler      *todoHandler.TodoHandler
	ConnectHandler   *connectHandler.ConnectHandler
	BackupHandler    *backupHandler.BackupHandler
	FediverseHandler *fediverseHandler.FediverseHandler
	DashboardHandler *dashboardHandler.DashboardHandler
	AgentHandler     *agentHandler.AgentHandler
}

// NewHandlers 创建Handlers实例
func NewHandlers(
	webHandler *webHandler.WebHandler,
	userHandler *userHandler.UserHandler,
	echoHandler *echoHandler.EchoHandler,
	commonHandler *commonHandler.CommonHandler,
	settingHandler *settingHandler.SettingHandler,
	inboxHandler *inboxHandler.InboxHandler,
	todoHandler *todoHandler.TodoHandler,
	connectHandler *connectHandler.ConnectHandler,
	backupHandler *backupHandler.BackupHandler,
	fediverseHandler *fediverseHandler.FediverseHandler,
	dashboardHandler *dashboardHandler.DashboardHandler,
	agentHandler *agentHandler.AgentHandler,
) *Handlers {
	return &Handlers{
		WebHandler:       webHandler,
		UserHandler:      userHandler,
		EchoHandler:      echoHandler,
		CommonHandler:    commonHandler,
		SettingHandler:   settingHandler,
		InboxHandler:     inboxHandler,
		TodoHandler:      todoHandler,
		ConnectHandler:   connectHandler,
		BackupHandler:    backupHandler,
		FediverseHandler: fediverseHandler,
		DashboardHandler: dashboardHandler,
		AgentHandler:     agentHandler,
	}
}

// ProvideCache 提供通用缓存实例给 wire 注入
func ProvideCache(factory *cache.CacheFactory) cache.ICache[string, any] {
	return factory.Cache()
}

// ProvideTransactionManager 提供事务管理器实例给 wire 注入
func ProvideTransactionManager(
	factory *transaction.TransactionManagerFactory,
) transaction.TransactionManager {
	return factory.TransactionManager()
}
