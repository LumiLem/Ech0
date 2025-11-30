package service

import "context"

type AgentServiceInterface interface {
	// 定义 Agent 服务接口方法
	GetRecent(ctx context.Context) (string, error)
}
