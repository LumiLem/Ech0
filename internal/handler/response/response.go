package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
)

// Response 代表 handler 层的执行结果封装
//
// swagger:model Response
type Response struct {
	// Code 状态码，非0时表示自定义HTTP业务状态码
	Code int `json:"code"`

	// Data 响应数据，具体内容因接口而异
	Data any `json:"data,omitempty"`

	// Msg 返回信息，通常是状态描述
	Msg string `json:"msg"`

	// Err 错误信息，序列化时忽略（仅供内部日志使用）
	// swagger:ignore
	Err error `json:"-"`
}

// Execute 包装器，自动根据 Response 返回统一格式的 HTTP 响应 (仅处理返回类型为JSON的handler)
func Execute(fn func(ctx *gin.Context) Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res := fn(ctx)
		if res.Err != nil {
			ctx.JSON(http.StatusBadRequest, commonModel.Fail[string](
				errorUtil.HandleError(&commonModel.ServerError{
					Msg: res.Msg,
					Err: res.Err,
				}),
			))
			return
		}

		// 支持自定义 code
		if res.Code != 0 {
			ctx.JSON(http.StatusOK, commonModel.OKWithCode(res.Data, res.Code, res.Msg))
		} else {
			ctx.JSON(http.StatusOK, commonModel.OK(res.Data, res.Msg))
		}
	}
}
