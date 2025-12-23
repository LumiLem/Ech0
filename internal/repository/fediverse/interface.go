package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/fediverse"
)

type FediverseRepositoryInterface interface {
	// 通过用户ID获取粉丝列表
	GetFollowers(userID uint) ([]model.Follower, error)

	// 通过用户ID获取关注列表
	GetFollowing(userID uint) ([]model.Follow, error)

	// 存储新的粉丝
	SaveFollower(ctx context.Context, follower *model.Follower) error

	// 检查粉丝记录是否存在
	FollowerExists(ctx context.Context, userID uint, actor string) (bool, error)

	// 保存或更新新的关注关系
	SaveOrUpdateFollow(ctx context.Context, follow *model.Follow) error

	// 根据用户和目标 Actor 获取关注关系
	GetFollowByUserAndObject(
		ctx context.Context,
		userID uint,
		objectID string,
	) (*model.Follow, error)

	// 删除关注关系
	DeleteFollow(ctx context.Context, followID uint) error

	// UpsertInboxStatus 存储或更新远端推文到本地收件箱
	UpsertInboxStatus(ctx context.Context, status *model.InboxStatus) error

	// ListInboxStatuses 获取本地收件箱推文
	ListInboxStatuses(
		ctx context.Context,
		userID uint,
		page, pageSize int,
	) ([]model.InboxStatus, int64, error)

	// UpdateFollowStatusByActivityID 根据 Follow Activity ID 更新关注状态
	UpdateFollowStatusByActivityID(
		ctx context.Context,
		userID uint,
		activityID, status string,
	) error
}
