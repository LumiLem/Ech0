import { getApiUrl } from '@/service/request/shared'
import { ImageSource, MusicProvider } from '@/enums/enums'

// 获取媒体链接（支持图片和视频）
export const getMediaUrl = (media: App.Api.Ech0.Media) => {
  if (media.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(media.media_url)
  } else if (media.media_source === ImageSource.URL) {
    return String(media.media_url)
  } else if (media.media_source === ImageSource.S3) {
    return String(media.media_url)
  } else {
    // 未知的媒体来源，按照本地媒体处理
    return getApiUrl() + String(media.media_url)
  }
}

// 获取图片链接（保留向后兼容）
export const getImageUrl = (image: App.Api.Ech0.Image) => {
  if (image.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(image.media_url)
  } else if (image.media_source === ImageSource.URL) {
    return String(image.media_url)
  } else if (image.media_source === ImageSource.S3) {
    return String(image.media_url)
  } else {
    // 未知的图片来源，按照本地图片处理
    return getApiUrl() + String(image.media_url)
  }
}

// 获取待添加媒体链接
export const getMediaToAddUrl = (media: App.Api.Ech0.MediaToAdd) => {
  if (media.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(media.media_url)
  } else if (media.media_source === ImageSource.URL) {
    return String(media.media_url)
  } else if (media.media_source === ImageSource.S3) {
    return String(media.media_url)
  } else {
    // 未知的媒体来源，按照本地媒体处理
    return getApiUrl() + String(media.media_url)
  }
}

// 获取待添加图片链接（向后兼容）
export const getImageToAddUrl = (image: App.Api.Ech0.ImageToAdd) => {
  if (image.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(image.media_url)
  } else if (image.media_source === ImageSource.URL) {
    return String(image.media_url)
  } else if (image.media_source === ImageSource.S3) {
    return String(image.media_url)
  } else {
    // 未知的图片来源，按照本地图片处理
    return getApiUrl() + String(image.media_url)
  }
}

export const formatDate = (dateString: string) => {
  // 当天则显示（时：分）
  // 非当天但是三内天则显示几天前
  // 超过三天则显示（时：分 年月日）
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const diffInDays = Math.floor(diff / (1000 * 60 * 60 * 24))
  const diffInHours = Math.floor(diff / (1000 * 60 * 60))
  const diffInMinutes = Math.floor(diff / (1000 * 60))

  const diffInSeconds = Math.floor(diff / 1000)
  if (diffInSeconds < 60) {
    return '刚刚'
  } else if (diffInMinutes < 60) {
    return `${diffInMinutes}分钟前`
  } else if (diffInHours < 24) {
    return `${diffInHours}小时前`
  } else if (diffInDays < 3) {
    return `${diffInDays}天前`
  } else {
    const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    const weekDay = weekDays[date.getDay()]
    
    // 如果是今年，不显示年份
    const isCurrentYear = date.getFullYear() === now.getFullYear()
    if (isCurrentYear) {
      return `${date.getMonth() + 1}月${date.getDate()}日 ${weekDay}`
    } else {
      return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 ${weekDay}`
    }
  }
}

// 格式化详细时间，今年不显示年份："12月10日 周二 14:30"，非今年显示年份："2023年12月10日 周二 14:30"
export const formatDetailedTime = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  const weekDay = weekDays[date.getDay()]
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')

  // 如果是今年，不显示年份
  const isCurrentYear = date.getFullYear() === now.getFullYear()
  if (isCurrentYear) {
    return `${date.getMonth() + 1}月${date.getDate()}日 ${weekDay} ${hours}:${minutes}`
  } else {
    return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 ${weekDay} ${hours}:${minutes}`
  }
}

// 解析音乐链接（网易云、QQ音乐、Apple Music）
export const parseMusicURL = (url: string) => {
  url = url.trim()

  /* =======================
   * 网易云音乐
   * ======================= */
  if (/^https:\/\/([a-z0-9-]+\.)*music\.163\.com/i.test(url)) {
    // id：永远从 query 拿
    const idMatch = url.match(/[?&]id=(\d+)/)
    if (!idMatch) return null

    let type: 'song' | 'playlist' | 'album' | undefined

    if (/(\/|#\/|\/m\/)song/.test(url)) {
      type = 'song'
    } else if (/(\/|#\/|\/m\/)playlist/.test(url)) {
      type = 'playlist'
    }
    // else if (/(\/|#\/|\/m\/)album/.test(url)) {
    //   type = 'album'
    // }

    if (!type) return null

    return {
      server: MusicProvider.NETEASE,
      type,
      id: idMatch[1],
    }
  }

  /* =======================
   * QQ 音乐
   * ======================= */
  if (/^https:\/\/([a-z0-9-]+\.)*qq\.com/i.test(url)) {
    // 新版 songDetail（字母数字混合）
    const newSongMatch = url.match(/songDetail\/([a-zA-Z0-9]+)/)
    if (newSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: newSongMatch[1],
      }
    }

    // 旧版 songid（纯数字）
    const oldSongMatch = url.match(/[?&]songid=(\d+)/)
    if (oldSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: oldSongMatch[1],
      }
    }

    // 匹配 /playlist/ 后面的数字ID
    const playlistMatch = url.match(/\/playlist\/(\d+)/i)
    if (playlistMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'playlist',
        id: playlistMatch[1],
      }
    }
    return null
  }

  /* =======================
   * Apple Music
   * ======================= */
  if (/^https:\/\/music\.apple\.com/i.test(url)) {
    const appleMatch = url.match(/\/(song|album)\/[^/]+\/(\d+)/)
    if (!appleMatch) return null

    return {
      server: MusicProvider.APPLE,
      type: appleMatch[1],
      id: appleMatch[2],
    }
  }
  return null
}

/**
 * 从一段文本中提取并返回最短、规范的音乐链接
 */
export const extractAndCleanMusicURL = (input: string): string | null => {
  const text = input.trim()

  // 1️⃣ 粗暴提取第一个 URL（足够鲁棒）
  const urlMatch = text.match(/https?:\/\/[^\s]+/i)
  if (!urlMatch) return null

  const rawUrl = urlMatch[0]

  // 2️⃣ 复用统一解析函数
  const parsed = parseMusicURL(rawUrl)
  if (!parsed) return null

  // 3️⃣ 规范化重组
  switch (parsed.server) {
    case MusicProvider.NETEASE: {
      // 统一为 PC 可打开的最短链接
      return `https://music.163.com/#/${parsed.type}?id=${parsed.id}`
    }

    case MusicProvider.QQ: {
      // 根据 type 分别生成歌曲或歌单的规范化链接
      if (parsed.type === 'song') {
        return `https://y.qq.com/n/ryqq_v2/songDetail/${parsed.id}`
      }

      if (parsed.type === 'playlist') {
        // 使用统一的歌单路径格式
        return `https://y.qq.com/n/ryqq_v2/playlist/${parsed.id}`
      }

      // 如果是其他未知类型，则返回 null
      return null
    }

    case MusicProvider.APPLE: {
      // 去掉后面的参数
      const cleanUrl = rawUrl.split('?')[0]
      return cleanUrl ?? rawUrl
    }

    default:
      return null
  }
}

// 获取图片尺寸
export function getImageSize(imgUrl: string): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      resolve({ width: img.width, height: img.height })
    }
    img.onerror = (err) => {
      reject(err)
    }
    img.src = imgUrl
  })
}

// 获取 HubEcho 的媒体链接（支持图片和视频）
export const getHubMediaUrl = (media: App.Api.Ech0.Media, baseurl: string) => {
  if (media.media_source === ImageSource.LOCAL) {
    return baseurl + '/api' + String(media.media_url)
  } else if (media.media_source === ImageSource.URL) {
    return String(media.media_url)
  } else if (media.media_source === ImageSource.S3) {
    return String(media.media_url)
  } else {
    // 未知的媒体来源，按照本地媒体处理
    return baseurl + '/api' + String(media.media_url)
  }
}

// 获取 HubEcho 的图片（保留向后兼容）
export const getHubImageUrl = (image: App.Api.Ech0.Image, baseurl: string) => {
  if (image.media_source === ImageSource.LOCAL) {
    return baseurl + '/api' + String(image.media_url)
  } else if (image.media_source === ImageSource.URL) {
    return String(image.media_url)
  } else if (image.media_source === ImageSource.S3) {
    return String(image.media_url)
  } else {
    // 未知的图片来源，按照本地图片处理
    return baseurl + '/api' + String(image.media_url)
  }
}

/**
 * Base64URL to Uint8Array
 * 用于解析服务端返回的 WebAuthn publicKey
 */
export function base64urlToUint8Array(input: string): Uint8Array<ArrayBuffer> {
  const base64 = input.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4 === 0 ? '' : '='.repeat(4 - (base64.length % 4))
  const binary = atob(base64 + pad)
  const buffer = new ArrayBuffer(binary.length)
  const bytes = new Uint8Array(buffer)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}

/**
 * Uint8Array to Base64URL
 * 用于生成客户端返回的 WebAuthn publicKey
 */
export function uint8ArrayToBase64url(bytes: ArrayBuffer | Uint8Array): string {
  const u8 = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  let binary = ''
  for (let i = 0; i < u8.length; i++) binary += String.fromCharCode(u8[i]!)
  const base64 = btoa(binary)
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}
