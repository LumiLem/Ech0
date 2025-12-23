package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/lin-snow/ech0/internal/fediverse"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
)

// HandleOutbox 处理 Outbox 消息
func (fediverseService *FediverseService) HandleOutboxPage(
	ctx context.Context,
	username string,
	page, pageSize int,
) (model.OutboxPage, error) {
	// 查询用户，确保用户存在
	user, err := fediverseService.userRepository.GetUserByUsername(username)
	if err != nil {
		return model.OutboxPage{}, errors.New(commonModel.USER_NOTFOUND)
	}

	// 获取 Actor和 setting
	actor, setting, err := fediverseService.core.BuildActor(&user)
	if err != nil {
		return model.OutboxPage{}, err
	}
	serverURL, err := fediverse.NormalizeServerURL(setting.ServerURL)
	if err != nil {
		return model.OutboxPage{}, err
	}

	// 查 Echos
	echosByPage, total := fediverseService.echoRepository.GetEchosByPage(page, pageSize, "", false)

	// 转 Avtivity
	var activities []model.Activity
	for i := range echosByPage {
		// 转换为 Activity
		activities = append(
			activities,
			fediverseService.core.ConvertEchoToActivity(&echosByPage[i], &actor, serverURL),
		)
	}

	// 拼装 OutboxPage
	outboxPage := model.OutboxPage{
		Context:      "https://www.w3.org/ns/activitystreams",
		ID:           fmt.Sprintf("%s/users/%s/outbox?page=%d", serverURL, username, page),
		Type:         "OrderedCollectionPage",
		PartOf:       fmt.Sprintf("%s/users/%s/outbox", serverURL, username),
		Next:         "",
		Prev:         "",
		OrderedItems: activities,
	}

	// 计算 Next && Prev
	if page > 1 {
		outboxPage.Prev = fmt.Sprintf("%s/users/%s/outbox?page=%d", serverURL, username, page-1)
	}
	if (page * pageSize) < int(total) {
		outboxPage.Next = fmt.Sprintf("%s/users/%s/outbox?page=%d", serverURL, username, page+1)
	}

	return outboxPage, nil
}

func (fediverseService *FediverseService) HandleOutbox(
	username string,
) (model.OutboxResponse, error) {
	outbox, err := fediverseService.core.BuildOutbox(username)
	if err != nil {
		return model.OutboxResponse{}, err
	}
	return outbox, nil
}
