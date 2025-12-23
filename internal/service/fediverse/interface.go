package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/fediverse"
)

type FediverseServiceInterface interface {
	// WebFinger 处理 Webfinger 请求
	Webfinger(username string) (model.WebFingerResponse, error)

	// GetActorByUsername 通过用户名获取 Actor 信息
	GetActorByUsername(username string) (model.Actor, error)

	// HandleInbox 处理接收到的 ActivityPub 消息
	HandleInbox(username string, activity *model.Activity) error

	// HandleOutbox 构建 Outbox 元信息
	HandleOutbox(username string) (model.OutboxResponse, error)

	// HandleOutbox 处理 Outbox 消息
	HandleOutboxPage(
		ctx context.Context,
		username string,
		page, pageSize int,
	) (model.OutboxPage, error)

	// GetFollowers 获取粉丝列表
	GetFollowers(username string) (model.FollowersResponse, error)

	// GetFollowersPage 获取粉丝列表分页内容
	GetFollowersPage(username string, page, pageSize int) (model.FollowersPage, error)

	// GetFollowing 获取关注列表
	GetFollowing(username string) (model.FollowingResponse, error)

	// GetFollowingPage 获取关注列表分页内容
	GetFollowingPage(username string, page, pageSize int) (model.FollowingPage, error)

	// GetObjectByID 通过 ID 获取内容对象
	GetObjectByID(id uint) (model.Object, error)

	// GetTimeline 获取关注人的时间线
	// GetTimeline(userID uint, page, pageSize int) (commonModel.PageQueryResult[[]model.TimelineItem], error)

	// PushEchoToFediverse 将 Echo 推送到联邦网络
	// PushEchoToFediverse(userId uint, echo echoModel.Echo) error

	// // SearchActorByActorID 根据 Actor URL 搜索远端 Actor
	// SearchActorByActorID(actorID string) (map[string]any, error)

	// // GetFollowStatus 获取关注状态
	// GetFollowStatus(userID uint, targetActor string) (string, error)

	// // FollowActor 发送关注请求
	// FollowActor(userID uint, req model.FollowActionRequest) (map[string]string, error)

	// // UnfollowActor 发送取消关注请求
	// UnfollowActor(userID uint, req model.FollowActionRequest) (map[string]string, error)

	// // LikeObject 发送点赞请求
	// LikeObject(userID uint, req model.LikeActionRequest) (map[string]string, error)

	// // UndoLikeObject 发送取消点赞请求
	// UndoLikeObject(userID uint, req model.LikeActionRequest) (map[string]string, error)
}
