/**
 * 实况照片（Live Photo）工具函数
 */

export interface LivePhotoPair {
  imageIndex: number
  videoIndex: number
  pairId: string // 前端生成的 UUID，用于后端建立关联
}

/**
 * 获取文件基础名（不含扩展名）
 * @param filename 文件名或URL
 * @returns 基础文件名（小写）
 */
export function getBaseName(filename: string): string {
  // 获取文件名（去除路径）
  let name = filename.split('/').pop() || filename
  
  // 处理 URL 中可能的查询参数
  const queryIndex = name.indexOf('?')
  if (queryIndex !== -1) {
    name = name.substring(0, queryIndex)
  }
  
  // 去除扩展名
  const dotIndex = name.lastIndexOf('.')
  if (dotIndex !== -1) {
    name = name.substring(0, dotIndex)
  }
  
  return name.toLowerCase()
}

/**
 * 生成唯一的配对 ID
 * @returns UUID 字符串
 */
export function generatePairId(): string {
  // 使用 crypto.randomUUID() 如果可用，否则使用备用方案
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  // 备用方案：生成简单的唯一 ID
  return `${Date.now()}-${Math.random().toString(36).substring(2, 11)}`
}

/**
 * 检测上传文件中的实况照片对，并生成配对 ID
 * 通过比较基础文件名（不含扩展名）来识别配对
 * @param files 文件列表
 * @returns 实况照片对数组（包含 pairId）
 */
export function detectLivePhotoPairs(files: File[]): LivePhotoPair[] {
  const pairs: LivePhotoPair[] = []
  
  // 分离图片和视频，记录原始索引
  const images: { index: number; baseName: string }[] = []
  const videos: { index: number; baseName: string }[] = []
  
  files.forEach((file, index) => {
    const baseName = getBaseName(file.name)
    if (file.type.startsWith('image/')) {
      images.push({ index, baseName })
    } else if (file.type.startsWith('video/')) {
      videos.push({ index, baseName })
    }
  })

  // 匹配同名的图片和视频
  for (const img of images) {
    for (const vid of videos) {
      if (img.baseName === vid.baseName && img.baseName !== '') {
        pairs.push({
          imageIndex: img.index,
          videoIndex: vid.index,
          pairId: generatePairId()
        })
        break // 一张图片只匹配一个视频
      }
    }
  }
  
  return pairs
}

/**
 * 检测 MediaToAdd 数组中的实况照片对，并生成配对 ID
 * @param media 媒体数组
 * @returns 实况照片对数组（包含 pairId）
 */
export function detectLivePhotoPairsFromMedia(
  media: App.Api.Ech0.MediaToAdd[]
): LivePhotoPair[] {
  const pairs: LivePhotoPair[] = []
  
  const images: { index: number; baseName: string }[] = []
  const videos: { index: number; baseName: string }[] = []
  
  media.forEach((item, index) => {
    const baseName = getBaseName(item.media_url)
    if (item.media_type === 'image') {
      images.push({ index, baseName })
    } else if (item.media_type === 'video') {
      videos.push({ index, baseName })
    }
  })
  
  for (const img of images) {
    for (const vid of videos) {
      if (img.baseName === vid.baseName && img.baseName !== '') {
        pairs.push({
          imageIndex: img.index,
          videoIndex: vid.index,
          pairId: generatePairId()
        })
        break
      }
    }
  }
  
  return pairs
}

/**
 * 为媒体数组应用实况照片配对 ID
 * @param media 媒体数组
 * @param pairs 实况照片对数组
 */
export function applyLivePhotoPairIds(
  media: App.Api.Ech0.MediaToAdd[],
  pairs: LivePhotoPair[]
): void {
  for (const pair of pairs) {
    const imgMedia = media[pair.imageIndex]
    const vidMedia = media[pair.videoIndex]
    if (imgMedia && vidMedia) {
      imgMedia.live_pair_id = pair.pairId
      vidMedia.live_pair_id = pair.pairId
    }
  }
}

/**
 * 检查媒体项是否为实况照片
 * @param item 媒体项
 * @returns 是否为实况照片
 */
export function isLivePhoto(item: App.Api.Ech0.Media | App.Api.Ech0.MediaToAdd): boolean {
  return item.live_video_id !== undefined && item.live_video_id > 0
}

/**
 * 检查媒体项是否为实况照片的视频部分
 * @param item 媒体项
 * @param allMedia 所有媒体项
 * @returns 是否为实况照片的视频部分
 */
export function isLivePhotoVideo(
  item: App.Api.Ech0.Media,
  allMedia: App.Api.Ech0.Media[]
): boolean {
  return allMedia.some(m => m.live_video_id === item.id)
}
