package service

import "context"

// MediaInfo 媒体信息，用于布局推荐
type MediaInfo struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	MediaType string `json:"media_type"` // image/video
}

// ContentInfo 内容信息，用于综合分析
type ContentInfo struct {
	// 文本基础信息
	ContentLength int    `json:"content_length"` // 文本长度（字符数）
	Content       string `json:"content"`        // 文本内容（用于深度分析）
	// 内容结构分析
	HasCode         bool `json:"has_code"`           // 是否包含代码块
	HasLinks        bool `json:"has_links"`          // 是否包含链接
	HasImagesInText bool `json:"has_images_in_text"` // Markdown中是否有图片引用
	HasHeaders      bool `json:"has_headers"`        // 是否有标题（#）
	HasLists        bool `json:"has_lists"`          // 是否有列表
	HasQuotes       bool `json:"has_quotes"`         // 是否有引用块
	LineCount       int  `json:"line_count"`         // 行数
	ParagraphCount  int  `json:"paragraph_count"`    // 段落数
	// 标签信息
	Tags []string `json:"tags"` // 标签列表
}

// LayoutRecommendRequest 布局推荐请求
type LayoutRecommendRequest struct {
	MediaList   []MediaInfo  `json:"media_list"`
	ContentInfo *ContentInfo `json:"content_info"` // 可选的内容信息
}

// LayoutRecommendResponse 布局推荐响应
type LayoutRecommendResponse struct {
	Layout string `json:"layout"` // 推荐的布局：waterfall/grid/horizontal/carousel
	Source string `json:"source"` // 推荐来源：ai/rule
	Reason string `json:"reason"` // 推荐理由
}

type AgentServiceInterface interface {
	// 定义 Agent 服务接口方法
	GetRecent(ctx context.Context) (string, error)
	// 推荐媒体布局
	RecommendLayout(ctx context.Context, req LayoutRecommendRequest) (*LayoutRecommendResponse, error)
}
