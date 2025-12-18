/**
 * 通用媒体 Fancybox 预览 composable
 * 统一处理图片、视频、实况照片的 Fancybox 预览逻辑
 */
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import '@/styles/fancybox-livephoto.css'

// 媒体项接口
export interface MediaItem {
  id?: number
  media_url?: string
  media_type?: 'image' | 'video'
  live_video_id?: number
  live_pair_id?: string
  [key: string]: any
}

// 配置选项
export interface UseMediaFancyboxOptions {
  // 获取媒体 URL 的函数
  getMediaUrl: (item: MediaItem) => string
  // 获取实况照片视频 URL 的函数（可选，如果不提供则使用默认逻辑）
  getLiveVideoUrl?: (item: MediaItem, allItems: MediaItem[]) => string | null
  // 判断是否为实况照片的函数（可选）
  isLivePhoto?: (item: MediaItem, allItems: MediaItem[]) => boolean
  // 判断是否为视频的函数（可选）
  isVideo?: (item: MediaItem) => boolean
}

// 实况照片自动播放设置
const livePhotoAutoPlay = ref<boolean>(
  typeof localStorage !== 'undefined'
    ? localStorage.getItem('livePhotoAutoPlay') !== 'false'
    : true
)

// 切换自动播放设置
export const toggleLivePhotoAutoPlay = () => {
  livePhotoAutoPlay.value = !livePhotoAutoPlay.value
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('livePhotoAutoPlay', String(livePhotoAutoPlay.value))
  }
}

// 获取自动播放状态
export const getLivePhotoAutoPlay = () => livePhotoAutoPlay.value

export function useMediaFancybox(options: UseMediaFancyboxOptions) {
  const { getMediaUrl } = options

  // 默认判断是否为视频
  const isVideo = options.isVideo || ((item: MediaItem) => item.media_type === 'video')

  // 默认判断是否为实况照片
  const isLivePhoto = options.isLivePhoto || ((item: MediaItem, allItems: MediaItem[]) => {
    // 情况1: 已保存的实况照片 - 有 live_video_id
    if (item.live_video_id !== undefined && item.live_video_id > 0) {
      return true
    }
    // 情况2: 新上传的实况照片 - 图片有 live_pair_id，且存在相同 live_pair_id 的视频
    if (item.media_type === 'image' && item.live_pair_id) {
      return allItems.some(m => m.media_type === 'video' && m.live_pair_id === item.live_pair_id)
    }
    return false
  })

  // 默认获取实况照片视频 URL
  const getLiveVideoUrl = options.getLiveVideoUrl || ((item: MediaItem, allItems: MediaItem[]) => {
    if (!isLivePhoto(item, allItems)) return null

    // 情况1: 已保存的实况照片 - 通过 live_video_id 查找
    if (item.live_video_id) {
      const video = allItems.find(m => m.id === item.live_video_id)
      return video ? getMediaUrl(video) : null
    }

    // 情况2: 新上传的实况照片 - 通过 live_pair_id 查找
    if (item.live_pair_id) {
      const video = allItems.find(m => m.media_type === 'video' && m.live_pair_id === item.live_pair_id)
      return video ? getMediaUrl(video) : null
    }

    return null
  })

  // 检查是否为实况照片的视频部分（应该隐藏）
  const isLivePhotoVideo = (item: MediaItem, allItems: MediaItem[]) => {
    if (item.media_type !== 'video') return false

    // 情况1: 新上传的实况照片的视频 - 通过 live_pair_id 判断
    if (item.live_pair_id) {
      return allItems.some(m => m.media_type === 'image' && m.live_pair_id === item.live_pair_id)
    }

    // 情况2: 已保存的实况照片的视频 - 通过 id 和 live_video_id 判断
    if (item.id) {
      return allItems.some(m => m.live_video_id === item.id)
    }

    return false
  }

  // 过滤可见媒体项（隐藏实况照片的视频部分）
  const getVisibleMediaItems = (items: MediaItem[]) => {
    return items.filter(item => !isLivePhotoVideo(item, items))
  }


  // 初始化实况照片交互
  function initLivePhotoInteraction(slide: any): void {
    try {
      // 查找容器
      let container: HTMLElement | null = null

      if (slide.htmlEl) {
        container = slide.htmlEl.classList?.contains('fancybox-livephoto-container')
          ? slide.htmlEl
          : slide.htmlEl.querySelector('.fancybox-livephoto-container')
      }
      if (!container && slide.el) {
        container = slide.el.querySelector('.fancybox-livephoto-container')
      }

      if (!container) return

      const video = container.querySelector('.livephoto-video') as HTMLVideoElement
      const image = container.querySelector('.livephoto-image') as HTMLImageElement
      const icon = container.querySelector('.livephoto-icon') as HTMLElement
      const dropdown = container.querySelector('.livephoto-dropdown') as HTMLElement
      const controlWrapper = container.querySelector('.livephoto-control-wrapper') as HTMLElement

      if (!video || !image || !icon) return

      const start = (e: Event) => {
        e.stopPropagation()
        e.preventDefault()
        container!.classList.add('zoom')
        video.currentTime = 0
        video.play().catch((err) => {
          console.error('Live photo video play error:', err)
        })
      }

      const leave = (e?: Event) => {
        if (e) {
          e.preventDefault()
        }
        container!.classList.remove('zoom')
        video.pause()
      }

      const handleVideoEnded = () => {
        container!.classList.remove('zoom')
      }

      // 阻止右键菜单和长按菜单
      const preventContextMenu = (e: Event) => {
        e.preventDefault()
        e.stopPropagation()
        return false
      }

      // 下拉菜单控制
      const chevron = icon.querySelector('.livephoto-chevron') as HTMLElement

      const showChevron = () => {
        controlWrapper?.classList.add('show-chevron')
      }

      const hideChevron = () => {
        if (!controlWrapper?.classList.contains('show-dropdown')) {
          controlWrapper?.classList.remove('show-chevron')
        }
      }

      const toggleDropdown = (e: Event) => {
        e.stopPropagation()
        e.preventDefault()
        controlWrapper?.classList.toggle('show-dropdown')
      }

      const hideDropdown = () => {
        controlWrapper?.classList.remove('show-dropdown')
      }

      // 检测是否为移动端
      const isMobile = () => window.innerWidth <= 768

      // 切换自动播放按钮
      const toggleAutoPlayButton = container.querySelector('[data-action="toggle-autoplay"]')
      if (toggleAutoPlayButton) {
        toggleAutoPlayButton.addEventListener('click', (e) => {
          e.stopPropagation()
          toggleLivePhotoAutoPlay()

          // 更新按钮文本和图标
          const text = toggleAutoPlayButton.querySelector('.dropdown-text')
          const iconSvg = toggleAutoPlayButton.querySelector('.dropdown-icon')
          const mainIconSvg = icon.querySelector('.livephoto-icon-svg')

          if (text) {
            text.textContent = livePhotoAutoPlay.value ? '关闭自动播放' : '开启自动播放'
          }
          if (iconSvg) {
            iconSvg.innerHTML = `
              <circle cx="12" cy="12" r="10"/>
              <circle cx="12" cy="12" r="6"/>
              <circle cx="12" cy="12" r="3" fill="currentColor"/>
              ${livePhotoAutoPlay.value ? '' : '<line x1="4" y1="4" x2="20" y2="20"/>'}
            `
          }
          if (mainIconSvg) {
            mainIconSvg.innerHTML = `
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5" fill="none"/>
              <circle cx="12" cy="12" r="6" stroke="currentColor" stroke-width="1.5" fill="none"/>
              <circle cx="12" cy="12" r="3" fill="currentColor"/>
              ${livePhotoAutoPlay.value ? '' : '<line x1="4" y1="4" x2="20" y2="20" stroke="currentColor" stroke-width="1.5"/>'}
            `
          }

          if (isMobile()) {
            hideDropdown()
          }
        })
      }

      // PC端：追踪鼠标是否悬停过菜单
      let hasHoveredDropdown = false

      if (controlWrapper) {
        controlWrapper.addEventListener('mouseenter', () => {
          if (!isMobile()) {
            showChevron()
          }
        })
        controlWrapper.addEventListener('mouseleave', () => {
          if (!isMobile()) {
            hideChevron()
            if (hasHoveredDropdown) {
              hideDropdown()
              hasHoveredDropdown = false
            }
          }
        })
      }

      if (dropdown) {
        dropdown.addEventListener('mouseenter', () => {
          if (!isMobile()) {
            hasHoveredDropdown = true
          }
        })
      }

      if (chevron) {
        chevron.addEventListener('click', (e) => {
          if (!isMobile()) {
            toggleDropdown(e)
            hasHoveredDropdown = false
          }
        })
      }

      if (icon) {
        icon.addEventListener('click', (e) => {
          if (isMobile()) {
            e.stopPropagation()
            toggleDropdown(e)
          }
        })
      }

      // 点击容器外部关闭菜单
      const closeOnClickOutside = (e: MouseEvent) => {
        if (!controlWrapper?.contains(e.target as Node)) {
          hideDropdown()
          hasHoveredDropdown = false
          if (!isMobile()) {
            hideChevron()
          }
        }
      }
      document.addEventListener('click', closeOnClickOutside)

      // 鼠标事件绑定到 icon（PC端悬停播放）
      icon.addEventListener('mouseenter', start)
      icon.addEventListener('mouseleave', leave)

      // 触摸事件绑定到 container
      const livephotoContent = container.querySelector('.livephoto-content') as HTMLElement
      if (livephotoContent) {
        livephotoContent.addEventListener('touchstart', (e: TouchEvent) => {
          if (controlWrapper?.contains(e.target as Node)) {
            return
          }
          start(e)
        }, { passive: false })
        livephotoContent.addEventListener('touchend', (e: TouchEvent) => {
          if (controlWrapper?.contains(e.target as Node)) {
            return
          }
          leave(e)
        }, { passive: false })
        livephotoContent.addEventListener('touchcancel', (e: TouchEvent) => {
          if (controlWrapper?.contains(e.target as Node)) {
            return
          }
          leave(e)
        }, { passive: false })
      }

      // 阻止长按菜单
      container.addEventListener('contextmenu', preventContextMenu, { passive: false })
      icon.addEventListener('contextmenu', preventContextMenu, { passive: false })
      video.addEventListener('ended', handleVideoEnded)

      // 自动播放一次实况照片
      if (livePhotoAutoPlay.value) {
        setTimeout(() => {
          container!.classList.add('zoom')
          video.currentTime = 0
          video.play().catch((err) => {
            console.error('Live photo auto-play error:', err)
          })
        }, 300)
      }

      // 暂停播放（切换 slide 时调用，不清空视频源）
      slide.livePhotoPause = () => {
        video.pause()
        video.currentTime = 0
        container!.classList.remove('zoom')
      }

      // 完全清理（关闭 Fancybox 时调用）
      slide.livePhotoCleanup = () => {
        document.removeEventListener('click', closeOnClickOutside)
        icon.removeEventListener('mouseenter', start)
        icon.removeEventListener('mouseleave', leave)
        if (livephotoContent) {
          livephotoContent.removeEventListener('touchstart', start as any)
          livephotoContent.removeEventListener('touchend', leave as any)
          livephotoContent.removeEventListener('touchcancel', leave as any)
        }
        container.removeEventListener('contextmenu', preventContextMenu)
        icon.removeEventListener('contextmenu', preventContextMenu)
        video.removeEventListener('ended', handleVideoEnded)
        video.pause()
        video.src = ''
      }
    } catch (error) {
      console.error('Live photo init error:', error)
    }
  }


  // 清理实况照片资源
  function cleanupLivePhoto(slide: any): void {
    if (slide?.livePhotoCleanup) {
      try {
        slide.livePhotoCleanup()
        slide.livePhotoCleanup = null
      } catch (error) {
        console.error('Failed to cleanup live photo:', error)
      }
    }
  }

  // 创建实况照片 HTML 内容
  function createLivePhotoHTML(imageUrl: string, videoUrl: string): string {
    const autoPlayEnabled = livePhotoAutoPlay.value
    return `<div class="fancybox-livephoto-container"><div class="livephoto-content"><video class="livephoto-video" src="${videoUrl}" preload="metadata" playsinline disablepictureinpicture></video><img class="livephoto-image" src="${imageUrl}" alt="实况照片" /><div class="livephoto-control-wrapper"><div class="livephoto-icon"><svg class="livephoto-icon-svg ${autoPlayEnabled ? '' : 'disabled'}" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5" fill="none"/><circle cx="12" cy="12" r="6" stroke="currentColor" stroke-width="1.5" fill="none"/><circle cx="12" cy="12" r="3" fill="currentColor"/>${autoPlayEnabled ? '' : '<line x1="4" y1="4" x2="20" y2="20" stroke="currentColor" stroke-width="1.5"/>'}</svg><span class="livephoto-label">LIVE</span><svg class="livephoto-chevron" width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M2 4L6 8L10 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg></div><div class="livephoto-dropdown"><div class="livephoto-dropdown-item" data-action="toggle-autoplay"><span class="dropdown-text">${autoPlayEnabled ? '关闭自动播放' : '开启自动播放'}</span><svg class="dropdown-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="3" fill="currentColor"/>${autoPlayEnabled ? '' : '<line x1="4" y1="4" x2="20" y2="20"/>'}</svg></div></div></div></div></div>`
  }

  // 打开 Fancybox 预览
  function openFancybox(mediaItems: MediaItem[], startIndex: number) {
    const visibleItems = getVisibleMediaItems(mediaItems)

    const items = visibleItems.map((item) => {
      const mediaUrl = getMediaUrl(item)

      if (isLivePhoto(item, mediaItems)) {
        const videoUrl = getLiveVideoUrl(item, mediaItems)
        if (videoUrl) {
          return {
            html: createLivePhotoHTML(mediaUrl, videoUrl),
            class: 'has-livephoto',
            thumb: mediaUrl,
          }
        }
        return {
          src: mediaUrl,
          type: 'image',
          thumb: mediaUrl,
        }
      } else if (isVideo(item)) {
        return {
          src: mediaUrl,
          thumb: mediaUrl,
        }
      } else {
        return {
          src: mediaUrl,
          type: 'image',
          thumb: mediaUrl,
        }
      }
    })

    Fancybox.show(items, {
      theme: 'auto',
      zoomEffect: true,
      fadeEffect: true,
      startIndex: startIndex,
      backdropClick: 'close',
      dragToClose: true,
      closeButton: 'auto',
      keyboard: {
        Escape: 'close',
        ArrowRight: 'next',
        ArrowLeft: 'prev',
        Delete: 'close',
        Backspace: 'close',
        ArrowDown: 'next',
        ArrowUp: 'prev',
        PageUp: 'close',
        PageDown: 'close',
      },
      Carousel: {
        Thumbs: {
          type: 'classic',
          showOnStart: true,
        },
      },
      on: {
        ready: (fancybox: any) => {
          const carousel = fancybox.getCarousel()
          if (!carousel) return

          // 获取起始索引（即 startIndex）
          const currentIndex = startIndex

          // 遍历所有 slides，但只初始化当前显示的那个
          carousel.getSlides().forEach((slide: any, index: number) => {
            if (!slide.html && !slide.htmlEl) return

            const slideEl = slide.el || slide.htmlEl
            const isLivePhotoSlide = slide.htmlEl?.classList?.contains('fancybox-livephoto-container') ||
              slideEl?.querySelector('.fancybox-livephoto-container')

            // 只初始化当前显示的 slide
            if (isLivePhotoSlide && index === currentIndex) {
              initLivePhotoInteraction(slide)
            }
          })

          carousel.on('change', (carousel: any, to: number, from?: number) => {
            if (from !== undefined) {
              const slides = carousel.getSlides()
              const prevSlide = slides[from]
              if (prevSlide) {
                // 只暂停，不完全清理（保留视频源和事件监听）
                if (prevSlide.livePhotoPause) {
                  prevSlide.livePhotoPause()
                } else {
                  // 兼容：如果没有 livePhotoPause，手动暂停
                  const container = prevSlide.el?.querySelector('.fancybox-livephoto-container') ||
                    prevSlide.htmlEl?.querySelector('.fancybox-livephoto-container')
                  if (container) {
                    const video = container.querySelector('.livephoto-video') as HTMLVideoElement
                    if (video) {
                      video.pause()
                      video.currentTime = 0
                    }
                    container.classList.remove('zoom')
                  }
                }
              }
            }

            const currentSlide = carousel.getSlides()[to]
            if (currentSlide && (currentSlide.html || currentSlide.htmlEl)) {
              setTimeout(() => {
                const slideEl = currentSlide.el || currentSlide.htmlEl
                const isLivePhotoSlide = currentSlide.htmlEl?.classList?.contains('fancybox-livephoto-container') ||
                  slideEl?.querySelector('.fancybox-livephoto-container')

                if (isLivePhotoSlide) {
                  // 如果已经初始化过，只需要触发自动播放
                  if (currentSlide.livePhotoCleanup) {
                    // 已初始化，触发自动播放
                    if (livePhotoAutoPlay.value) {
                      const container = slideEl?.querySelector('.fancybox-livephoto-container')
                      const video = container?.querySelector('.livephoto-video') as HTMLVideoElement
                      if (container && video) {
                        setTimeout(() => {
                          container.classList.add('zoom')
                          video.currentTime = 0
                          video.play().catch((err: any) => {
                            console.error('Live photo auto-play error:', err)
                          })
                        }, 300)
                      }
                    }
                  } else {
                    // 未初始化，进行初始化
                    initLivePhotoInteraction(currentSlide)
                  }
                }
              }, 50)
            }
          })
        },
        destroy: (fancybox: any) => {
          const carousel = fancybox.getCarousel()
          if (carousel) {
            carousel.getSlides().forEach((slide: any) => {
              cleanupLivePhoto(slide)
            })
          }
        }
      },
    } as any)
  }

  // 初始化 Fancybox 绑定
  function initFancybox() {
    Fancybox.bind('[data-fancybox]', {})
  }

  // 清理 Fancybox 实例
  function cleanupFancybox() {
    const fancybox = Fancybox.getInstance()
    if (fancybox) {
      const carousel = fancybox.getCarousel()
      if (carousel) {
        carousel.getSlides().forEach((slide: any) => {
          cleanupLivePhoto(slide)
        })
      }
      fancybox.close()
    }
  }

  // 生命周期钩子
  onMounted(() => {
    initFancybox()
  })

  onBeforeUnmount(() => {
    cleanupFancybox()
  })

  return {
    openFancybox,
    getVisibleMediaItems,
    isLivePhoto,
    isLivePhotoVideo,
    isVideo,
    getLiveVideoUrl,
    initFancybox,
    cleanupFancybox,
  }
}
