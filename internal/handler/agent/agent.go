package handler

import (
	"github.com/gin-gonic/gin"

	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/agent"
)

type AgentHandler struct {
	agentService service.AgentServiceInterface
}

func NewAgentHandler(agentService service.AgentServiceInterface) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

func (agentHandler *AgentHandler) GetRecent() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 调用服务层获取作者近况信息
		gen, err := agentHandler.agentService.GetRecent(ctx)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: gen,
			Msg:  commonModel.AGENT_GET_RECENT_SUCCESS,
		}
	})
}

// RecommendLayout 推荐媒体布局
func (agentHandler *AgentHandler) RecommendLayout() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req service.LayoutRecommendRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{
				Msg: "参数错误",
				Err: err,
			}
		}

		result, err := agentHandler.agentService.RecommendLayout(ctx, req)
		if err != nil {
			return res.Response{
				Msg: "布局推荐失败",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  "布局推荐成功",
		}
	})
}
