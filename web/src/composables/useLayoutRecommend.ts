import { ref } from 'vue'
import { fetchRecommendLayout, type MediaInfo, type ContentInfo, type LayoutRecommendRequest } from '@/service/api/agent'
import { ImageLayout } from '@/enums/enums'

/**
 * AI 布局推荐 composable
 * 根据媒体信息和内容信息调用 AI 推荐最佳布局
 */
export function useLayoutRecommend() {
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  /**
   * 根据媒体列表和内容信息推荐布局
   * @param mediaList 媒体信息列表
   * @param contentInfo 可选的内容信息（文本长度、标签等）
   * @returns 推荐的布局类型
   */
  const recommendLayout = async (
    mediaList: MediaInfo[],
    contentInfo?: ContentInfo
  ): Promise<ImageLayout> => {
    if (!mediaList || mediaList.length === 0) {
      console.log('[AI Layout] 媒体列表为空，使用默认布局')
      return ImageLayout.GRID
    }

    isLoading.value = true
    error.value = null

    try {
      console.log('[AI Layout] 开始推荐，媒体数量:', mediaList.length, '文本长度:', contentInfo?.content_length || 0)
      const request: LayoutRecommendRequest = {
        media_list: mediaList,
        content_info: contentInfo,
      }
      const res = await fetchRecommendLayout(request)

      if (res.data) {
        // 新的响应格式：{ layout: string, source: string }
        const result = res.data as { layout?: string; source?: string } | string

        // 兼容旧格式（直接返回字符串）和新格式（返回对象）
        let layout: string
        let source: string

        if (typeof result === 'string') {
          layout = result
          source = 'unknown'
        } else {
          layout = result.layout || ''
          source = result.source || 'unknown'
        }

        if (Object.values(ImageLayout).includes(layout as ImageLayout)) {
          const sourceLabel = source === 'ai' ? '🤖 AI推荐' : source === 'rule' ? '📐 规则引擎' : '推荐'
          console.log(`[AI Layout] ${sourceLabel}: ${layout}`)
          return layout as ImageLayout
        }
        console.warn('[AI Layout] 无效布局:', layout, '使用默认')
      }
      return ImageLayout.GRID
    } catch (e: any) {
      error.value = e.message || '布局推荐失败'
      console.error('[AI Layout] 推荐失败:', e.message || e)
      return ImageLayout.GRID
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 从 Media 对象数组中提取 MediaInfo
   * 注意：实况照片和视频在布局预览中与普通图片处理相同，只需要基本的宽高比信息
   * 但需要过滤掉实况照片的视频部分（它们在 visibleMediaItems 中不会显示）
   */
  const extractMediaInfo = (media: App.Api.Ech0.Media[]): MediaInfo[] => {
    // 过滤掉实况照片的视频部分（它们在 visibleMediaItems 中不会显示）
    const visibleMedia = media.filter(m => {
      // 如果是视频，检查是否被某个图片作为 live_video_id 引用
      if (m.media_type === 'video') {
        return !media.some(other => other.live_video_id === m.id)
      }
      return true
    })

    return visibleMedia.map(m => ({
      width: m.width || 0,
      height: m.height || 0,
      media_type: m.media_type || 'image',
    }))
  }

  /**
   * 深度分析文本内容，提取 ContentInfo
   * @param content 文本内容
   * @param tags 标签列表
   */
  const extractContentInfo = (content: string, tags?: { name: string }[]): ContentInfo => {
    // 代码块检测（```code``` 或 `inline code`）
    const hasCode = /```[\s\S]*?```|`[^`\n]+`/.test(content)

    // 链接检测（[text](url) 或 https://url）
    const hasLinks = /\[.*?\]\(.*?\)|https?:\/\/\S+/.test(content)

    // Markdown 图片检测 ![alt](url)
    const hasImagesInText = /!\[.*?\]\(.*?\)/.test(content)

    // 标题检测（# ## ### 等）
    const hasHeaders = /^#{1,6}\s+.+/m.test(content)

    // 列表检测（- * 1. 等）
    const hasLists = /^[\s]*[-*+]\s+.+|^[\s]*\d+\.\s+.+/m.test(content)

    // 引用块检测（> ）
    const hasQuotes = /^>\s+.+/m.test(content)

    // 行数统计
    const lines = content.split('\n')
    const lineCount = lines.length

    // 段落数统计（通过空行分隔）
    const paragraphs = content.split(/\n\s*\n/).filter(p => p.trim().length > 0)
    const paragraphCount = paragraphs.length

    return {
      content_length: content.length,
      content: content, // 传递原始内容用于 AI 深度分析
      has_code: hasCode,
      has_links: hasLinks,
      has_images_in_text: hasImagesInText,
      has_headers: hasHeaders,
      has_lists: hasLists,
      has_quotes: hasQuotes,
      line_count: lineCount,
      paragraph_count: paragraphCount,
      tags: tags?.map(t => t.name) || [],
    }
  }

  return {
    isLoading,
    error,
    recommendLayout,
    extractMediaInfo,
    extractContentInfo,
  }
}
