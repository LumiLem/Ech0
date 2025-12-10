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

// 常量定义
const FTYP_SIGNATURE = [0x66, 0x74, 0x79, 0x70] as const // "ftyp" in ASCII
const MIN_EMBEDDED_FILE_SIZE = 1024 * 1024 // 1MB
const SEARCH_START_PERCENTAGE = 0.05 // 5%
const SEARCH_END_PERCENTAGE = 0.8 // 80% - ftyp后面必须有足够空间存放视频数据
const MP4_BOX_HEADER_SIZE = 4 // MP4 box header size

/**
 * 在字节数组中查找 ftyp 标记位置
 * 搜索范围：5% → 80%
 * 
 * @param uint8Array 文件字节数组
 * @param ftypSignature ftyp 标记字节序列
 * @returns ftyp 位置，未找到返回 -1
 */
function findFtypPosition(uint8Array: Uint8Array, ftypSignature: readonly number[]): number {
  // 搜索范围：
  // - 从5%开始：跳过JPEG图片数据的主要部分，但保留足够的覆盖范围
  // - 到80%结束：ftyp后面必须有足够空间存放完整的视频数据
  const searchStart = Math.floor(uint8Array.length * SEARCH_START_PERCENTAGE) // 5%
  const searchEnd = Math.floor(uint8Array.length * SEARCH_END_PERCENTAGE) // 80%
  
  for (let i = searchStart; i < searchEnd; i++) {
    if (
      uint8Array[i] === ftypSignature[0] &&
      uint8Array[i + 1] === ftypSignature[1] &&
      uint8Array[i + 2] === ftypSignature[2] &&
      uint8Array[i + 3] === ftypSignature[3]
    ) {
      if (i >= MP4_BOX_HEADER_SIZE) {
        return i
      }
    }
  }
  
  return -1 // 未找到
}

/**
 * 检测文件是否为嵌入式实况照片，并返回检测结果
 * 支持小米、三星等厂商的嵌入式实况照片格式
 * 格式：图片数据 + 特殊标记 + 视频数据
 * 
 * @param file 要检测的文件
 * @returns 检测结果：{ isEmbedded: boolean, ftypPosition?: number, uint8Array?: Uint8Array }
 */
export async function detectEmbeddedMotionPhoto(file: File): Promise<{
  isEmbedded: boolean
  ftypPosition?: number
  uint8Array?: Uint8Array
}> {
  const startTime = performance.now()
  
  try {
    // 只检测 JPEG 格式的图片
    const isJpeg = file.type === 'image/jpeg' || 
                   file.type === 'image/jpg' ||
                   file.name.toLowerCase().endsWith('.jpg') ||
                   file.name.toLowerCase().endsWith('.jpeg')
    
    if (!isJpeg) {
      return { isEmbedded: false }
    }

    // 文件大小必须大于 1MB
    if (!file.size || file.size < MIN_EMBEDDED_FILE_SIZE) {
      return { isEmbedded: false }
    }

    // 读取文件内容
    const arrayBuffer = await file.arrayBuffer()
    const uint8Array = new Uint8Array(arrayBuffer)

    // 使用统一的搜索函数查找 MP4 文件头标记
    const ftypPosition = findFtypPosition(uint8Array, FTYP_SIGNATURE)
    const endTime = performance.now()
    const duration = Math.round(endTime - startTime)
    
    if (ftypPosition !== -1) {
      console.log('✅ 检测到嵌入式实况照片:', file.name, `(${duration}ms)`, 'ftyp位置:', ((ftypPosition / uint8Array.length) * 100).toFixed(1) + '%')
      return { 
        isEmbedded: true, 
        ftypPosition, 
        uint8Array 
      }
    }
    
    return { isEmbedded: false }
  } catch (error) {
    console.error('检测嵌入式实况照片失败:', error)
    return { isEmbedded: false }
  }
}

/**
 * 检测文件是否为嵌入式实况照片（简化版本，保持向后兼容）
 * @param file 要检测的文件
 * @returns 是否为嵌入式实况照片
 */
export async function isEmbeddedMotionPhoto(file: File): Promise<boolean> {
  const result = await detectEmbeddedMotionPhoto(file)
  return result.isEmbedded
}

/**
 * 分离嵌入式实况照片为图片和视频两个文件
 * 
 * @param file 嵌入式实况照片文件
 * @param ftypPosition 可选的 ftyp 位置（如果已知，避免重复搜索）
 * @param uint8Array 可选的文件字节数组（如果已读取，避免重复读取）
 * @returns 分离后的图片和视频文件，失败返回 null
 */
export async function separateEmbeddedMotionPhoto(
  file: File,
  ftypPosition?: number,
  uint8Array?: Uint8Array
): Promise<{ imageFile: File; videoFile: File } | null> {
  try {
    console.log('🔧 分离嵌入式实况照片:', file.name)
    
    // 如果没有提供字节数组，则读取文件
    let fileData = uint8Array
    if (!fileData) {
      const arrayBuffer = await file.arrayBuffer()
      fileData = new Uint8Array(arrayBuffer)
    }

    // 如果没有提供 ftyp 位置，则搜索
    let ftypPos = ftypPosition
    if (ftypPos === undefined) {
      ftypPos = findFtypPosition(fileData, FTYP_SIGNATURE)
      if (ftypPos === -1) {
        console.error('❌ 分离失败：未找到视频数据起始位置')
        return null
      }
    }
    
    // MP4 box 格式：4字节大小 + 4字节类型，所以视频起始位置是 ftyp 前面的 box header
    const videoStartPos = ftypPos - MP4_BOX_HEADER_SIZE
    console.log('✅ 使用已知位置分离视频数据，起始位置:', videoStartPos)

    // console.log('📍 视频数据起始位置:', videoStartPos, '占文件比例:', ((videoStartPos / uint8Array.length) * 100).toFixed(1) + '%')

    // 分离图片数据（从开始到视频起始位置）
    const imageData = fileData.slice(0, videoStartPos)
    const imageBlob = new Blob([imageData], { type: 'image/jpeg' })
    
    // 分离视频数据（从视频起始位置到文件末尾）
    const videoData = fileData.slice(videoStartPos)
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
      originalSize: fileData.length,
      originalSizeMB: (fileData.length / (1024 * 1024)).toFixed(2) + 'MB',
      imageName: imageFile.name,
      imageSize: imageBlob.size,
      imageSizeMB: (imageBlob.size / (1024 * 1024)).toFixed(2) + 'MB',
      videoName: videoFile.name,
      videoSize: videoBlob.size,
      videoSizeMB: (videoBlob.size / (1024 * 1024)).toFixed(2) + 'MB',
      videoStartPos: videoStartPos,
      videoStartPercentage: ((videoStartPos / fileData.length) * 100).toFixed(1) + '%'
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
