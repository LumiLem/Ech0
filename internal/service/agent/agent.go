package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/lin-snow/ech0/internal/agent"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

type AgentService struct {
	settingService settingService.SettingServiceInterface
	echoService    echoService.EchoServiceInterface
	todoService    todoService.TodoServiceInterface
	kvRepository   keyvalueRepository.KeyValueRepositoryInterface
	recentGenGroup singleflight.Group
}

func NewAgentService(
	settingService settingService.SettingServiceInterface,
	echoService echoService.EchoServiceInterface,
	todoService todoService.TodoServiceInterface,
	kvRepository keyvalueRepository.KeyValueRepositoryInterface,
) AgentServiceInterface {
	return &AgentService{
		settingService: settingService,
		echoService:    echoService,
		todoService:    todoService,
		kvRepository:   kvRepository,
	}
}

func (agentService *AgentService) GetRecent(ctx context.Context) (string, error) {
	const cacheKey = string(agent.GEN_RECENT)

	if value, ok := agentService.getRecentFromCache(cacheKey); ok {
		return value, nil
	}

	value, err, _ := agentService.recentGenGroup.Do(cacheKey, func() (any, error) {
		if cached, ok := agentService.getRecentFromCache(cacheKey); ok {
			return cached, nil
		}

		output, err := agentService.buildRecentSummary(ctx)
		if err != nil {
			return "", err
		}

		if err := agentService.kvRepository.AddOrUpdateKeyValue(ctx, cacheKey, output); err != nil {
			logUtil.GetLogger().Error("Failed to add or update key value", zap.String("error", err.Error()))
		}

		return output, nil
	})
	if err != nil {
		return "", err
	}

	recent, ok := value.(string)
	if !ok {
		return "", errors.New("recent summary type assertion failed")
	}

	return recent, nil
}

func (agentService *AgentService) getRecentFromCache(cacheKey string) (string, bool) {
	cachedValue, err := agentService.kvRepository.GetKeyValue(cacheKey)
	if err != nil {
		return "", false
	}

	value, ok := cachedValue.(string)
	return value, ok
}

func (agentService *AgentService) buildRecentSummary(ctx context.Context) (string, error) {
	echos, err := agentService.echoService.GetEchosByPage(authModel.NO_USER_LOGINED, commonModel.PageQueryDto{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		return "", err
	}

	var memos []*schema.Message
	for i, e := range echos.Items {
		content := fmt.Sprintf(
			"用户 %s 在 %s 发布了内容 %d ：%s 。 内容标签为：%v。",
			e.Username,
			e.CreatedAt.Format("2006-01-02 15:04"),
			i+1,
			e.Content,
			e.Tags,
		)

		memos = append(memos, &schema.Message{
			Role:    schema.User,
			Content: content,
		})
	}

	in := []*schema.Message{
		{
			Role: schema.System,
			Content: `
				你只能输出纯文本。
				不能输出代码块、格式化标记、Markdown 符号（如井号、星号、反引号、方括号、尖括号）。
				不能输出任何结构化格式（如列表、表格）。
				回复中只能出现正常文字、标点符号和 Emoji 和 换行。
				确保输出始终是自然语言连续文本。`,
		},
		{
			Role:    schema.User,
			Content: "请根据提供的近期互动内容（内容可能包括日常生活、句子诗词摘抄、吐槽等等），总结该用户最近的活动和状态，突出作者状态即可，不需要详细描述内容，如果没有任何内容，请回复作者最近很神秘~",
		},
	}

	in = append(in, memos...)

	var setting model.AgentSetting
	if err := agentService.settingService.GetAgentInfo(&setting); err != nil {
		return "", errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}

	output, err := agent.Generate(ctx, setting, in, true)
	if err != nil {
		return "", err
	}

	return output, nil
}

// RecommendLayout 根据媒体信息推荐最佳布局
func (agentService *AgentService) RecommendLayout(ctx context.Context, req LayoutRecommendRequest) (*LayoutRecommendResponse, error) {
	// 深度分析媒体特征（结合内容信息）
	// 注意：即使 AI 调用失败，规则引擎也需要这个分析结果
	analysis := analyzeMediaFeatures(req.MediaList, req.ContentInfo)

	logUtil.GetLogger().Info("[AI Layout] 分析完成",
		zap.Int("media", analysis.TotalCount),
		zap.Int("content_len", analysis.ContentLength),
		zap.String("type", analysis.ContentType),
		zap.String("hint", analysis.TextPositionHint))

	// 获取 Agent 设置
	var setting model.AgentSetting
	if err := agentService.settingService.GetAgentInfo(&setting); err != nil {
		logUtil.GetLogger().Warn("[AI Layout] 获取 Agent 设置失败，使用规则引擎", zap.Error(err))
		return &LayoutRecommendResponse{
			Layout: analysis.RuleBasedRecommend(),
			Source: "rule",
		}, nil
	}

	// 如果 AI 未启用，直接使用规则引擎
	if !setting.Enable {
		logUtil.GetLogger().Info("[AI Layout] AI 未启用，使用规则引擎")
		return &LayoutRecommendResponse{
			Layout: analysis.RuleBasedRecommend(),
			Source: "rule",
		}, nil
	}

	in := []*schema.Message{
		{
			Role: schema.System,
			Content: `你是一个专业的社交媒体图片布局专家。请根据媒体特征和文本内容，推荐最佳布局。

## 四种布局的实际效果

### 1. grid (九宫格布局)
**显示方式**: 
- 网格排列，图片裁切为正方形（object-cover）
- 预览显示前9张，超过9张时第9张显示"+N"提示
- 点击任意图片可在弹窗中浏览全部
- 单图时智能调整：横图占满3列、方图占2列、竖图占1列
- 根据数量动态调整列数（1张/2张/4张特殊处理，其他3列）
**文字位置**: 文字在图片**上方**
**适合场景**: 
- 需要一览展示多张图片
- 有较长文字说明的内容
- 日常分享、美食、产品展示
- 代码/技术内容（文字在上便于阅读）

### 2. waterfall (瀑布流布局)
**显示方式**: 
- 2列网格，**保持原始宽高比**（不裁切）
- 所有图片都显示，无数量限制
- 错落有致，视觉上更丰富
- 单图时居中展示
- 奇数图片时第1张跨2列居中
**文字位置**: 文字在图片**下方**
**适合场景**: 
- 摄影作品（需保持原始比例）
- 横竖图混合、比例差异大
- 短文字或无文字的图片分享
- 艺术性、设计感强的内容

### 3. horizontal (水平滚动布局)
**显示方式**: 
- 横向滑动浏览，固定高度200px
- 所有图片都显示，通过左右滑动查看
- 沉浸式画廊体验
- 底部有"← 左右滑动查看更多 →"提示
**文字位置**: 文字在图片**上方**
**适合场景**: 
- 横图为主的内容（横图占比>=60%）
- 故事性、时间线叙事
- 旅行风景、全景摄影
- 漫画条漫、步骤展示

### 4. carousel (单图轮播布局)
**显示方式**: 
- 一次展示一张，完整显示
- 有前后导航按钮，显示"当前/总数"
- 所有图片都可以逐张浏览
**文字位置**: 文字在图片**下方**
**适合场景**: 
- 图片数量较多（建议10张以上）
- 教程步骤、产品细节
- 需要逐张仔细查看的内容
- 故事叙述、旅行日记

## 核心决策因素

### 1. 文字位置（根据内容决定）
- **文字在上（grid/horizontal）**: 长文本(>=100字)、代码块、标题、列表、多段落 → 先读文字再看图
- **文字在下（waterfall/carousel）**: 短文本(<100字)或无文字 → 先看图再读文字

### 2. 图片比例
- **需保持原始比例**: 比例差异大、混合横竖图 → waterfall
- **可裁切为正方形**: 比例相近 → grid

### 3. 图片数量
- **10张以上**: carousel（避免一次展示太多，逐张浏览体验更好）
- **1-9张**: 根据其他因素决定

### 4. 横图主导
- **横图占比>=60%且数量>=3**: horizontal（发挥横图优势）

## 决策优先级
1. 数量>=10 → carousel
2. 长文本或代码 → grid（文字在上）
3. 横图主导(>=60%) → horizontal
4. 比例差异大或混合比例 → waterfall
5. 摄影类内容 → waterfall
6. 短文本+少量图片 → waterfall
7. 默认 → grid

## 输出要求
只输出一个单词: grid、waterfall、horizontal 或 carousel`,
		},
		{
			Role:    schema.User,
			Content: analysis.BuildPrompt(),
		},
	}

	output, err := agent.Generate(ctx, setting, in, false, 0.2)
	if err != nil {
		logUtil.GetLogger().Warn("[AI Layout] AI 调用失败，使用规则引擎", zap.Error(err))
		return &LayoutRecommendResponse{
			Layout: analysis.RuleBasedRecommend(),
			Source: "rule",
		}, nil
	}

	// 清理并验证输出
	output = extractLayoutFromOutput(output)
	validLayouts := map[string]bool{
		"waterfall":  true,
		"grid":       true,
		"horizontal": true,
		"carousel":   true,
	}

	source := "ai"
	if !validLayouts[output] {
		logUtil.GetLogger().Warn("[AI Layout] AI 输出无效，使用规则引擎", zap.String("output", output))
		output = analysis.RuleBasedRecommend()
		source = "rule"
	}

	logUtil.GetLogger().Info("[AI Layout] 推荐结果", zap.String("layout", output), zap.String("source", source))
	return &LayoutRecommendResponse{
		Layout: output,
		Source: source,
	}, nil
}

// MediaAnalysis 媒体分析结果
type MediaAnalysis struct {
	TotalCount     int
	LandscapeCount int     // 横图数量 (ratio > 1.2)
	PortraitCount  int     // 竖图数量 (ratio < 0.8)
	SquareCount    int     // 方图数量 (0.8 <= ratio <= 1.2)
	VideoCount     int     // 视频数量
	AvgRatio       float64 // 平均宽高比
	MinRatio       float64 // 最小宽高比
	MaxRatio       float64 // 最大宽高比
	RatioVariance  float64 // 宽高比方差（衡量差异程度）
	DominantType   string  // 主导类型: landscape/portrait/square/mixed
	MediaDetails   []string
	// 内容分析（从 ContentInfo 填充）
	ContentLength   int      // 文本长度
	Content         string   // 文本内容
	HasCode         bool     // 是否包含代码块
	HasLinks        bool     // 是否包含链接
	HasImagesInText bool     // Markdown中是否有图片引用
	HasHeaders      bool     // 是否有标题
	HasLists        bool     // 是否有列表
	HasQuotes       bool     // 是否有引用块
	LineCount       int      // 行数
	ParagraphCount  int      // 段落数
	Tags            []string // 标签列表
	ContentType     string   // 内容类型推断：diary/photography/social/code/article
	// 文字位置建议
	TextPositionHint string // 建议文字位置：top/bottom
}

// analyzeMediaFeatures 深度分析媒体特征
func analyzeMediaFeatures(mediaList []MediaInfo, contentInfo *ContentInfo) *MediaAnalysis {
	analysis := &MediaAnalysis{
		TotalCount: len(mediaList),
		MinRatio:   999,
		MaxRatio:   0,
	}

	if len(mediaList) == 0 {
		return analysis
	}

	// 填充内容信息
	if contentInfo != nil {
		analysis.ContentLength = contentInfo.ContentLength
		analysis.Content = contentInfo.Content
		analysis.HasCode = contentInfo.HasCode
		analysis.HasLinks = contentInfo.HasLinks
		analysis.HasImagesInText = contentInfo.HasImagesInText
		analysis.HasHeaders = contentInfo.HasHeaders
		analysis.HasLists = contentInfo.HasLists
		analysis.HasQuotes = contentInfo.HasQuotes
		analysis.LineCount = contentInfo.LineCount
		analysis.ParagraphCount = contentInfo.ParagraphCount
		analysis.Tags = contentInfo.Tags
		analysis.ContentType = inferContentType(contentInfo)
		analysis.TextPositionHint = inferTextPosition(contentInfo)
	}

	var ratios []float64

	for i, m := range mediaList {
		if m.MediaType == "video" {
			analysis.VideoCount++
		}

		ratio := 1.0 // 默认方图
		shape := "未知"

		if m.Width > 0 && m.Height > 0 {
			ratio = float64(m.Width) / float64(m.Height)
			ratios = append(ratios, ratio)

			if ratio < analysis.MinRatio {
				analysis.MinRatio = ratio
			}
			if ratio > analysis.MaxRatio {
				analysis.MaxRatio = ratio
			}

			if ratio > 1.5 {
				shape = "超宽横图"
				analysis.LandscapeCount++
			} else if ratio > 1.2 {
				shape = "横图"
				analysis.LandscapeCount++
			} else if ratio < 0.67 {
				shape = "超长竖图"
				analysis.PortraitCount++
			} else if ratio < 0.8 {
				shape = "竖图"
				analysis.PortraitCount++
			} else {
				shape = "方图"
				analysis.SquareCount++
			}
		}

		analysis.MediaDetails = append(analysis.MediaDetails,
			fmt.Sprintf("第%d张: %s, %dx%d, 宽高比=%.2f, %s", i+1, m.MediaType, m.Width, m.Height, ratio, shape))
	}

	// 计算平均值和方差
	if len(ratios) > 0 {
		sum := 0.0
		for _, r := range ratios {
			sum += r
		}
		analysis.AvgRatio = sum / float64(len(ratios))

		// 计算方差
		varianceSum := 0.0
		for _, r := range ratios {
			diff := r - analysis.AvgRatio
			varianceSum += diff * diff
		}
		analysis.RatioVariance = varianceSum / float64(len(ratios))
	}

	// 确定主导类型
	analysis.DominantType = determineDominantType(analysis)

	return analysis
}

// inferContentType 根据内容信息推断内容类型
func inferContentType(info *ContentInfo) string {
	if info == nil {
		return "unknown"
	}

	// 检查标签来推断内容类型
	for _, tag := range info.Tags {
		tagLower := strings.ToLower(tag)
		// 摄影相关标签
		if strings.Contains(tagLower, "摄影") || strings.Contains(tagLower, "photo") ||
			strings.Contains(tagLower, "photography") || strings.Contains(tagLower, "风景") ||
			strings.Contains(tagLower, "portrait") || strings.Contains(tagLower, "街拍") {
			return "photography"
		}
		// 日记相关标签
		if strings.Contains(tagLower, "日记") || strings.Contains(tagLower, "diary") ||
			strings.Contains(tagLower, "生活") || strings.Contains(tagLower, "daily") {
			return "diary"
		}
		// 技术相关标签
		if strings.Contains(tagLower, "code") || strings.Contains(tagLower, "编程") ||
			strings.Contains(tagLower, "技术") || strings.Contains(tagLower, "开发") {
			return "code"
		}
	}

	// 根据内容特征推断
	if info.HasCode {
		return "code"
	}

	// 根据文本长度推断
	if info.ContentLength > 500 {
		return "article"
	} else if info.ContentLength > 100 {
		return "diary"
	}

	return "social"
}

// determineDominantType 确定主导类型
func determineDominantType(a *MediaAnalysis) string {
	if a.TotalCount == 0 {
		return "unknown"
	}

	landscapeRatio := float64(a.LandscapeCount) / float64(a.TotalCount)
	portraitRatio := float64(a.PortraitCount) / float64(a.TotalCount)
	squareRatio := float64(a.SquareCount) / float64(a.TotalCount)

	if landscapeRatio >= 0.6 {
		return "landscape_dominant"
	}
	if portraitRatio >= 0.6 {
		return "portrait_dominant"
	}
	if squareRatio >= 0.6 {
		return "square_dominant"
	}
	if a.LandscapeCount > 0 && a.PortraitCount > 0 {
		return "mixed_orientation"
	}
	return "balanced"
}

// inferTextPosition 根据内容特征推断文字应该在图片的位置
// 返回 "top"（文字在上，适合 grid/horizontal）或 "bottom"（文字在下，适合 waterfall/carousel）
func inferTextPosition(info *ContentInfo) string {
	if info == nil {
		return "bottom"
	}

	// 长文本（>= 100 字符）应该在上方，便于先阅读文字再看图
	if info.ContentLength >= 100 {
		return "top"
	}

	// 有代码块的内容，文字应该在上方
	if info.HasCode {
		return "top"
	}

	// 多段落内容（>= 2 段），文字应该在上方
	if info.ParagraphCount >= 2 {
		return "top"
	}

	// 有标题的内容，文字应该在上方
	if info.HasHeaders {
		return "top"
	}

	// 有列表的内容，文字应该在上方
	if info.HasLists {
		return "top"
	}

	// 短文本或无文本，图片先行，文字在下方
	return "bottom"
}

// BuildPrompt 构建给 AI 的提示
func (a *MediaAnalysis) BuildPrompt() string {
	landscapePercent := 0.0
	portraitPercent := 0.0
	squarePercent := 0.0
	if a.TotalCount > 0 {
		landscapePercent = float64(a.LandscapeCount) / float64(a.TotalCount) * 100
		portraitPercent = float64(a.PortraitCount) / float64(a.TotalCount) * 100
		squarePercent = float64(a.SquareCount) / float64(a.TotalCount) * 100
	}

	// 构建内容信息部分
	contentSection := ""
	if a.ContentLength > 0 || len(a.Tags) > 0 {
		tagsStr := "无"
		if len(a.Tags) > 0 {
			tagsStr = strings.Join(a.Tags, ", ")
		}
		contentSection = fmt.Sprintf(`
## 内容信息
- 文本长度: %d 字符
- 行数: %d
- 段落数: %d
- 标签: %s
- 包含代码块: %v
- 包含链接: %v
- 包含标题: %v
- 包含列表: %v
- 包含引用: %v
- 内容类型推断: %s
- 建议文字位置: %s
`, a.ContentLength, a.LineCount, a.ParagraphCount, tagsStr,
			a.HasCode, a.HasLinks, a.HasHeaders, a.HasLists, a.HasQuotes,
			a.ContentType, a.TextPositionHint)
	}

	// 构建文本内容摘要（如果有的话）
	contentPreview := ""
	if a.Content != "" {
		preview := a.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		contentPreview = fmt.Sprintf(`
## 文本内容摘要
%s
`, preview)
	}

	prompt := fmt.Sprintf(`请为以下媒体内容推荐最佳布局：

## 媒体统计
- 总数量: %d 张/个
- 横图: %d 张 (%.0f%%)
- 竖图: %d 张 (%.0f%%)
- 方图: %d 张 (%.0f%%)
- 视频: %d 个
- 主导类型: %s

## 宽高比分析
- 平均宽高比: %.2f
- 最小宽高比: %.2f
- 最大宽高比: %.2f
- 宽高比方差: %.3f (方差越大说明图片比例差异越大)
%s%s
## 媒体详情
%s

请根据以上信息，结合四种布局的特点（grid和horizontal文字在上，waterfall和carousel文字在下），
综合考虑文本长度、内容结构和图片特征，选择最合适的布局。`,
		a.TotalCount,
		a.LandscapeCount, landscapePercent,
		a.PortraitCount, portraitPercent,
		a.SquareCount, squarePercent,
		a.VideoCount,
		a.DominantType,
		a.AvgRatio, a.MinRatio, a.MaxRatio, a.RatioVariance,
		contentSection,
		contentPreview,
		strings.Join(a.MediaDetails, "\n"))

	return prompt
}

// RuleBasedRecommend 基于规则的推荐（作为 AI 的兜底）
// 核心逻辑：
// - grid/horizontal: 文字在图片上方，适合长文本、代码、结构化内容
// - waterfall/carousel: 文字在图片下方，适合短文本、摄影、图片为主的内容
func (a *MediaAnalysis) RuleBasedRecommend() string {
	// 规则1: 数量 >= 10 → carousel（避免信息过载，逐张浏览）
	if a.TotalCount >= 10 {
		return "carousel"
	}

	// 规则2: 基于内容决定文字位置
	// 长文本、代码、结构化内容 → 文字在上（grid/horizontal）
	needsTextOnTop := a.ContentLength >= 100 || a.HasCode || a.HasHeaders || a.HasLists || a.ParagraphCount >= 2

	// 规则3: 横图主导 (>= 60%) 且数量 >= 3
	if a.TotalCount >= 3 {
		landscapeRatio := float64(a.LandscapeCount) / float64(a.TotalCount)
		if landscapeRatio >= 0.6 {
			// 横图主导 → horizontal（文字在上，适合故事性内容）
			return "horizontal"
		}
	}

	// 规则4: 需要文字在上方的内容
	if needsTextOnTop {
		// 文字在上 → grid（经典社交媒体风格）
		return "grid"
	}

	// 规则5: 摄影类内容 → waterfall（保持原始比例，文字在下）
	if a.ContentType == "photography" {
		return "waterfall"
	}

	// 规则6: 宽高比差异大 或 横竖混合 → waterfall（保持原始比例）
	if a.RatioVariance > 0.25 || (a.LandscapeCount > 0 && a.PortraitCount > 0) {
		return "waterfall"
	}

	// 规则7: 全是竖图且数量 >= 3 → waterfall（竖图用瀑布流更好看）
	if a.PortraitCount == a.TotalCount && a.TotalCount >= 3 {
		return "waterfall"
	}

	// 规则8: 短文本或无文本 + 少量图片 → waterfall（图片先行，简洁展示）
	if a.ContentLength < 50 && a.TotalCount <= 3 {
		return "waterfall"
	}

	// 默认: grid（经典社交媒体风格，文字在上）
	return "grid"
}

// extractLayoutFromOutput 从 AI 输出中提取布局名称
func extractLayoutFromOutput(output string) string {
	output = strings.ToLower(strings.TrimSpace(output))

	// 直接匹配
	layouts := []string{"waterfall", "grid", "horizontal", "carousel"}
	for _, layout := range layouts {
		if output == layout {
			return layout
		}
	}

	// 包含匹配（处理 AI 可能输出 "grid." 或 "推荐 grid" 等情况）
	for _, layout := range layouts {
		if strings.Contains(output, layout) {
			return layout
		}
	}

	return ""
}
