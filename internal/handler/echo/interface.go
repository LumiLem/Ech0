package handler

import "github.com/gin-gonic/gin"

type EchoHandlerInterface interface {
	// PostEcho 发布新的 Echo
	PostEcho() gin.HandlerFunc

	// GetEchosByPage 获取 Echo 列表，支持分页
	GetEchosByPage() gin.HandlerFunc

	// DeleteEcho 删除 Echo
	DeleteEcho() gin.HandlerFunc

	// GetTodayEchos 获取今天的 Echo 列表
	GetTodayEchos() gin.HandlerFunc

	// UpdateEcho 更新 Echo
	UpdateEcho() gin.HandlerFunc

	// LikeEcho 点赞 Echo
	LikeEcho() gin.HandlerFunc

	// GetEchoById 获取指定 ID 的 Echo
	GetEchoById() gin.HandlerFunc

	// GetAllTags 获取所有标签
	GetAllTags() gin.HandlerFunc

	// DeleteTag 删除标签
	DeleteTag() gin.HandlerFunc

	// GetEchosByTagId 获取指定标签 ID 的 Echo 列表
	GetEchosByTagId() gin.HandlerFunc

	// GetEchosByDate 获取指定日期的 Echo 列表
	GetEchosByDate() gin.HandlerFunc
}
