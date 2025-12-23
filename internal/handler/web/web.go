package handler

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/template"
)

type WebHandler struct{}

// NewWebHandler WebHandler 的构造函数
func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// Templates 返回一个处理前端编译后文件的 gin.HandlerFunc
func (webHandler *WebHandler) Templates() gin.HandlerFunc {
	// 提取 dist 子目录
	subFS, _ := fs.Sub(template.WebFS, "dist")
	fileServer := http.FS(subFS)

	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.Path
		if requestPath == "/" {
			requestPath = "/index.html"
		}

		if strings.Contains(requestPath, "..") {
			ctx.Status(http.StatusForbidden)
			return
		}

		fullPath := path.Clean("." + requestPath)
		f, err := fileServer.Open(fullPath)
		if err != nil {
			// fallback 到 index.html
			fallback, err := fileServer.Open("index.html")
			if err != nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			defer func() { _ = fallback.Close() }()
			fallbackStat, _ := fallback.Stat()
			ctx.Header("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(
				ctx.Writer,
				ctx.Request,
				"index.html",
				fallbackStat.ModTime(),
				fallback,
			)
			return
		}
		defer func() { _ = f.Close() }()

		// 获取文件信息
		stat, _ := f.Stat()

		// 适配资源压缩Gzip 算法
		encoding := ctx.GetHeader("Accept-Encoding")
		if strings.Contains(encoding, "gzip") {
			gzPath := fullPath + ".gz"
			gzFile, err := fileServer.Open(gzPath)
			if err == nil {
				defer func() { _ = gzFile.Close() }()
				stat, _ := gzFile.Stat()
				ctx.Header("Content-Encoding", "gzip")
				ctx.Header("Content-Type", getMimeType(fullPath))
				http.ServeContent(ctx.Writer, ctx.Request, gzPath, stat.ModTime(), gzFile)
				return
			}
		}

		ctx.Header("Content-Type", getMimeType(fullPath))
		http.ServeContent(ctx.Writer, ctx.Request, fullPath, stat.ModTime(), f)
	}
}

// getMimeType 根据文件扩展名返回 MIME 类型，带默认值
func getMimeType(path string) string {
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}
