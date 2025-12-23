package util

import (
	model "github.com/lin-snow/ech0/internal/model/common"
	util "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// HandleError 处理错误信息，记录日志并返回错误消息
func HandleError(se *model.ServerError) string {
	if se.Err != nil {
		if se.Msg == "" || len(se.Msg) == 0 {
			se.Msg = se.Err.Error()
		}
		util.GetLogger().Error(se.Msg, zap.Error(se.Err))
	}

	return se.Msg
}

// HandlePanicError 处理 panic 错误，记录日志并触发 panic
func HandlePanicError(se *model.ServerError) {
	if se.Err != nil {
		if se.Msg == "" || len(se.Msg) == 0 {
			se.Msg = se.Err.Error()
		}
		util.GetLogger().Panic(se.Msg, zap.Error(se.Err))
	}

	panic(se.Msg)
}
