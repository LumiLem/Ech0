package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/lin-snow/ech0/internal/agent"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
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
			logUtil.GetLogger().
				Error("Failed to add or update key value", zap.String("error", err.Error()))
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
	echos, err := agentService.echoService.GetEchosByPage(
		authModel.NO_USER_LOGINED,
		commonModel.PageQueryDto{
			Page:     1,
			PageSize: 10,
		},
	)
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
		layout, reason := analysis.RuleBasedRecommend()
		return &LayoutRecommendResponse{
			Layout: layout,
			Source: "rule",
			Reason: reason,
		}, nil
	}

	// 如果 AI 未启用，直接使用规则引擎
	if !setting.Enable {
		logUtil.GetLogger().Info("[AI Layout] AI 未启用，使用规则引擎")
		layout, reason := analysis.RuleBasedRecommend()
		return &LayoutRecommendResponse{
			Layout: layout,
			Source: "rule",
			Reason: reason,
		}, nil
	}

	in := []*schema.Message{
		{
			Role: schema.System,
			Content: `你是社交媒体布局专家。请**综合评估所有信息**，推荐最佳布局。

## 重要前提
用户可以点击任何图片进入全屏查看，所以布局决定的是**首次展示的体验**。

## 四种布局精确特点

### grid（九宫格）
- **图片**：裁切为方形缩略图，最多显示9张（超出显示+N）
- **文字**：在图片**上方**，读者先读文字再看图
- **单图**：智能调整（横图占满、方图2/3、竖图1/3）
- **适合**：
  - 有重要文字内容需要先阅读（代码、长文、讨论）
  - 图片可以被裁切成方形而不损失重点
  - 快速预览很多图（>9张时显示+N提示）

### waterfall（瀑布流）
- **图片**：保持原始比例完整显示，2列错落有致
- **文字**：在图片**下方**，读者先看图再读文字
- **单图**：居中完整展示；奇数图第1张跨2列
- **适合**：
  - 图片本身是重点（摄影、设计、穿搭、美食）
  - 图片比例不一致（有横有竖）需要保持原貌
  - 短文本或无文本的纯图片分享

### horizontal（水平滚动）
- **图片**：固定高度横向排列，左右滑动浏览
- **文字**：在图片**上方**
- **体验**：沉浸式画廊感，有"← 左右滑动 →"提示
- **适合**：
  - 以横图为主的内容（风景、全景）
  - 有连续性/时间顺序的内容（旅程、过程）
  - 图片之间有叙事关系

### carousel（单图轮播）
- **图片**：一次显示一张完整图片，有前后导航
- **文字**：在图片**下方**
- **体验**：显示"当前/总数"，逐张专注查看
- **适合**：
  - 每张图都需要仔细看（教程步骤、产品多角度）
  - 图片较多（>=10张）避免信息过载
  - 对比展示（前后对比、A/B选择）

## 文本语义分析（最重要）

仔细阅读用户的文字内容，理解其**意图和语气**：

### 用户在"表达观点/分享经验" → grid
- 语义特征：描述性文字、解释性内容、问答讨论
- 关键词：今天学到了、给大家推荐、分享一下、请问、有人知道吗
- 判断：文字是主体，需要先读懂

### 用户在"展示图片/作品" → waterfall
- 语义特征：简短感叹、表情符号、作品名称
- 关键词：拍的、随拍、好美、❤️、看！、今天的、记录
- 判断：图片是主体，文字只是点缀

### 用户在"记录过程/旅程" → horizontal
- 语义特征：时间词、顺序词、地点变化
- 关键词：从...到...、第一天、接着、然后、一路、全景
- 判断：图片有连续性，需要按序浏览

### 用户在"教学/对比" → carousel
- 语义特征：步骤说明、对比描述、选择询问
- 关键词：第一步、如何、教程、vs、对比、哪个好
- 判断：每张图都重要，需要逐一查看

## 综合评分逻辑

对每种布局计算适合度分数，综合考虑所有维度：

| 维度 | 权重 | 具体评分 |
|------|------|----------|
| 文本语义 | 最高 | 根据用户意图判断（表达→grid, 展示→waterfall, 旅程→horizontal, 教学→carousel）|
| 文本特征 | 高 | 代码+35grid, 长文(>=150)+30grid, 短文(<30)+25waterfall |
| 图片比例 | 中 | 全横图(>=90%)+30horizontal, 横竖混合+20waterfall, 比例差异大+15waterfall |
| 图片数量 | 中 | >=15+25carousel, <=2+18waterfall, 单图+20waterfall |
| 内容类型 | 中 | 摄影+30waterfall, 教程+25carousel, 故事+25horizontal |
| 标签关键词 | 低 | 相关标签+12 |

## 关键判断点

1. **理解文字意图**：用户在说什么？想让读者先看什么？
2. **有代码块** → grid +35（必须先读代码）
3. **长文本(>=100字)** → grid +25（文字在上）
4. **短文本(<30字)或emoji** → waterfall +25（图片优先）
5. **横竖比例混合** → waterfall +20（保持各自比例）
6. **全是横图(>=70%)且>=3张** → horizontal +25（画廊体验）
7. **摄影/美食/穿搭语义** → waterfall +30（展示作品）
8. **教程/步骤语义** → carousel +25（逐步查看）

## 输出格式
布局|理由（10字内，说明主要依据）

示例：
- grid|代码分享，先读后看
- waterfall|展示摄影，保持比例
- horizontal|全横图，画廊浏览
- carousel|教程步骤，逐张查看`,
		},
		{
			Role:    schema.User,
			Content: analysis.BuildPrompt(),
		},
	}

	output, err := agent.Generate(ctx, setting, in, false, 0.2)
	if err != nil {
		logUtil.GetLogger().Warn("[AI Layout] AI 调用失败，使用规则引擎", zap.Error(err))
		layout, reason := analysis.RuleBasedRecommend()
		return &LayoutRecommendResponse{
			Layout: layout,
			Source: "rule",
			Reason: reason,
		}, nil
	}

	// 解析 AI 输出（格式：布局|理由）
	layout, reason := parseLayoutOutput(output)
	validLayouts := map[string]bool{
		"waterfall":  true,
		"grid":       true,
		"horizontal": true,
		"carousel":   true,
	}

	source := "ai"
	if !validLayouts[layout] {
		logUtil.GetLogger().Warn("[AI Layout] AI 输出无效，使用规则引擎", zap.String("output", output))
		layout, reason = analysis.RuleBasedRecommend()
		source = "rule"
	}

	logUtil.GetLogger().Info("[AI Layout] 推荐结果", zap.String("layout", layout), zap.String("source", source), zap.String("reason", reason))
	return &LayoutRecommendResponse{
		Layout: layout,
		Source: source,
		Reason: reason,
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
// 返回类型与规则引擎匹配：
// - technical, code: 技术/代码内容
// - photography, art: 摄影/艺术作品
// - tutorial, guide: 教程/指南
// - timeline, story: 故事/时间线
// - discussion, question: 讨论/问题
// - diary: 日记/生活
// - social: 社交分享（默认）
func inferContentType(info *ContentInfo) string {
	if info == nil {
		return "social"
	}

	content := strings.ToLower(info.Content)

	// 1. 检查标签来推断内容类型
	for _, tag := range info.Tags {
		tagLower := strings.ToLower(tag)

		// 摄影/艺术相关
		if containsAny(tagLower, []string{"摄影", "photo", "photography", "风景", "portrait", "街拍", "随拍", "art", "艺术", "设计", "插画"}) {
			return "photography"
		}
		// 教程相关
		if containsAny(tagLower, []string{"教程", "tutorial", "指南", "guide", "步骤", "how"}) {
			return "tutorial"
		}
		// 旅行/故事相关
		if containsAny(tagLower, []string{"旅行", "travel", "游记", "旅途", "故事", "story"}) {
			return "timeline"
		}
		// 技术相关
		if containsAny(tagLower, []string{"code", "编程", "技术", "开发", "代码", "程序"}) {
			return "code"
		}
	}

	// 2. 根据内容特征推断
	if info.HasCode {
		return "code"
	}

	// 3. 根据文本内容语义推断
	// 教程/步骤类
	if containsAny(content, []string{"第一步", "第二步", "步骤", "如何", "教程", "方法"}) {
		return "tutorial"
	}
	// 问题/讨论类
	if containsAny(content, []string{"请问", "有人", "怎么", "为什么", "吗？", "呢？"}) {
		return "discussion"
	}
	// 时间线/故事类
	if containsAny(content, []string{"今天", "昨天", "从...到", "第一天", "一路", "旅途"}) {
		return "timeline"
	}
	// 摄影/展示类（短文本+感叹）
	if info.ContentLength < 30 && containsAny(content, []string{"拍", "好美", "美丽", "漂亮", "❤", "😍", "🌸"}) {
		return "photography"
	}

	// 4. 根据文本长度推断
	if info.ContentLength > 200 {
		// 长文本，看是否有结构
		if info.HasHeaders || info.HasLists {
			return "tutorial" // 有结构的长文可能是教程
		}
		return "diary" // 普通长文当日记
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
	// 构建标签字符串
	tagsStr := "无"
	if len(a.Tags) > 0 {
		tagsStr = strings.Join(a.Tags, ", ")
	}

	// 构建文本内容摘要（放在最前面，最重要）
	contentPreview := ""
	if a.Content != "" {
		preview := a.Content
		if len(preview) > 300 {
			preview = preview[:300] + "..."
		}
		contentPreview = fmt.Sprintf(`## 用户发帖内容（最重要的判断依据）

%s

`, preview)
	} else {
		contentPreview = `## 用户发帖内容

（无文字内容，仅图片分享）

`
	}

	// 计算横图占比
	landscapeRatio := 0.0
	if a.TotalCount > 0 {
		landscapeRatio = float64(a.LandscapeCount) / float64(a.TotalCount) * 100
	}

	// 判断比例混合情况
	ratioMix := "比例一致"
	if a.LandscapeCount > 0 && a.PortraitCount > 0 {
		ratioMix = "横竖混合"
	} else if a.RatioVariance > 0.25 {
		ratioMix = "比例差异大"
	}

	prompt := fmt.Sprintf(`%s## 内容特征分析
- 文本长度: %d 字符
- 行数/段落数: %d/%d
- 标签: %s
- 包含代码块: %v
- 包含标题/列表/引用: %v/%v/%v
- 推断的内容类型: %s

## 媒体信息
- 图片数量: %d 张
- 横图/竖图/方图: %d/%d/%d（横图占比 %.0f%%）
- 比例情况: %s
- 主导类型: %s

请综合**文本语义**和**图片特征**，推荐最合适的布局。`,
		contentPreview,
		a.ContentLength, a.LineCount, a.ParagraphCount, tagsStr,
		a.HasCode, a.HasHeaders, a.HasLists, a.HasQuotes,
		a.ContentType,
		a.TotalCount-a.VideoCount,
		a.LandscapeCount, a.PortraitCount, a.SquareCount, landscapeRatio,
		ratioMix, a.DominantType)

	return prompt
}

// RuleBasedRecommend 基于综合评分的推荐
// 综合评估所有信息，计算每种布局的适合度分数
//
// 四种布局特点：
// - grid: 方形缩略图，最多9张+N，文字在上，适合快速预览
// - waterfall: 保持原比例，2列错落，文字在下，适合欣赏完整图片
// - horizontal: 固定高度横滑，文字在上，适合连续浏览
// - carousel: 一次一张完整显示，文字在下，适合逐张查看
//
// 返回: (布局, 理由)
func (a *MediaAnalysis) RuleBasedRecommend() (string, string) {
	scores := map[string]float64{
		"grid":       0,
		"waterfall":  0,
		"horizontal": 0,
		"carousel":   0,
	}
	reasons := map[string]string{
		"grid":       "",
		"waterfall":  "",
		"horizontal": "",
		"carousel":   "",
	}

	// ============================================
	// 维度1：文本特征（决定文字是否重要）
	// ============================================

	// 代码块 → 强烈需要先读文字 → grid
	if a.HasCode {
		scores["grid"] += 35
		scores["horizontal"] += 10 // horizontal 也是文字在上
		reasons["grid"] = "代码分享"
	}

	// 结构化内容（标题/列表/引用）→ 文字重要
	structureScore := 0.0
	if a.HasHeaders {
		structureScore += 10
	}
	if a.HasLists {
		structureScore += 10
	}
	if a.HasQuotes {
		structureScore += 5
	}
	if structureScore > 0 {
		scores["grid"] += structureScore
		scores["horizontal"] += structureScore * 0.5
		if reasons["grid"] == "" {
			reasons["grid"] = "结构内容"
		}
	}

	// 文本长度评分
	switch {
	case a.ContentLength >= 150:
		// 很长的文本 → 强烈需要先读
		scores["grid"] += 30
		scores["horizontal"] += 15
		if reasons["grid"] == "" {
			reasons["grid"] = "长文分享"
		}
	case a.ContentLength >= 80:
		// 中长文本
		scores["grid"] += 20
		scores["horizontal"] += 10
		scores["waterfall"] += 5
	case a.ContentLength >= 30:
		// 中等文本
		scores["grid"] += 10
		scores["waterfall"] += 10
		scores["horizontal"] += 5
	case a.ContentLength < 30:
		// 短文本或无文本 → 图片优先
		scores["waterfall"] += 25
		scores["carousel"] += 15
		if reasons["waterfall"] == "" {
			reasons["waterfall"] = "图片展示"
		}
	}

	// ============================================
	// 维度2：图片数量（不同布局对数量的适应性）
	// ============================================

	switch {
	case a.TotalCount >= 15:
		// 非常多的图片 → carousel 逐张或 grid 预览
		scores["carousel"] += 25
		scores["grid"] += 15
		if reasons["carousel"] == "" {
			reasons["carousel"] = "图片较多"
		}
	case a.TotalCount >= 10:
		// 多图 → carousel 或 grid
		scores["carousel"] += 20
		scores["grid"] += 15
		scores["waterfall"] += 5
	case a.TotalCount >= 5:
		// 中等数量
		scores["grid"] += 15
		scores["waterfall"] += 10
		scores["carousel"] += 8
		scores["horizontal"] += 8
	case a.TotalCount == 3 || a.TotalCount == 4:
		// 3-4张 → 各布局都适合
		scores["waterfall"] += 15
		scores["grid"] += 12
		scores["horizontal"] += 10
	case a.TotalCount == 2:
		// 2张 → waterfall 错落好看
		scores["waterfall"] += 18
		scores["grid"] += 10
	case a.TotalCount == 1:
		// 单图 → waterfall 完整展示 或 grid 智能调整
		scores["waterfall"] += 20
		scores["grid"] += 12
		if reasons["waterfall"] == "" {
			reasons["waterfall"] = "单图展示"
		}
	}

	// ============================================
	// 维度3：图片比例特征（决定是否需要保持原比例）
	// ============================================

	if a.TotalCount > 0 {
		landscapeRatio := float64(a.LandscapeCount) / float64(a.TotalCount)
		portraitRatio := float64(a.PortraitCount) / float64(a.TotalCount)
		squareRatio := float64(a.SquareCount) / float64(a.TotalCount)

		// 全是横图 → horizontal 最佳
		if landscapeRatio >= 0.9 && a.TotalCount >= 3 {
			scores["horizontal"] += 30
			reasons["horizontal"] = "全横图"
		} else if landscapeRatio >= 0.7 && a.TotalCount >= 3 {
			scores["horizontal"] += 25
			if reasons["horizontal"] == "" {
				reasons["horizontal"] = "横图为主"
			}
		} else if landscapeRatio >= 0.5 {
			scores["horizontal"] += 15
		}

		// 全是竖图 → waterfall 保持比例更好
		if portraitRatio >= 0.8 {
			scores["waterfall"] += 20
			if reasons["waterfall"] == "" {
				reasons["waterfall"] = "竖图展示"
			}
		}

		// 全是方图 → grid 裁切无损失
		if squareRatio >= 0.8 {
			scores["grid"] += 15
		}

		// 横竖混合 → waterfall 保持各自比例
		if a.LandscapeCount > 0 && a.PortraitCount > 0 {
			mixScore := 20.0
			// 差异越大越需要 waterfall
			if a.RatioVariance > 0.3 {
				mixScore += 10
			}
			scores["waterfall"] += mixScore
			if reasons["waterfall"] == "" {
				reasons["waterfall"] = "比例混合"
			}
		}

		// 宽高比差异大 → waterfall（grid 裁切会损失内容）
		if a.RatioVariance > 0.25 {
			scores["waterfall"] += 15
			scores["grid"] -= 10 // grid 裁切会损失
		}
	}

	// ============================================
	// 维度4：内容类型推断
	// ============================================

	switch a.ContentType {
	case "technical", "code":
		scores["grid"] += 20
		if reasons["grid"] == "" {
			reasons["grid"] = "技术内容"
		}
	case "photography", "art":
		scores["waterfall"] += 30
		reasons["waterfall"] = "摄影作品"
	case "tutorial", "guide":
		scores["carousel"] += 25
		scores["grid"] += 15
		if reasons["carousel"] == "" {
			reasons["carousel"] = "教程步骤"
		}
	case "timeline", "story":
		scores["horizontal"] += 25
		scores["waterfall"] += 15
		if reasons["horizontal"] == "" {
			reasons["horizontal"] = "故事过程"
		}
	case "discussion", "question":
		scores["grid"] += 20
		if reasons["grid"] == "" {
			reasons["grid"] = "讨论内容"
		}
	}

	// ============================================
	// 维度5：标签分析
	// ============================================

	for _, tag := range a.Tags {
		tagLower := strings.ToLower(tag)

		// 摄影/视觉相关 → waterfall
		if containsAny(tagLower, []string{"摄影", "photo", "风景", "随拍", "设计", "插画", "美食", "穿搭"}) {
			scores["waterfall"] += 12
			if reasons["waterfall"] == "" {
				reasons["waterfall"] = "视觉内容"
			}
		}

		// 教程相关 → carousel
		if containsAny(tagLower, []string{"教程", "tutorial", "步骤", "指南", "how"}) {
			scores["carousel"] += 12
			if reasons["carousel"] == "" {
				reasons["carousel"] = "教程内容"
			}
		}

		// 旅行/故事相关 → horizontal
		if containsAny(tagLower, []string{"旅行", "travel", "游记", "旅途", "全景"}) {
			scores["horizontal"] += 12
			if reasons["horizontal"] == "" {
				reasons["horizontal"] = "旅行记录"
			}
		}

		// 技术相关 → grid
		if containsAny(tagLower, []string{"代码", "code", "技术", "开发", "编程"}) {
			scores["grid"] += 10
		}
	}

	// ============================================
	// 维度6：特殊组合加成
	// ============================================

	// 短文本 + 摄影内容 + 少量图片 → waterfall 强加成
	if a.ContentLength < 50 && a.ContentType == "photography" && a.TotalCount <= 6 {
		scores["waterfall"] += 15
	}

	// 长文本 + 代码 → grid 强加成
	if a.ContentLength >= 100 && a.HasCode {
		scores["grid"] += 15
	}

	// 多张横图 + 故事内容 → horizontal 强加成
	if a.LandscapeCount >= 3 && a.ContentType == "timeline" {
		scores["horizontal"] += 15
	}

	// ============================================
	// 选择最高分的布局
	// ============================================

	bestLayout := "grid"
	bestScore := scores["grid"]
	for layout, score := range scores {
		if score > bestScore {
			bestScore = score
			bestLayout = layout
		}
	}

	// 生成理由
	reason := reasons[bestLayout]
	if reason == "" {
		switch bestLayout {
		case "grid":
			reason = "综合推荐"
		case "waterfall":
			reason = "完整展示"
		case "horizontal":
			reason = "连续浏览"
		case "carousel":
			reason = "逐张查看"
		}
	}

	return bestLayout, reason
}

// containsAny 检查字符串是否包含任意一个子串
func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// parseLayoutOutput 解析 AI 输出（格式：布局|理由）
func parseLayoutOutput(output string) (string, string) {
	output = strings.TrimSpace(output)

	// 尝试按 | 分割
	if strings.Contains(output, "|") {
		parts := strings.SplitN(output, "|", 2)
		layout := strings.ToLower(strings.TrimSpace(parts[0]))
		reason := strings.TrimSpace(parts[1])
		return layout, reason
	}

	// 没有 | 分隔符，尝试提取布局名称
	layout := extractLayoutFromOutput(output)
	return layout, ""
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

// AIWrite 使用 AI 对文本进行写作辅助（创作、摘要、纠错、扩写、润色）
func (agentService *AgentService) AIWrite(ctx context.Context, req AIWriteRequest) (*AIWriteResponse, error) {
	if req.Action != "generate" && strings.TrimSpace(req.OriginalContent) == "" {
		return nil, errors.New("内容不能为空")
	}

	// 获取 Agent 设置
	var setting model.AgentSetting
	if err := agentService.settingService.GetAgentInfo(&setting); err != nil {
		return nil, errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}

	if !setting.Enable {
		return nil, errors.New(commonModel.AGENT_NOT_ENABLED)
	}

	var systemPrompt string
	var userMessage string

	switch req.Action {
	case "summarize":
		systemPrompt = `你是一位专业的文字提炼专家。你的任务是对用户提供的文本进行摘要提取。
要求：
1. 提取出最核心的思想和要点
2. 语言简练，重点突出
3. 如果原文较长，请考虑使用列表形式呈现要点
4. 保持客观，不随意添加原文没有的信息
`
		userMessage = fmt.Sprintf("请对以下文本进行摘要提取：\n\n%s", req.OriginalContent)

	case "correct":
		systemPrompt = `你是一位专业的文字校对专家。你的任务是对用户提供的文本进行错别字和语法纠错。
要求：
1. **只修改**错别字、病句和严重的标点符号错误，不要过度修改原本通顺的句子
2. **保持原有排版格式和 Markdown 标签**不受影响
3. 让语句显得自然地道
`
		userMessage = fmt.Sprintf("请对以下文本进行纠错：\n\n%s", req.OriginalContent)

	case "expand":
		systemPrompt = `你是一位富有创造力的作家。你的任务是对用户提供的简短文本进行合理、生动的扩写。
要求：
1. 扩充细节，补充上下文情境，让内容更加丰富饱满、具有吸引力
2. 保持与原文一致的基调和情感倾向
3. 语段过渡自然，**保持原有的 Markdown 格式排版风格**
`
		userMessage = fmt.Sprintf("请合理扩写丰富以下文本：\n\n%s", req.OriginalContent)

	case "generate":
		systemPrompt = `你是一位全能的创意写作助手。请根据用户的需求，创作出高质量的文本内容。
要求：
1. 回答要贴合需求、结构清晰
2. 在合适的地方使用丰富的 Markdown 格式排版（标题、加粗、列表、代码块引用等）来提升内容呈现质量
3. 语言自然生动
`
		if req.OriginalContent != "" {
			userMessage = fmt.Sprintf("基于以下背景参考上下文:\n%s\n\n---\n请执行创作指令：\n%s", req.OriginalContent, req.Prompt)
		} else {
			userMessage = fmt.Sprintf("请帮我创作：\n%s", req.Prompt)
		}

	case "polish":
		// 构建润色风格提示
		styleHint := ""
		switch req.Prompt {
		case "professional":
			styleHint = "请使用专业、克制且严谨的书面风格进行润色，提升其专业度。"
		case "friendly":
			styleHint = "请使用友好、热情且带有亲和力的口吻进行润色，让人感到亲切活泼。"
		case "concise":
			styleHint = "请尽可能精简干练，重点清晰，删除多余废话和定语，保留核心信息。"
		case "general":
			styleHint = "请使用通用流畅的日常文案风格进行润色，力求通顺自然即可。"
		default:
			styleHint = "请根据原文本的内容风格和调性，自动判断最合适的提升和优化方向。"
		}
		systemPrompt = fmt.Sprintf(`你是一位资深的文字编辑和文章润色专家。你的任务是对用户提供的文本进行深度润色优化。

## 润色要求
1. **保持原意**：绝对不改变原文的核心含义和表达意图
2. **保留格式**：保持原有的 Markdown 格式语法（标题、列表、代码块、引用等）
3. **优化表达**：修正语法错误、优化措辞、提升阅读节奏感
4. **自然流畅**：润色后的文字应自然通顺，不显刻板或机器味

## 风格偏好要求
%s
`, styleHint)
		userMessage = fmt.Sprintf("请润色以下文本：\n\n%s", req.OriginalContent)

	default:
		return nil, errors.New("不支持的操作类型")
	}

	// 统一要求格式
	systemPrompt += `
## 返回格式要求
请你严格按以下格式输出结果，必须使用 ===SUMMARY=== 这一行独立文本，分隔你的生成结果文本和你对本次处理的简短摘要说明：

[生成或修改的最终完整文本]
===SUMMARY===
[请用一句简短的话概括你做了哪些处理和优化，例如：已修复多处标点并精简冗余表述，不超过30个汉字]`

	in := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: userMessage,
		},
	}

	output, err := agent.Generate(ctx, setting, in, false, 0.7)
	if err != nil {
		logUtil.GetLogger().Error("[AI Write] AI 调用失败", zap.Error(err))
		return nil, fmt.Errorf("AI 操作失败: %w", err)
	}

	content, summary := parseAIWriteOutput(output)

	logUtil.GetLogger().Info("[AI Write] 工作完成",
		zap.String("action", req.Action),
		zap.Int("original_len", len(req.OriginalContent)),
		zap.Int("result_len", len(content)))

	return &AIWriteResponse{
		Content: content,
		Summary: summary,
	}, nil
}

// parseAIWriteOutput 解析输出，分离主要文本和修改摘要
func parseAIWriteOutput(output string) (string, string) {
	output = strings.TrimSpace(output)

	separator := "===SUMMARY==="
	if strings.Contains(output, separator) {
		parts := strings.SplitN(output, separator, 2)
		content := strings.TrimSpace(parts[0])
		summary := strings.TrimSpace(parts[1])
		return content, summary
	}

	return output, "已执行指定创作/修改操作"
}
