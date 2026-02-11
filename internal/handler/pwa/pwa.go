package handler

import (
	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	pwaModel "github.com/lin-snow/ech0/internal/model/pwa"
	service "github.com/lin-snow/ech0/internal/service/pwa"
)

type PwaHandler struct {
	pwaService service.PwaServiceInterface
}

func NewPwaHandler(pwaService service.PwaServiceInterface) *PwaHandler {
	return &PwaHandler{pwaService: pwaService}
}

// GetVapidPublicKey 获取 VAPID 公钥
func (h *PwaHandler) GetVapidPublicKey() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		pub, err := h.pwaService.GetVapidPublicKey(ctx)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: pub}
	})
}

// Subscribe 订阅 Web Push
func (h *PwaHandler) Subscribe() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)

		var sub pwaModel.PushSubscription
		if err := ctx.ShouldBindJSON(&sub); err != nil {
			return res.Response{Err: err}
		}

		sub.UserAgent = ctx.GetHeader("User-Agent")
		if err := h.pwaService.Subscribe(ctx, userid, &sub); err != nil {
			return res.Response{Err: err}
		}

		return res.Response{Msg: "订阅成功"}
	})
}

// Unsubscribe 取消订阅
func (h *PwaHandler) Unsubscribe() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var body struct {
			Endpoint string `json:"endpoint"`
		}
		if err := ctx.ShouldBindJSON(&body); err != nil {
			return res.Response{Err: err}
		}

		if err := h.pwaService.Unsubscribe(ctx, body.Endpoint); err != nil {
			return res.Response{Err: err}
		}

		return res.Response{Msg: "取消订阅成功"}
	})
}

// GetSnapshot 获取推送快照
func (h *PwaHandler) GetSnapshot() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)
		snapshot, err := h.pwaService.GetSnapshot(ctx, userid)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: snapshot}
	})
}

// UpdateSnapshot 更新推送快照
func (h *PwaHandler) UpdateSnapshot() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)
		var snapshot pwaModel.PwaPushSnapshot
		if err := ctx.ShouldBindJSON(&snapshot); err != nil {
			return res.Response{Err: err}
		}
		if err := h.pwaService.UpdateSnapshot(ctx, userid, &snapshot); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: "更新成功"}
	})
}

// GetAggregatedStatus 获取聚合后的状态
func (h *PwaHandler) GetAggregatedStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)
		status, err := h.pwaService.GetAggregatedStatus(ctx, userid)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: status}
	})
}
