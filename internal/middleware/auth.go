package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

// JWTAuthMiddleware JWT 拦截器中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 Authorization 头部信息
		auth := ctx.Request.Header.Get("Authorization")

		// 将 Authorization 头部信息分割成两部分
		parts := strings.SplitN(auth, " ", 2)

		// 如果 Authorization 头部信息为空，或者格式不正确，或者 token 为空，则返回错误
		if auth == "" || len(parts) != 2 || len(parts[1]) == 0 || parts[1] == "null" ||
			parts[1] == "undefined" {
			// 如果只是分页获取首页Echo，则不需要鉴权
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/echo/page") {
				// 设置 userid 为 NO_USER_LOGINED
				ctx.Set("userid", authModel.NO_USER_LOGINED)
				ctx.Next()
				return
			}

			// 获取当日Echo
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/echo/today") {
				// 设置 userid 为 NO_USER_LOGINED
				ctx.Set("userid", authModel.NO_USER_LOGINED)
				ctx.Next()
				return
			}

			// 查看Echo详情
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/echo") &&
				ctx.Request.Method == http.MethodGet {
				// 设置 userid 为 NO_USER_LOGINED
				ctx.Set("userid", authModel.NO_USER_LOGINED)
				ctx.Next()
				return
			}

			// 获取 S3 存储设置
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/s3/settings") &&
				ctx.Request.Method == http.MethodGet {
				// 设置 userid 为 NO_USER_LOGINED
				ctx.Set("userid", authModel.NO_USER_LOGINED)
				ctx.Next()
				return
			}

			// 根据 Tag ID 获取 Echo 列表
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/echo/tag/") &&
				ctx.Request.Method == http.MethodGet {
				// 设置 userid 为 NO_USER_LOGINED
				ctx.Set("userid", authModel.NO_USER_LOGINED)
				ctx.Next()
				return
			}

			// 如果 Authorization 头部信息为空，或者格式不正确，或者 token 为空，则返回错误
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.TOKEN_NOT_FOUND,
					Err: nil,
				})),
			)
			ctx.Abort()
			return
		}

		// 如果 Authorization 头部信息格式不正确，或者 token 格式不正确，则返回错误
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.TOKEN_NOT_VALID,
					Err: nil,
				})),
			)
			ctx.Abort()
			return
		}

		// 解析 token
		mc, err := jwtUtil.ParseToken(parts[1])
		if err != nil {
			// 如果 token 解析失败，则返回错误
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.TOKEN_PARSE_ERROR,
					Err: err,
				})),
			)
			ctx.Abort()
			return
		}

		// 如果 token 解析成功，则将用户 ID 存入上下文
		ctx.Set("userid", mc.Userid)
		ctx.Next()
	}
}
