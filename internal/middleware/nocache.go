package middleware

import (
	"github.com/gin-gonic/gin"
)

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 所有缓存节点（浏览器、CDN、代理）完全不要存储响应内容
		c.Header(
			"Cache-Control",
			"no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0",
		)
		// 针对 HTTP/1.0 兼容
		c.Header("Pragma", "no-cache")
		// 立刻过期
		c.Header("Expires", "0")
		// 针对 Surrogate 缓存（如 CDN）
		c.Header("Surrogate-Control", "no-store")

		c.Next()
	}
}
