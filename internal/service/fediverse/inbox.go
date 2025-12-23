package service

import (
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ProcessInbox 处理接收到的 ActivityPub 消息
func (fediverseService *FediverseService) HandleInbox(
	username string,
	activity *model.Activity,
) error {
	// 查询用户，确保用户存在
	user, err := fediverseService.userRepository.GetUserByUsername(username)
	if err != nil {
		return errors.New(commonModel.USER_NOTFOUND)
	}

	// 处理不同类型的 Activity
	switch activity.Type {
	// 处理关注请求
	case model.ActivityTypeFollow:
		if err := fediverseService.handleFollowActivity(&user, activity); err != nil {
			return err
		}
	// 处理接收到的推文推送
	// case model.ActivityTypeCreate:
	// 	if err := fediverseService.handleCreateActivity(&user, activity); err != nil {
	// 		return err
	// 	}
	// 处理接收到的接受关注请求
	// case model.ActivityTypeAccept:
	// 	if err := fediverseService.handleAcceptActivity(&user, activity); err != nil {
	// 		return err
	// 	}

	default:
		return errors.New(
			"Unsupported activity type: " + cases.Title(language.English).String(activity.Type),
		)
	}

	return nil
}
