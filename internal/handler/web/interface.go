package handler

import "github.com/gin-gonic/gin"

type WebHandlerInterface interface {
	// Templates 返回前端Web项目编译后的静态文件
	Templates() gin.HandlerFunc
	// HandleDynamicIcon 处理动态图标生成请求
	HandleDynamicIcon(ctx *gin.Context)
}
