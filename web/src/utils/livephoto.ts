/**
 * 实况照片（Live Photo）工具函数
 * 包含实况照片配对检测和嵌入式实况照片分离功能
 */

export interface LivePhotoPair {
  imageIndex: number
  videoIndex: number
  pairId: string // 前端生成的 UUID，用于后端建立关联
}

/* ==================== 嵌入式实况照片分离 ==================== */

/**
 * 检测文件是否为嵌入式实况照片
 * 支持小米、三星等厂商的嵌入式实况照片格式
 * 格式：图片数据 + 特殊标记 + 视频数据
 * 
 * @param file 要检测的文件
 * @returns 是否为嵌入式实况照片
 */
export async function isEmbeddedMotionPhoto(file: File): Promise<boolean> {
  try {
    // 只检测图片文件
    if (!file.type.startsWith('image/')) {
      return false
    }

    // 读取文件内容
    const arrayBuffer = await file.arrayBuffer()
    const uint8Array = new Uint8Array(arrayBuffer)

    // 查找 MP4 文件头标记 (ftyp)
    // MP4 文件通常以 "ftyp" 标记开始，位于文件的某个位置
    const ftypSignature = [0x66, 0x74, 0x79, 0x70] // "ftyp" in ASCII
    
    // 从文件中间开始搜索（跳过图片数据）
    // 通常图片数据不会超过文件的前 80%
    const searchStart = Math.floor(uint8Array.length * 0.3)
    
    for (let i = searchStart; i < uint8Array.length - 4; i++) {
      if (
        uint8Array[i] === ftypSignature[0] &&
        uint8Array[i + 1] === ftypSignature[1] &&
        uint8Array[i + 2] === ftypSignature[2] &&
        uint8Array[i + 3] === ftypSignature[3]
      ) {
        // 找到 ftyp 标记，检查前面是否有 MP4 box size
        // MP4 box 格式：4字节大小 + 4字节类型
        if (i >= 4) {
          console.log('检测到嵌入式实况照片，ftyp 位置:', i)
          return true
        }
      }
    }

    return false
  } catch (error) {
    console.error('检测嵌入式实况照片失败:', error)
    return false
  }
}

/**
 * 分离嵌入式实况照片为图片和视频两个文件
 * 
 * @param file 嵌入式实况照片文件
 * @returns 分离后的图片和视频文件，失败返回 null
 */
export async function separateEmbeddedMotionPhoto(
  file: File
): Promise<{ imageFile: File; videoFile: File } | null> {
  try {
    const arrayBuffer = await file.arrayBuffer()
    const uint8Array = new Uint8Array(arrayBuffer)

    // 查找 MP4 文件头标记
    const ftypSignature = [0x66, 0x74, 0x79, 0x70] // "ftyp"
    const searchStart = Math.floor(uint8Array.length * 0.3)
    
    let videoStartPos = -1
    
    for (let i = searchStart; i < uint8Array.length - 4; i++) {
      if (
        uint8Array[i] === ftypSignature[0] &&
        uint8Array[i + 1] === ftypSignature[1] &&
        uint8Array[i + 2] === ftypSignature[2] &&
        uint8Array[i + 3] === ftypSignature[3]
      ) {
        // 找到 ftyp，往前找 MP4 box 的起始位置（4字节大小信息）
        if (i >= 4) {
          videoStartPos = i - 4
          break
        }
      }
    }

    if (videoStartPos === -1) {
      console.error('未找到视频数据起始位置')
      return null
    }

    console.log('视频数据起始位置:', videoStartPos)

    // 分离图片数据（从开始到视频起始位置）
    const imageData = uint8Array.slice(0, videoStartPos)
    const imageBlob = new Blob([imageData], { type: 'image/jpeg' })
    
    // 分离视频数据（从视频起始位置到文件末尾）
    const videoData = uint8Array.slice(videoStartPos)
    const videoBlob = new Blob([videoData], { type: 'video/mp4' })

    // 生成文件名
    const originalName = file.name.replace(/\.[^/.]+$/, '') // 去除扩展名
    const imageFile = new File([imageBlob], `${originalName}.jpg`, {
      type: 'image/jpeg',
      lastModified: file.lastModified,
    })
    const videoFile = new File([videoBlob], `${originalName}.mp4`, {
      type: 'video/mp4',
      lastModified: file.lastModified,
    })

    console.log('✅ 嵌入式实况照片分离成功:', {
      original: file.name,
      originalName: originalName,
      imageName: imageFile.name,
      videoName: videoFile.name,
      imageSize: imageBlob.size,
      videoSize: videoBlob.size,
    })

    return { imageFile, videoFile }
  } catch (error) {
    console.error('分离嵌入式实况照片失败:', error)
    return null
  }
}

/* ==================== 实况照片配对检测 ==================== */

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
