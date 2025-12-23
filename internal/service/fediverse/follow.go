package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/lin-snow/ech0/internal/fediverse"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
)

//=======================================
//	处理 Inbox
//=======================================

// HandleInbox 处理接收到的 ActivityPub 消息
func (fediverseService *FediverseService) handleFollowActivity(
	user *userModel.User,
	activity *model.Activity,
) error {
	fmt.Println("Handling follow activity:", activity)

	followerActor := activity.ActorURL
	if followerActor == "" {
		followerActor = activity.ActorID
	}
	if followerActor == "" {
		return errors.New("follow activity missing actor")
	}

	actor, setting, err := fediverseService.core.BuildActor(user)
	if err != nil {
		return err
	}
	serverURL := httpUtil.TrimURL(setting.ServerURL)

	acceptPayload, err := fediverseService.core.BuildAcceptActivityPayload(
		&actor,
		activity,
		followerActor,
		serverURL,
	)
	if err != nil {
		fmt.Printf("Error building accept activity payload: %v\n", err)
		return err
	}

	inboxURL, err := fediverseService.core.FetchRemoteActorInbox(followerActor)
	if err != nil {
		fmt.Printf("Error fetching follower inbox: %v\n", err)
		return err
	}

	if err := httpUtil.PostActivity(acceptPayload, inboxURL, actor.ID); err != nil {
		fmt.Printf("Error posting accept activity: %v\n", err)
		return err
	}

	// 保存粉丝记录
	fmt.Printf("Saving follower: userID=%d, actor=%s\n", user.ID, followerActor)

	// 检查是否已经存在该粉丝记录
	exists, err := fediverseService.fediverseRepository.FollowerExists(
		context.Background(),
		user.ID,
		followerActor,
	)
	if err != nil {
		fmt.Printf("Error checking if follower exists: %v\n", err)
		return err
	}

	// 如果不存在，则保存
	if !exists {
		return fediverseService.txManager.Run(func(ctx context.Context) error {
			return fediverseService.fediverseRepository.SaveFollower(ctx, &model.Follower{
				UserID:  user.ID,
				ActorID: followerActor,
			})
		})
	} else {
		return nil
	}
}

//============================================================
//	处理 Follwer
//============================================================

// GetFollowers 获取粉丝列表
func (fediverseService *FediverseService) GetFollowers(
	username string,
) (model.FollowersResponse, error) {
	actor, followerURLs, err := fediverseService.loadFollowersData(username)
	if err != nil {
		return model.FollowersResponse{}, err
	}

	firstPage := buildFollowersPage(&actor, followerURLs, 1, model.DefaultCollectionPageSize)

	return model.FollowersResponse{
		Context:    "https://www.w3.org/ns/activitystreams",
		ID:         actor.Followers,
		Type:       "OrderedCollection",
		TotalItems: len(followerURLs),
		First:      firstPage,
	}, nil
}

// buildFollowersPage 构建粉丝列表分页
func buildFollowersPage(
	actor *model.Actor,
	followerURLs []string,
	page, pageSize int,
) model.FollowersPage {
	total := len(followerURLs)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pageID := fmt.Sprintf("%s?page=%d", actor.Followers, page)
	next := ""
	if end < total {
		next = fmt.Sprintf("%s?page=%d", actor.Followers, page+1)
	}
	prev := ""
	if page > 1 {
		prev = fmt.Sprintf("%s?page=%d", actor.Followers, page-1)
	}

	orderedItems := followerURLs[start:end]

	return model.FollowersPage{
		Context:      "https://www.w3.org/ns/activitystreams",
		ID:           pageID,
		Type:         "OrderedCollectionPage",
		PartOf:       actor.Followers,
		Next:         next,
		Prev:         prev,
		OrderedItems: orderedItems,
	}
}

// GetFollowingPage 获取关注列表分页
func (fediverseService *FediverseService) GetFollowersPage(
	username string,
	page, pageSize int,
) (model.FollowersPage, error) {
	page, pageSize = fediverse.NormalizePageParams(page, pageSize)

	actor, followerURLs, err := fediverseService.loadFollowersData(username)
	if err != nil {
		return model.FollowersPage{}, err
	}

	return buildFollowersPage(&actor, followerURLs, page, pageSize), nil
}

// loadFollowersData 加载用户的粉丝数据
func (fediverseService *FediverseService) loadFollowersData(
	username string,
) (model.Actor, []string, error) {
	// 查询用户
	user, err := fediverseService.userRepository.GetUserByUsername(username)
	if err != nil {
		return model.Actor{}, nil, errors.New(commonModel.USER_NOTFOUND)
	}

	// 构建 Actor 对象
	actor, _, err := fediverseService.core.BuildActor(&user)
	if err != nil {
		return model.Actor{}, nil, err
	}

	// 获取粉丝列表
	followers, err := fediverseService.fediverseRepository.GetFollowers(user.ID)
	if err != nil {
		return model.Actor{}, nil, err
	}

	// 提取唯一的粉丝 Actor ID 列表
	return actor, uniqueFollowerActorIDs(followers), nil
}

// uniqueFollowerActorIDs 提取唯一的粉丝 Actor ID 列表,去重并过滤空值
func uniqueFollowerActorIDs(followers []model.Follower) []string {
	if len(followers) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(followers))
	unique := make([]string, 0, len(followers))
	for _, follower := range followers {
		if follower.ActorID == "" {
			continue
		}
		if _, ok := seen[follower.ActorID]; ok {
			continue
		}
		seen[follower.ActorID] = struct{}{}
		unique = append(unique, follower.ActorID)
	}
	return unique
}

//================================================================
//	处理 Following
//================================================================

// GetFollowing 获取关注列表
func (fediverseService *FediverseService) GetFollowing(
	username string,
) (model.FollowingResponse, error) {
	actor, followingURLs, err := fediverseService.loadFollowingData(username)
	if err != nil {
		return model.FollowingResponse{}, err
	}

	firstPage := buildFollowingPage(&actor, followingURLs, 1, model.DefaultCollectionPageSize)

	return model.FollowingResponse{
		Context:    "https://www.w3.org/ns/activitystreams",
		ID:         actor.Following,
		Type:       "OrderedCollection",
		TotalItems: len(followingURLs),
		First:      firstPage,
	}, nil
}

// buildFollowersPage 构建粉丝列表分页
func buildFollowingPage(
	actor *model.Actor,
	followingURLs []string,
	page, pageSize int,
) model.FollowingPage {
	total := len(followingURLs)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pageID := fmt.Sprintf("%s?page=%d", actor.Following, page)
	next := ""
	if end < total {
		next = fmt.Sprintf("%s?page=%d", actor.Following, page+1)
	}
	prev := ""
	if page > 1 {
		prev = fmt.Sprintf("%s?page=%d", actor.Following, page-1)
	}

	orderedItems := followingURLs[start:end]

	return model.FollowingPage{
		Context:      "https://www.w3.org/ns/activitystreams",
		ID:           pageID,
		Type:         "OrderedCollectionPage",
		PartOf:       actor.Following,
		Next:         next,
		Prev:         prev,
		OrderedItems: orderedItems,
	}
}

// GetFollowingPage 获取关注列表分页
func (fediverseService *FediverseService) GetFollowingPage(
	username string,
	page, pageSize int,
) (model.FollowingPage, error) {
	page, pageSize = fediverse.NormalizePageParams(page, pageSize)

	actor, followingURLs, err := fediverseService.loadFollowingData(username)
	if err != nil {
		return model.FollowingPage{}, err
	}

	return buildFollowingPage(&actor, followingURLs, page, pageSize), nil
}

// loadFollowingData 加载用户的关注数据
func (fediverseService *FediverseService) loadFollowingData(
	username string,
) (model.Actor, []string, error) {
	user, err := fediverseService.userRepository.GetUserByUsername(username)
	if err != nil {
		return model.Actor{}, nil, errors.New(commonModel.USER_NOTFOUND)
	}

	actor, _, err := fediverseService.core.BuildActor(&user)
	if err != nil {
		return model.Actor{}, nil, err
	}

	following, err := fediverseService.fediverseRepository.GetFollowing(user.ID)
	if err != nil {
		return model.Actor{}, nil, err
	}

	return actor, uniqueFollowingObjectIDs(following), nil
}

// uniqueFollowingObjectIDs 提取唯一的关注对象 ID 列表,去重并过滤空值
func uniqueFollowingObjectIDs(following []model.Follow) []string {
	if len(following) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(following))
	unique := make([]string, 0, len(following))
	for _, follow := range following {
		if follow.ObjectID == "" {
			continue
		}
		if _, ok := seen[follow.ObjectID]; ok {
			continue
		}
		seen[follow.ObjectID] = struct{}{}
		unique = append(unique, follow.ObjectID)
	}
	return unique
}
