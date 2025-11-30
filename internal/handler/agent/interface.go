package handler

import "github.com/gin-gonic/gin"

type AgentHandlerInterface interface {
	// 定义 Agent 处理器接口方法
	GetRecent() gin.HandlerFunc
}
