/**
 * 图片处理工具
 * 
 * 获取 Store 中的配置，自动把用户填写的拼接参数加到 URL 后面。
 * 支持两种场景（scene）:
 * - 'thumb': 列表预览使用（缩略图拼参）
 * - 'full': 详情大图使用（大图拼参）
 */

import { useSettingStore } from '@/stores'

/**
 * 构建包含图片处理参数的 URL
 *
 * @param originalUrl - 原始图片 URL
 * @param source - 图片来源 ('local' | 's3' | 'url')
 * @param scene - 处理场景 ('thumb' | 'full')
 * @returns 拼接处理参数后的 URL
 */
export function buildProcessedImageUrl(
  originalUrl: string,
  source: string,
  scene: 'thumb' | 'full',
): string {
  if (!originalUrl) return originalUrl

  // 外链图片不由我们处理
  if (source === 'url') return originalUrl

  const settingStore = useSettingStore()
  const settings = settingStore.ImageProcessSetting

  let param = ''

  if (source === 'local') {
    if (!settings.local_process) return originalUrl
    param = scene === 'thumb' ? settings.local_thumb_param : settings.local_full_param
  } else if (source === 's3') {
    if (!settings.s3_process) return originalUrl
    param = scene === 'thumb' ? settings.s3_thumb_param : settings.s3_full_param
  }

  // 如果没有具体参数，直接返回原链接
  if (!param) return originalUrl

  param = String(param).trim()

  // 智能拼接参数
  const hasQuery = originalUrl.includes('?')
  if (param.startsWith('?') || param.startsWith('&')) {
    const separator = hasQuery ? '&' : '?'
    return `${originalUrl}${separator}${param.slice(1)}`
  }

  // 极端后备：如果用户没写 ?，我们给加上（如果是合法情况）
  return `${originalUrl}${hasQuery ? '&' : '?'}${param}`
}
