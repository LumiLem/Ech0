package fediverse

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
)

//==============================================================================
// Build
//==============================================================================

// BuildActor 构建 Actor 对象
func (core *FediverseCore) BuildActor(
	user *userModel.User,
) (model.Actor, *settingModel.SystemSetting, error) {
	// 从设置服务获取服务器域名
	var setting settingModel.SystemSetting
	settingStr, err := core.keyvalueRepo.GetKeyValue(commonModel.SystemSettingsKey)
	if err != nil {
		return model.Actor{}, nil, err
	}
	if err := jsonUtil.JSONUnmarshal([]byte(settingStr.(string)), &setting); err != nil {
		return model.Actor{}, nil, err
	}

	serverURL, err := NormalizeServerURL(setting.ServerURL)
	if err != nil {
		return model.Actor{}, nil, err
	}
	// 构建头像信息 (域名 + /api + 头像路径)
	if user.Avatar == "" {
		user.Avatar = "/Ech0.png" // 默认头像路径
	} else {
		user.Avatar = "/api" + user.Avatar
	}
	avatarURL := serverURL + user.Avatar
	avatarMIME := httpUtil.GetMIMETypeFromFilenameOrURL(avatarURL)

	// 构建 Actor 对象
	return model.Actor{
		Context: []any{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		ID:                serverURL + "/users/" + user.Username, // 实例地址拼接 域名 + /users/ + username
		Type:              "Person",                              // 固定值
		Name:              setting.ServerName,                    // 显示名称
		PreferredUsername: user.Username,                         // 用户名
		Summary:           "你好呀!👋 我是来自Ech0的" + user.Username,     // 简介
		Icon: model.Preview{
			Type:      "Image",
			MediaType: avatarMIME,
			URL:       avatarURL,
		},
		Image: model.Preview{
			Type:      "Image",
			MediaType: "image/png",
			URL:       serverURL + "/banner.png", // 封面图片，固定为 /banner.png
		},
		Followers: serverURL + "/users/" + user.Username + "/followers", // 粉丝列表地址
		Following: serverURL + "/users/" + user.Username + "/following", // 关注列表地址
		Inbox:     serverURL + "/users/" + user.Username + "/inbox",     // 收件箱地址
		Outbox:    serverURL + "/users/" + user.Username + "/outbox",    // 发件箱地址
		PublicKey: model.PublicKey{
			ID:           serverURL + "/users/" + user.Username + "#main-key",
			Owner:        serverURL + "/users/" + user.Username,
			PublicKeyPem: string(config.RSA_PUBLIC_KEY),
			Type:         "Key",
		},
	}, &setting, nil
}

// BuildOutbox 构建 Outbox 元信息
func (core *FediverseCore) BuildOutbox(username string) (model.OutboxResponse, error) {
	// 查询用户，确保用户存在
	user, err := core.userRepository.GetUserByUsername(username)
	if err != nil {
		return model.OutboxResponse{}, errors.New(commonModel.USER_NOTFOUND)
	}

	// 获取 Actor和 setting
	actor, setting, err := core.BuildActor(&user)
	if err != nil {
		return model.OutboxResponse{}, err
	}

	serverURL, err := NormalizeServerURL(setting.ServerURL)
	if err != nil {
		return model.OutboxResponse{}, err
	}

	// 查 Echos
	_, total := core.echoRepository.GetEchosByPage(1, 10, "", false)

	firstPage := fmt.Sprintf("%s?page=1", actor.Outbox)
	lastPage := ""
	if total > 0 {
		totalPages := int(total) / 10
		if total%10 != 0 {
			totalPages++
		}
		lastPage = fmt.Sprintf("%s?page=%d", actor.Outbox, totalPages)
	}

	return model.OutboxResponse{
		Context:    "https://www.w3.org/ns/activitystreams",
		ID:         fmt.Sprintf("%s/users/%s/outbox", serverURL, username),
		Type:       "OrderedCollection",
		TotalItems: int(total),
		First:      firstPage,
		Last:       lastPage,
	}, nil
}

// BuildAcceptActivityPayload 构建 Accept Activity 的 JSON Payload
func (core *FediverseCore) BuildAcceptActivityPayload(
	actor *model.Actor,
	follow *model.Activity,
	followerActor, serverURL string,
) ([]byte, error) {
	if follow.ActivityID == "" {
		return nil, errors.New("follow activity missing id")
	}

	target := follow.ObjectID
	if target == "" {
		target = actor.ID
	}

	now := time.Now().UTC()
	acceptID := fmt.Sprintf(
		"%s/activities/%s/accept/%d",
		serverURL,
		actor.PreferredUsername,
		now.UnixNano(),
	)

	payload := map[string]any{
		"@context": []any{"https://www.w3.org/ns/activitystreams"},
		"id":       acceptID,
		"type":     model.ActivityTypeAccept,
		"actor":    actor.ID,
		"object": map[string]any{
			"id":     follow.ActivityID,
			"type":   model.ActivityTypeFollow,
			"actor":  followerActor,
			"object": target,
		},
		"to":        []string{followerActor},
		"published": now.Format(time.RFC3339),
	}

	return json.Marshal(payload)
}

// BuildFollowActivityPayload 构建 Follow Activity 的 JSON Payload
func BuildFollowActivityPayload(
	actor *model.Actor,
	targetActor string,
	activityID string,
	published time.Time,
) ([]byte, error) {
	if actor == nil {
		return nil, errors.New("actor is nil")
	}
	if activityID == "" {
		return nil, errors.New("activity id is empty")
	}
	if targetActor == "" {
		return nil, errors.New("target actor is empty")
	}

	payload := map[string]any{
		"@context":  []any{"https://www.w3.org/ns/activitystreams"},
		"id":        activityID,
		"type":      model.ActivityTypeFollow,
		"actor":     actor.ID,
		"object":    targetActor,
		"to":        []string{targetActor},
		"published": published.Format(time.RFC3339),
	}

	return json.Marshal(payload)
}

// BuildUndoFollowActivityPayload 构建 Undo Follow Activity 的 JSON Payload
func BuildUndoFollowActivityPayload(
	actor *model.Actor,
	targetActor string,
	undoID string,
	followActivityID string,
	published time.Time,
) ([]byte, error) {
	if actor == nil {
		return nil, errors.New("actor is nil")
	}
	if undoID == "" || followActivityID == "" {
		return nil, errors.New("activity id is empty")
	}
	if targetActor == "" {
		return nil, errors.New("target actor is empty")
	}

	payload := map[string]any{
		"@context": []any{"https://www.w3.org/ns/activitystreams"},
		"id":       undoID,
		"type":     model.ActivityTypeUndo,
		"actor":    actor.ID,
		"object": map[string]any{
			"id":     followActivityID,
			"type":   model.ActivityTypeFollow,
			"actor":  actor.ID,
			"object": targetActor,
		},
		"to":        []string{targetActor},
		"published": published.Format(time.RFC3339),
	}

	return json.Marshal(payload)
}

// BuildLikeActivityPayload 构建 Like Activity 的 JSON Payload
func BuildLikeActivityPayload(
	actor *model.Actor,
	targetActor string,
	object string,
	activityID string,
	published time.Time,
) ([]byte, error) {
	if actor == nil {
		return nil, errors.New("actor is nil")
	}
	if activityID == "" {
		return nil, errors.New("activity id is empty")
	}
	if targetActor == "" || object == "" {
		return nil, errors.New("target actor or object is empty")
	}

	payload := map[string]any{
		"@context":  []any{"https://www.w3.org/ns/activitystreams"},
		"id":        activityID,
		"type":      model.ActivityTypeLike,
		"actor":     actor.ID,
		"object":    object,
		"to":        []string{targetActor},
		"published": published.Format(time.RFC3339),
	}

	return json.Marshal(payload)
}

// BuildUndoLikeActivityPayload 构建 Undo Like Activity 的 JSON Payload
func BuildUndoLikeActivityPayload(
	actor *model.Actor,
	targetActor string,
	object string,
	likeActivityID string,
	undoID string,
	published time.Time,
) ([]byte, error) {
	if actor == nil {
		return nil, errors.New("actor is nil")
	}
	if likeActivityID == "" || undoID == "" {
		return nil, errors.New("activity id is empty")
	}
	if targetActor == "" || object == "" {
		return nil, errors.New("target actor or object is empty")
	}

	payload := map[string]any{
		"@context": []any{"https://www.w3.org/ns/activitystreams"},
		"id":       undoID,
		"type":     model.ActivityTypeUndo,
		"actor":    actor.ID,
		"object": map[string]any{
			"id":     likeActivityID,
			"type":   model.ActivityTypeLike,
			"actor":  actor.ID,
			"object": object,
		},
		"to":        []string{targetActor},
		"published": published.Format(time.RFC3339),
	}

	return json.Marshal(payload)
}
