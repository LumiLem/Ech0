package handler

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	imgUtil "github.com/lin-snow/ech0/internal/util/img"
	"github.com/lin-snow/ech0/template"
)

type WebHandler struct {
	settingService settingService.SettingServiceInterface
	echoService    echoService.EchoServiceInterface
}

// NewWebHandler WebHandler 的构造函数
func NewWebHandler(
	settingService settingService.SettingServiceInterface,
	echoService echoService.EchoServiceInterface,
) *WebHandler {
	return &WebHandler{
		settingService: settingService,
		echoService:    echoService,
	}
}

// Templates 返回一个处理前端编译后文件的 gin.HandlerFunc
func (webHandler *WebHandler) Templates() gin.HandlerFunc {
	// 提取 dist 子目录
	subFS, _ := fs.Sub(template.WebFS, "dist")

	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.Path
		if requestPath == "/" {
			requestPath = "/index.html"
		}

		if strings.Contains(requestPath, "..") {
			ctx.Status(http.StatusForbidden)
			return
		}

		// 处理特殊的动态 Meta 注入请求 (index.html 及其 fallback)
		// 如果是访问 SPA 的页面路由 (没有扩展名或者是 .html)，通常会走向 index.html
		ext := filepath.Ext(requestPath)
		if ext == "" || ext == ".html" {
			webHandler.handleHTMLRequest(ctx, subFS)
			return
		}

		// 处理 PWA Manifest 动态注入
		if strings.HasSuffix(requestPath, "/app.webmanifest") {
			webHandler.handleManifestRequest(ctx, subFS)
			return
		}

		// 正常静态资源处理
		webHandler.handleStaticRequest(ctx, subFS, requestPath)
	}
}

// handleManifestRequest 处理 PWA Manifest 请求并注入动态信息
func (webHandler *WebHandler) handleManifestRequest(ctx *gin.Context, subFS fs.FS) {
	// 读取 app.webmanifest
	content, err := fs.ReadFile(subFS, "app.webmanifest")
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	manifest := string(content)

	// 获取系统设置
	var settings settingModel.SystemSetting
	_ = webHandler.settingService.GetSetting(&settings)

	// 注入站点名称和描述 (PWA 名称优先使用服务名)
	pwaName := settings.ServerName
	if pwaName == "" {
		pwaName = settings.SiteTitle
	}

	if pwaName != "" {
		// 使用正则匹配第一个 "name" 和 "short_name"，避免影响 shortcuts 或参数
		reName := regexp.MustCompile(`"name":\s*"Ech0"`)
		if loc := reName.FindStringIndex(manifest); loc != nil {
			manifest = manifest[:loc[0]] + fmt.Sprintf(`"name": "%s"`, pwaName) + manifest[loc[1]:]
		}

		reShortName := regexp.MustCompile(`"short_name":\s*"Ech0"`)
		if loc := reShortName.FindStringIndex(manifest); loc != nil {
			manifest = manifest[:loc[0]] + fmt.Sprintf(`"short_name": "%s"`, pwaName) + manifest[loc[1]:]
		}
	}
	if settings.SiteDescription != "" {
		// 匹配原有描述字段并替换 (仅替换第一个，避免影响 shortcuts)
		reDesc := regexp.MustCompile(`"description":\s*"[^"]*"`)
		description := truncate(settings.SiteDescription, 200)
		loc := reDesc.FindStringIndex(manifest)
		if loc != nil {
			manifest = manifest[:loc[0]] + fmt.Sprintf(`"description": "%s"`, description) + manifest[loc[1]:]
		}
	}

	// 注入 PWA 图标 (使用动态图标接口)
	if settings.ServerLogo != "" {
		// 💡 提取文件名作为版本号
		v := filepath.Base(settings.ServerLogo)

		// 定义 manifest 要求的各种标准尺寸 (使用 mode=maskable 触发后端 Padding 算法)
		newIcons := fmt.Sprintf(`"icons": [
    {
      "src": "/api/icon?s=32&v=%s",
      "sizes": "32x32",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=64&v=%s",
      "sizes": "64x64",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=96&v=%s",
      "sizes": "96x96",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=128&v=%s",
      "sizes": "128x128",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=144&v=%s",
      "sizes": "144x144",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=180&v=%s",
      "sizes": "180x180",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=192&v=%s",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=192&mode=maskable&v=%s",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "maskable"
    },
    {
      "src": "/api/icon?s=384&v=%s",
      "sizes": "384x384",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=512&v=%s",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/api/icon?s=512&mode=maskable&v=%s",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ]`, v, v, v, v, v, v, v, v, v, v, v)

		// 替换第一个 icons 块 (顶层图标)，避免影响 shortcuts 里的图标
		reIconsBlock := regexp.MustCompile(`(?is)"icons":\s*\[.*?]`)
		loc := reIconsBlock.FindStringIndex(manifest)
		if loc != nil {
			manifest = manifest[:loc[0]] + newIcons + manifest[loc[1]:]
		}
	}

	ctx.Header("Content-Type", "application/manifest+json; charset=utf-8")
	ctx.String(http.StatusOK, manifest)
}

// handleHTMLRequest 处理 HTML 请求并注入动态 Meta
func (webHandler *WebHandler) handleHTMLRequest(ctx *gin.Context, subFS fs.FS) {
	requestPath := ctx.Request.URL.Path

	// 读取 index.html
	htmlContent, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	html := string(htmlContent)

	// 获取系统设置
	var settings settingModel.SystemSetting
	_ = webHandler.settingService.GetSetting(&settings)

	// 默认 Meta 信息 (从系统设置获取)
	title := settings.SiteTitle
	description := settings.SiteDescription
	keywords := settings.SiteKeywords
	author := settings.ServerName
	serverURL := strings.TrimRight(settings.ServerURL, "/")
	image := webHandler.ensureAbsoluteURL(settings.ServerLogo, serverURL)
	url := serverURL + requestPath
	pageType := "website"

	// 如果是 Echo 详情页，抓取动态内容
	echoIDRegex := regexp.MustCompile(`/echo/(\d+)`)
	matches := echoIDRegex.FindStringSubmatch(requestPath)
	if len(matches) > 1 {
		id, _ := strconv.ParseUint(matches[1], 10, 64)
		// 使用匿名用户权限获取公开 Echo
		echo, err := webHandler.echoService.GetEchoById(authModel.NO_USER_LOGINED, uint(id))
		if err == nil && echo != nil {
			author = echo.Username
			// 转换为 2006年1月2日 格式
			dateCN := fmt.Sprintf("%d年%d月%d日", echo.CreatedAt.Year(), int(echo.CreatedAt.Month()), echo.CreatedAt.Day())

			// Title 优化：作者 + 日期 + 站点名 (稳健且符合社媒风格)
			title = fmt.Sprintf("%s发表于%s的动态 - %s", author, dateCN, settings.SiteTitle)

			// Description 优化：正文摘要，若为空则描述媒体
			description = truncate(echo.Content, 200)
			if description == "" {
				mediaCount := len(echo.Media)
				if mediaCount > 0 {
					description = fmt.Sprintf("%s分享了%d个媒体文件", author, mediaCount)
				} else {
					description = fmt.Sprintf("这是来自%s的一条动态，点击查看详情。", author)
				}
			}

			if len(echo.Media) > 0 {
				// 💡 只有确认是图片类型才作为预览图，视频/音频无法直接被社交平台抓取为图片
				for _, m := range echo.Media {
					if m.MediaType == "image" {
						image = webHandler.ensureAbsoluteURL(m.MediaURL, serverURL)
						break
					}
				}
			}
			pageType = "article"
		}
	}

	// 1. 替换标题
	titleTagRegex := regexp.MustCompile(`(?i)<title>.*?</title>`)
	html = titleTagRegex.ReplaceAllString(html, "<title>"+title+"</title>")

	// 2. 注入自定义全局 Meta (如果用户在后台配置了)
	if settings.CustomMeta != "" {
		html = strings.Replace(html, "</head>", settings.CustomMeta+"\n</head>", 1)
	}

	// 3. 替换或注入 SEO Meta 标签
	html = replaceOrInjectMeta(html, "description", description)
	html = replaceOrInjectMeta(html, "keywords", keywords)
	html = replaceOrInjectMeta(html, "author", author)

	// 4. 替换或注入 OpenGraph Meta 标签
	html = replaceOrInjectProperty(html, "og:title", title)
	html = replaceOrInjectProperty(html, "og:description", description)
	html = replaceOrInjectProperty(html, "og:image", image)
	html = replaceOrInjectProperty(html, "og:url", url)
	html = replaceOrInjectProperty(html, "og:type", pageType)

	ogSiteName := settings.ServerName
	if ogSiteName == "" {
		ogSiteName = settings.SiteTitle
	}
	html = replaceOrInjectProperty(html, "og:site_name", ogSiteName)

	// 5. 替换或注入 Canonical URL
	html = replaceOrInjectLink(html, "canonical", url)

	// 5.5 替换或注入 Favicon/Apple Touch Icon
	favicon := "/api/icon?s=32&fmt=ico"
	appleIcon := "/api/icon?s=180"

	if settings.ServerLogo != "" {
		v := filepath.Base(settings.ServerLogo)
		favicon = fmt.Sprintf("/api/icon?s=32&fmt=ico&v=%s", v)
		appleIcon = fmt.Sprintf("/api/icon?s=180&v=%s", v)
	}

	html = replaceOrInjectFavicon(html, favicon)
	html = replaceOrInjectLink(html, "apple-touch-icon", appleIcon)

	// 6. 替换或注入 JSON-LD (Schema.org)
	logoURL := webHandler.ensureAbsoluteURL(settings.ServerLogo, serverURL)
	ldJson := generateLDJson(settings, echoIDRegex.MatchString(requestPath), title, description, image, url, author, logoURL)
	html = replaceOrInjectJSONLD(html, ldJson)

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}

// HandleDynamicIcon 处理动态图标生成请求 (具备规范化的 API 参数体系)
func (webHandler *WebHandler) HandleDynamicIcon(ctx *gin.Context) {
	sizeStr := ctx.DefaultQuery("s", "512")
	mode := ctx.DefaultQuery("mode", "any")
	fmtStr := ctx.DefaultQuery("fmt", "png")
	paddingStr := ctx.DefaultQuery("p", "0")
	bgStr := ctx.Query("bg")
	fillStr := ctx.DefaultQuery("fill", "on")

	size, _ := strconv.Atoi(sizeStr)
	if size <= 0 || size > 1024 {
		size = 512
	}

	padding, _ := strconv.Atoi(paddingStr)
	if padding < 0 || padding > 50 {
		padding = 0
	}
	// 模式预设：如果是 maskable 且没传 padding，默认应用 18% 标准边距
	if mode == "maskable" && padding == 0 {
		padding = 18
	}

	// 🎨 尺寸规范化
	standardSizes := []int{16, 32, 48, 64, 72, 96, 128, 144, 180, 192, 256, 384, 512, 1024}
	finalSize := 1024
	for _, s := range standardSizes {
		if size <= s {
			finalSize = s
			break
		}
	}

	// 获取系统设置中的 Logo
	var settings settingModel.SystemSetting
	_ = webHandler.settingService.GetSetting(&settings)
	if settings.ServerLogo == "" {
		ctx.Status(http.StatusNotFound)
		return
	}

	// 1. 解析源图片路径
	logoPath := settings.ServerLogo
	var fullPath string
	if strings.HasPrefix(logoPath, "/images/") {
		fullPath = filepath.Join("data", "images", strings.TrimPrefix(logoPath, "/images/"))
	} else if strings.HasPrefix(logoPath, "http") {
		ctx.Redirect(http.StatusFound, logoPath)
		return
	} else {
		fullPath = filepath.Join("template", "dist", strings.TrimPrefix(logoPath, "/"))
	}

	sourceStat, err := os.Stat(fullPath)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	// 2. 生成指纹 (包含所有规范化后的参数)
	id := fmt.Sprintf("%s-%d-%d-%d-%s-%s-%d-%s-%s", fullPath, sourceStat.ModTime().Unix(), sourceStat.Size(), finalSize, mode, fmtStr, padding, bgStr, fillStr)
	fingerprint := fmt.Sprintf("%x", md5.Sum([]byte(id)))
	etag := fmt.Sprintf(`W/"%s"`, fingerprint)

	// 3. 浏览器协商缓存
	ctx.Header("ETag", etag)
	ctx.Header("Cache-Control", "no-cache")

	if ctx.GetHeader("If-None-Match") == etag {
		ctx.Status(http.StatusNotModified)
		return
	}

	// 4. 物理缓存路径
	outExt := "png"
	if fmtStr == "ico" {
		outExt = "ico"
	}
	cacheDir := filepath.Join("data", "cache", "icons")
	_ = os.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s.%s", fingerprint, outExt))

	if _, err := os.Stat(cachePath); err == nil {
		if fmtStr == "ico" {
			ctx.Header("Content-Type", "image/x-icon")
		} else {
			ctx.Header("Content-Type", "image/png")
		}
		ctx.File(cachePath)
		return
	}

	// 5. 缓存不命中，调用核心引擎处理
	f, err := os.Open(fullPath)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// 准备处理选项
	opts := imgUtil.IconOptions{
		Size:     finalSize,
		Padding:  padding,
		BgColor:  bgStr,
		AutoFill: fillStr != "off",
		Maskable: mode == "maskable",
		Format:   fmtStr,
	}

	// 调用重构后的核心引擎
	data, contentType, err := imgUtil.ProcessIcon(f, opts)
	if err != nil {
		// 修改点：确保回退跳转的 URL 是带 /api 前缀的可访问路径
		ctx.Redirect(http.StatusFound, webHandler.ensureRelativeURL(logoPath))
		return
	}

	// 6. 原子化写入并输出
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err == nil {
		_ = os.Rename(tmpPath, cachePath)
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Cache-Control", "no-cache")
	_, _ = ctx.Writer.Write(data)
}

// handleStaticRequest 处理静态资源
func (webHandler *WebHandler) handleStaticRequest(ctx *gin.Context, subFS fs.FS, requestPath string) {
	fileServer := http.FS(subFS)
	fullPath := path.Clean("." + requestPath)
	// 为了确保 Service Worker 能够及时更新，针对 sw.js 和 custom-sw.js 禁用服务器缓存
	if requestPath == "/sw.js" || requestPath == "/custom-sw.js" {
		ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	f, err := fileServer.Open(fullPath)
	if err != nil {
		// 如果没找到静态资源，fallback 到 HTML 处理以便 SPA 路由正常工作
		webHandler.handleHTMLRequest(ctx, subFS)
		return
	}
	defer func() { _ = f.Close() }()

	stat, _ := f.Stat()
	encoding := ctx.GetHeader("Accept-Encoding")
	if strings.Contains(encoding, "gzip") {
		gzPath := fullPath + ".gz"
		gzFile, err := fileServer.Open(gzPath)
		if err == nil {
			defer func() { _ = gzFile.Close() }()
			gzStat, _ := gzFile.Stat()
			ctx.Header("Content-Encoding", "gzip")
			ctx.Header("Content-Type", getMimeType(fullPath))
			http.ServeContent(ctx.Writer, ctx.Request, gzPath, gzStat.ModTime(), gzFile)
			return
		}
	}

	ctx.Header("Content-Type", getMimeType(fullPath))
	http.ServeContent(ctx.Writer, ctx.Request, fullPath, stat.ModTime(), f)
}

// getMimeType 根据扩展名获取 MIME
func getMimeType(path string) string {
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}

// replaceOrInjectMeta 替换或注入 <meta name="...">
// replaceOrInjectTag 核心辅助函数：匹配正则则替换，否则注入到 </head> 之前
func replaceOrInjectTag(html, pattern, newTag string) string {
	re := regexp.MustCompile(pattern)
	if re.MatchString(html) {
		return re.ReplaceAllString(html, newTag)
	}
	return strings.Replace(html, "</head>", newTag+"\n</head>", 1)
}

// replaceOrInjectMeta 替换或注入 <meta name="...">
func replaceOrInjectMeta(html, name, content string) string {
	if content == "" {
		return html
	}
	pattern := fmt.Sprintf(`(?is)<meta[^>]*?name=["']%s["'][^>]*?>`, regexp.QuoteMeta(name))
	tag := fmt.Sprintf(`<meta name="%s" content="%s" />`, name, content)
	return replaceOrInjectTag(html, pattern, tag)
}

// replaceOrInjectProperty 替换或注入 <meta property="...">
func replaceOrInjectProperty(html, property, content string) string {
	if content == "" {
		return html
	}
	pattern := fmt.Sprintf(`(?is)<meta[^>]*?property=["']%s["'][^>]*?>`, regexp.QuoteMeta(property))
	tag := fmt.Sprintf(`<meta property="%s" content="%s" />`, property, content)
	return replaceOrInjectTag(html, pattern, tag)
}

// replaceOrInjectLink 替换或注入 <link ... rel="...">
func replaceOrInjectLink(html, rel, href string) string {
	if href == "" {
		return html
	}
	pattern := fmt.Sprintf(`(?is)<link[^>]*?rel=["']%s["'][^>]*?>`, regexp.QuoteMeta(rel))
	tag := fmt.Sprintf(`<link rel="%s" href="%s" />`, rel, href)
	return replaceOrInjectTag(html, pattern, tag)
}

// replaceOrInjectFavicon 替换或注入 <link ... rel="icon" ... >
func replaceOrInjectFavicon(html, faviconURL string) string {
	if faviconURL == "" {
		return html
	}
	// 匹配带有 id="favicon" 的链接标签
	re := regexp.MustCompile(`(?is)<link[^>]*?id=["']favicon["'][^>]*?>`)
	newLink := fmt.Sprintf(`<link rel="icon" href="%s" id="favicon" />`, faviconURL)
	if re.MatchString(html) {
		return re.ReplaceAllString(html, newLink)
	}

	// 如果没有 id="favicon"，退而求其次匹配普通的 rel="icon"
	reRel := regexp.MustCompile(`(?is)<link[^>]*?rel=["'](?:shortcut )?icon["'][^>]*?>`)
	if reRel.MatchString(html) {
		return reRel.ReplaceAllString(html, newLink)
	}

	return strings.Replace(html, "</head>", newLink+"\n</head>", 1)
}

// replaceOrInjectJSONLD 替换或注入 <script type="application/ld+json">
func replaceOrInjectJSONLD(html, content string) string {
	if content == "" {
		return html
	}
	re := regexp.MustCompile(`(?is)<script\s+id=["']ldjson-schema["'][^>]*?>.*?</script>`)
	newScript := fmt.Sprintf(`<script type="application/ld+json" id="ldjson-schema">%s</script>`, content)
	if re.MatchString(html) {
		return re.ReplaceAllString(html, newScript)
	}
	// Fallback to type regex if id not found (for legacy/initial injection)
	reType := regexp.MustCompile(`(?is)<script\s+type=["']application/ld\+json["'][^>]*?>.*?</script>`)
	if reType.MatchString(html) {
		return reType.ReplaceAllString(html, newScript)
	}
	return strings.Replace(html, "</head>", newScript+"\n</head>", 1)
}

// generateLDJson 生成针对不同页面的 JSON-LD
func generateLDJson(settings settingModel.SystemSetting, isEchoPage bool, title, description, image, url, author, logoURL string) string {
	var data map[string]interface{}

	// 统一获取服务/组织名称
	serviceName := settings.ServerName
	if serviceName == "" {
		serviceName = settings.SiteTitle
	}

	if isEchoPage {
		// Echo 详情页使用 Article / BlogPosting
		data = map[string]interface{}{
			"@context":    "https://schema.org",
			"@type":       "BlogPosting",
			"headline":    title,
			"description": description,
			"image":       image,
			"url":         url,
			"author": map[string]interface{}{
				"@type": "Person",
				"name":  author,
			},
			"publisher": map[string]interface{}{
				"@type": "Organization",
				"name":  serviceName,
				"logo": map[string]interface{}{
					"@type": "ImageObject",
					"url":   logoURL,
				},
			},
		}
	} else {
		// 首页使用 WebSite
		data = map[string]interface{}{
			"@context":    "https://schema.org",
			"@type":       "WebSite",
			"name":        serviceName,
			"url":         url,
			"description": settings.SiteDescription,
			"publisher": map[string]interface{}{
				"@type": "Organization",
				"name":  serviceName,
				"logo": map[string]interface{}{
					"@type": "ImageObject",
					"url":   logoURL,
				},
			},
		}
	}

	res, _ := json.MarshalIndent(data, "      ", "  ")
	return "\n      " + string(res) + "\n    "
}

// ensureAbsoluteURL 确保路径为绝对 URL，并处理 /api 前缀逻辑
func (webHandler *WebHandler) ensureAbsoluteURL(urlPath, serverURL string) string {
	if urlPath == "" || strings.HasPrefix(urlPath, "http") {
		return urlPath
	}

	// 如果是本地数据路径 (/images/, /videos/, /audios/)，需要补全 /api
	// 对应 internal/router/router.go 中的 r.Static("api/images", ...)
	if strings.HasPrefix(urlPath, "/images/") ||
		strings.HasPrefix(urlPath, "/videos/") ||
		strings.HasPrefix(urlPath, "/audios/") {
		return serverURL + "/api" + urlPath
	}

	// 特殊情况：/Ech0.svg 回退为 /Ech0.png 因为有些社交平台不识别 svg
	if urlPath == "/Ech0.svg" {
		return serverURL + "/Ech0.png"
	}

	// 否则，直接拼接服务器地址 (针对 template/dist 里的静态资源)
	return serverURL + urlPath
}

// ensureRelativeURL 确保路径为相对根路径 (/api/...)，不带域名
func (webHandler *WebHandler) ensureRelativeURL(urlPath string) string {
	if urlPath == "" || strings.HasPrefix(urlPath, "http") {
		return urlPath
	}

	if strings.HasPrefix(urlPath, "/images/") ||
		strings.HasPrefix(urlPath, "/videos/") ||
		strings.HasPrefix(urlPath, "/audios/") {
		return "/api" + urlPath
	}

	if urlPath == "/Ech0.svg" {
		return "/Ech0.png"
	}

	return urlPath
}

// truncate 截短字符串并简单清理 HTML/MD
func truncate(s string, maxLen int) string {
	// 简单的清理
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")       // 移除 HTML
	s = regexp.MustCompile(`[*#_~`+"`"+`]`).ReplaceAllString(s, "") // 移除 MD
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)

	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
