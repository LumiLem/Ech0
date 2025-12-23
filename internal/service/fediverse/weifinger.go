package service

import (
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
)

// Webfinger 处理 Webfinger 请求
func (fediverseService *FediverseService) Webfinger(
	username string,
) (model.WebFingerResponse, error) {
	// 查询用户
	user, err := fediverseService.userRepository.GetUserByUsername(username)
	if err != nil {
		return model.WebFingerResponse{}, errors.New(commonModel.USER_NOTFOUND)
	}

	// 构建 Actor 对象
	actor, setting, err := fediverseService.core.BuildActor(&user)
	if err != nil {
		return model.WebFingerResponse{}, err
	}

	return model.WebFingerResponse{
		Subject: "acct:" + user.Username + "@" + httpUtil.ExtractDomain(
			httpUtil.TrimURL(setting.ServerURL),
		),
		Aliases: []string{
			actor.ID,
			"acct:" + user.Username + "@" + httpUtil.ExtractDomain(
				httpUtil.TrimURL(setting.ServerURL),
			),
		},
		Links: []model.Link{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: actor.ID,
			},
		},
	}, nil
}
