package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
	service "github.com/lin-snow/ech0/internal/service/fediverse"
)

type FediverseHandler struct {
	service service.FediverseServiceInterface
}

func NewFediverseHandler(fediverseService service.FediverseServiceInterface) *FediverseHandler {
	return &FediverseHandler{
		service: fediverseService,
	}
}

// Webfinger 处理 Webfinger 请求
func (h *FediverseHandler) Webfinger(ctx *gin.Context) {
	resource := ctx.Query("resource")
	if resource == "" {
		ctx.JSON(http.StatusBadRequest, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   "Missing resource parameter",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// 解析 resource，格式应为 acct:username@domain
	if !strings.HasPrefix(resource, "acct:") {
		ctx.JSON(http.StatusBadRequest, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   "Invalid resource format",
			Status:  http.StatusBadRequest,
		})
		return
	}
	parts := strings.SplitN(resource[5:], "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		ctx.JSON(http.StatusBadRequest, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   "Invalid resource format",
			Status:  http.StatusBadRequest,
		})
		return
	}
	username := parts[0]

	// 调用服务层获取 Actor 信息
	webfingerRes, err := h.service.Webfinger(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusNotFound,
		})
		return
	}

	// 设置 Content-Type 为 application/jrd+json
	ctx.Header("Content-Type", "application/jrd+json")

	// 返回 Actor 信息
	ctx.JSON(http.StatusOK, webfingerRes)
}

// GetActor 获取 Actor 信息
func (h *FediverseHandler) GetActor(ctx *gin.Context) {
	// 从 URL 参数中获取用户名
	username := ctx.Param("username")

	// 调用服务层获取 Actor 信息
	actor, err := h.service.GetActorByUsername(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusNotFound,
		})
		return
	}

	// 设置 Content-Type 为 application/activity+json
	ctx.Header("Content-Type", "application/activity+json")

	// 返回 Actor 信息
	ctx.JSON(http.StatusOK, actor)
}

// PostInbox 处理接收到的 ActivityPub 消息
func (h *FediverseHandler) PostInbox(ctx *gin.Context) {
	// 从 URL 参数中获取用户名
	username := ctx.Param("username")

	var activity model.Activity
	if err := ctx.ShouldBindJSON(&activity); err != nil {
		ctx.JSON(http.StatusBadRequest, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   "Invalid JSON",
			Status:  http.StatusBadRequest,
		})
		return
	}

	if err := h.service.HandleInbox(username, &activity); err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.Status(http.StatusAccepted)
}

// GetOutbox 获取 Outbox 消息
func (h *FediverseHandler) GetOutbox(ctx *gin.Context) {
	// 从 URL 参数中获取用户名
	username := ctx.Param("username")

	var pageStr, pageSizeStr string
	pageStr = ctx.Query("page")
	pageSizeStr = ctx.Query("pageSize")

	// 检查是否带有查询参数
	if pageStr == "" {
		// 不带查询参数，返回 Outbox 元信息
		outbox, err := h.service.HandleOutbox(username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
				Context: "https://www.w3.org/ns/activitystreams",
				Type:    "Error",
				Error:   err.Error(),
				Status:  http.StatusInternalServerError,
			})
			return
		}

		// 设置 Content-Type 为 application/activity+json
		ctx.Header("Content-Type", "application/activity+json")

		// 返回 Outbox 元信息
		ctx.JSON(http.StatusOK, outbox)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	// 校验 page 和 pageSize 的合理范围
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 50 {
		pageSize = 50
	}

	// 打印分页参数
	fmt.Printf("Pagination params - page: %d, pageSize: %d\n", page, pageSize)

	// 调用服务层获取 Outbox 信息
	outbox, err := h.service.HandleOutboxPage(ctx, username, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// 设置 Content-Type 为 application/activity+json
	ctx.Header("Content-Type", "application/activity+json")

	// 返回 Outbox 信息
	ctx.JSON(http.StatusOK, outbox)
}

// GetFollowers 获取粉丝列表
func (h *FediverseHandler) GetFollowers(ctx *gin.Context) {
	// 从 URL 参数中获取用户名
	username := ctx.Param("username")

	pageParam := ctx.Query("page")
	pageSizeParam := ctx.Query("pageSize")

	// 如果没有分页参数或显式请求 false，则返回集合摘要
	if pageParam == "" || strings.EqualFold(pageParam, "false") {
		followers, err := h.service.GetFollowers(username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
				Context: "https://www.w3.org/ns/activitystreams",
				Type:    "Error",
				Error:   err.Error(),
				Status:  http.StatusInternalServerError,
			})
			return
		}

		// 设置 Content-Type 为 application/activity+json
		ctx.Header("Content-Type", "application/activity+json")

		// 返回粉丝列表摘要
		ctx.JSON(http.StatusOK, followers)
		return
	}

	page := 1
	if strings.EqualFold(pageParam, "true") {
		page = 1
	} else {
		parsedPage, err := strconv.Atoi(pageParam)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	pageSize := 20
	if pageSizeParam != "" {
		if parsedSize, err := strconv.Atoi(pageSizeParam); err == nil {
			pageSize = parsedSize
		}
	}

	// 调用服务层获取粉丝列表分页
	followersPage, err := h.service.GetFollowersPage(username, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// 设置 Content-Type 为 application/activity+json
	ctx.Header("Content-Type", "application/activity+json")

	// 返回粉丝列表分页
	ctx.JSON(http.StatusOK, followersPage)
}

// GetFollowing 获取关注列表
func (h *FediverseHandler) GetFollowing(ctx *gin.Context) {
	// 从 URL 参数中获取用户名
	username := ctx.Param("username")

	pageParam := ctx.Query("page")
	pageSizeParam := ctx.Query("pageSize")

	// 如果没有分页参数或显式请求 false，则返回集合摘要
	if pageParam == "" || strings.EqualFold(pageParam, "false") {
		following, err := h.service.GetFollowing(username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
				Context: "https://www.w3.org/ns/activitystreams",
				Type:    "Error",
				Error:   err.Error(),
				Status:  http.StatusInternalServerError,
			})
			return
		}

		ctx.Header("Content-Type", "application/activity+json")
		ctx.JSON(http.StatusOK, following)
		return
	}

	page := 1
	if strings.EqualFold(pageParam, "true") {
		page = 1
	} else {
		if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	pageSize := 20
	if pageSizeParam != "" {
		if parsedSize, err := strconv.Atoi(pageSizeParam); err == nil {
			pageSize = parsedSize
		}
	}

	followingPage, err := h.service.GetFollowingPage(username, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.Header("Content-Type", "application/activity+json")
	ctx.JSON(http.StatusOK, followingPage)
}

// GetObject 获取内容对象
func (h *FediverseHandler) GetObject(ctx *gin.Context) {
	// 从 URL 参数中获取对象 ID
	id := ctx.Param("id")

	// 将 ID 转换为 uint
	uintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   "Invalid object ID",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// 调用服务层获取对象信息
	object, err := h.service.GetObjectByID(uint(uintID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.ActivityPubError{
			Context: "https://www.w3.org/ns/activitystreams",
			Type:    "Error",
			Error:   err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// 设置 Content-Type 为 application/activity+json
	ctx.Header("Content-Type", "application/activity+json")

	// 返回对象信息
	ctx.JSON(http.StatusOK, object)
}

// GetFollowStatus 获取关注状态
// func (h *FediverseHandler) GetFollowStatus(ctx *gin.Context) {
// 	userID := ctx.MustGet("userid").(uint)
// 	targetActor := strings.TrimSpace(ctx.Query("actor"))
// 	if targetActor == "" {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]any](commonModel.FEDIVERSE_INVALID_INPUT))
// 		return
// 	}

// 	followStatus, err := h.service.GetFollowStatus(userID, targetActor)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]any](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(followStatus, commonModel.FEDIVERSE_GET_FOLLOW_STATUS_SUCCESS))
// }

// GetTimeline 获取关注的 Actor 的推文
// func (h *FediverseHandler) GetTimeline(ctx *gin.Context) {
// 	userID := ctx.MustGet("userid").(uint)

// 	page := 1
// 	if pageParam := strings.TrimSpace(ctx.Query("page")); pageParam != "" {
// 		if parsed, err := strconv.Atoi(pageParam); err == nil && parsed > 0 {
// 			page = parsed
// 		}
// 	}

// 	pageSize := 20
// 	if pageSizeParam := strings.TrimSpace(ctx.Query("pageSize")); pageSizeParam != "" {
// 		if parsed, err := strconv.Atoi(pageSizeParam); err == nil && parsed > 0 {
// 			pageSize = parsed
// 		}
// 	}

// 	timeline, err := h.service.GetTimeline(userID, page, pageSize)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[commonModel.PageQueryResult[[]model.TimelineItem]](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(timeline, commonModel.FEDIVERSE_GET_TIMELINE_SUCCESS))
// }

// SearchActorByActorID 根据 Actor URL 搜索远端 Actor
// func (h *FediverseHandler) SearchActorByActorID(ctx *gin.Context) {
// 	actorID := strings.TrimSpace(ctx.Query("actor"))
// 	if actorID == "" {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]any](commonModel.FEDIVERSE_INVALID_INPUT))
// 		return
// 	}

// 	actor, err := h.service.SearchActorByActorID(actorID)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]any](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(actor, commonModel.FEDIVERSE_SEARCH_ACTOR_SUCCESS))
// }

// PostFollow 发起关注请求
// func (h *FediverseHandler) PostFollow(ctx *gin.Context) {
// 	userID := ctx.MustGet("userid").(uint)
// 	var req model.FollowActionRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](commonModel.INVALID_REQUEST_BODY))
// 		return
// 	}

// 	result, err := h.service.FollowActor(userID, req)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(result, commonModel.FEDIVERSE_FOLLOW_SUCCESS))
// }

// PostUnfollow 发起取消关注请求
// func (h *FediverseHandler) PostUnfollow(ctx *gin.Context) {
// 	userID := ctx.MustGet("userid").(uint)
// 	var req model.FollowActionRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](commonModel.INVALID_REQUEST_BODY))
// 		return
// 	}

// 	result, err := h.service.UnfollowActor(userID, req)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(result, commonModel.FEDIVERSE_UNFOLLOW_SUCCESS))
// }

// PostLike 发起点赞请求
// func (h *FediverseHandler) PostLike(ctx *gin.Context) {
// 	userID := ctx.MustGet("userid").(uint)
// 	var req model.LikeActionRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](commonModel.INVALID_REQUEST_BODY))
// 		return
// 	}

// 	result, err := h.service.LikeObject(userID, req)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(result, commonModel.FEDIVERSE_LIKE_SUCCESS))
// }

// PostUndoLike 发起取消点赞请求
// func (h *FediverseHandler) PostUndoLike(ctx *gin.Context) {
// 	userID := ctx.MustGet("userid").(uint)
// 	var req model.LikeActionRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](commonModel.INVALID_REQUEST_BODY))
// 		return
// 	}

// 	result, err := h.service.UndoLikeObject(userID, req)
// 	if err != nil {
// 		ctx.JSON(http.StatusOK, commonModel.Fail[map[string]string](err.Error()))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, commonModel.OK(result, commonModel.FEDIVERSE_UNDO_LIKE_SUCCESS))
// }
