import { request } from '../request'

// 获取近况总结
export function fetchGetRecent() {
  return request<string>({
    url: '/agent/recent',
    method: 'GET',
  })
}

// 媒体信息
export interface MediaInfo {
  width: number
  height: number
  media_type: string // image, video
}

// 内容信息（用于综合分析）
export interface ContentInfo {
  // 文本基础信息
  content_length: number // 文本长度（字符数）
  content: string // 文本内容（用于深度分析）
  // 内容结构分析
  has_code: boolean // 是否包含代码块
  has_links: boolean // 是否包含链接
  has_images_in_text: boolean // Markdown中是否有图片引用
  has_headers: boolean // 是否有标题（#）
  has_lists: boolean // 是否有列表
  has_quotes: boolean // 是否有引用块
  line_count: number // 行数
  paragraph_count: number // 段落数
  // 标签信息
  tags: string[] // 标签列表
}

// 推荐布局请求
export interface LayoutRecommendRequest {
  media_list: MediaInfo[]
  content_info?: ContentInfo // 可选的内容信息
}

// 推荐媒体布局
export function fetchRecommendLayout(data: LayoutRecommendRequest) {
  return request<string>({
    url: '/agent/recommend-layout',
    method: 'POST',
    data,
  })
}
