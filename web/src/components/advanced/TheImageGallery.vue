<template>
  <div v-if="visibleMediaItems.length" class="image-gallery-container">
    <!-- 瀑布流布局 -->
    <div
      v-if="layout === ImageLayout.WATERFALL || !layout"
      :class="[
        'imgwidth mx-auto grid gap-2 mb-4',
        visibleMediaItems.length === 1 ? 'grid-cols-1 justify-items-center' : 'grid-cols-2',
      ]"
    >
      <template v-for="(item, idx) in visibleMediaItems" :key="idx">
        <!-- 实况照片 -->
        <button
          v-if="isLivePhoto(item)"
          class="livephoto-preview bg-transparent border-0 p-0 cursor-pointer w-fit relative"
          :class="getColSpan(idx, visibleMediaItems.length)"
          @click="openFancybox(idx)"
        >
          <img
            :src="getMediaUrlCompat(item)"
            :alt="`实况照片${idx + 1}`"
            loading="lazy"
            class="echoimg block max-w-full h-auto"
          />
          <div class="livephoto-overlay">
            <LivePhotoIcon class="livephoto-icon" color="#ffffff" />
            <span class="livephoto-text">LIVE</span>
          </div>
        </button>
        
        <!-- 普通图片 -->
        <button
          v-else-if="!isVideo(item)"
          class="bg-transparent border-0 p-0 cursor-pointer w-fit"
          :class="getColSpan(idx, visibleMediaItems.length)"
          @click="openFancybox(idx)"
        >
          <img
            :src="getMediaUrlCompat(item)"
            :alt="`预览图片${idx + 1}`"
            loading="lazy"
            class="echoimg block max-w-full h-auto"
          />
        </button>
        
        <!-- 视频 -->
        <button
          v-else
          class="video-preview bg-transparent border-0 p-0 cursor-pointer w-fit relative"
          :class="getColSpan(idx, visibleMediaItems.length)"
          @click="openFancybox(idx)"
        >
          <video
            :src="getMediaUrlCompat(item) + '#t=0.1'"
            preload="metadata"
            muted
            playsinline
            class="echoimg video-thumb block max-w-full h-auto"
          ></video>
          <div class="play-overlay">
            <Play class="play-icon" color="#ffffff" />
          </div>
        </button>
      </template>
    </div>

    <!-- 九宫格布局 -->
    <div v-if="layout === ImageLayout.GRID" class="imgwidth mx-auto mb-4">
      <div class="grid grid-cols-3 gap-2">
        <template v-for="(item, idx) in displayedVisibleImages" :key="idx">
          <!-- 实况照片 -->
          <button
            v-if="isLivePhoto(item)"
            class="livephoto-preview bg-transparent border-0 p-0 cursor-pointer overflow-hidden aspect-square relative"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaUrlCompat(item)"
              :alt="`实况照片${idx + 1}`"
              loading="lazy"
              class="echoimg w-full h-full object-cover"
            />
            <div class="livephoto-overlay livephoto-overlay-small">
              <LivePhotoIcon class="livephoto-icon-small" color="#ffffff" />
            </div>
            <div v-if="extraVisibleCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraVisibleCount }}
            </div>
          </button>
          
          <!-- 普通图片 -->
          <button
            v-else-if="!isVideo(item)"
            class="bg-transparent border-0 p-0 cursor-pointer overflow-hidden aspect-square relative"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaUrlCompat(item)"
              :alt="`预览图片${idx + 1}`"
              loading="lazy"
              class="echoimg w-full h-full object-cover"
            />
            <div v-if="extraVisibleCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraVisibleCount }}
            </div>
          </button>
          
          <!-- 视频 -->
          <button
            v-else
            class="video-preview bg-transparent border-0 p-0 cursor-pointer overflow-hidden aspect-square relative"
            @click="openFancybox(idx)"
          >
            <video
              :src="getMediaUrlCompat(item) + '#t=0.1'"
              preload="metadata"
              muted
              playsinline
              class="echoimg video-thumb w-full h-full object-cover"
            ></video>
            <div class="play-overlay">
              <Play class="play-icon" color="#ffffff" />
            </div>
            <div v-if="extraVisibleCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraVisibleCount }}
            </div>
          </button>
        </template>
      </div>
    </div>

    <!-- 单图轮播布局 -->
    <div v-if="layout === ImageLayout.CAROUSEL" class="imgwidth mx-auto mb-4">
      <div class="carousel-container rounded-lg overflow-hidden">
        <template v-if="visibleMediaItems[carouselIndex]">
          <!-- 实况照片 -->
          <button
            v-if="isLivePhoto(visibleMediaItems[carouselIndex]!)"
            class="livephoto-preview carousel-slide bg-transparent border-0 p-0 cursor-pointer w-full relative"
            @click="openFancybox(carouselIndex)"
          >
            <img
              :src="getMediaUrlCompat(visibleMediaItems[carouselIndex]!)"
              :alt="`实况照片${carouselIndex + 1}`"
              loading="lazy"
              class="echoimg w-full h-auto"
            />
            <div class="livephoto-overlay">
              <LivePhotoIcon class="livephoto-icon" color="#ffffff" />
              <span class="livephoto-text">LIVE</span>
            </div>
          </button>
          
          <!-- 普通图片 -->
          <button
            v-else-if="!isVideo(visibleMediaItems[carouselIndex]!)"
            class="carousel-slide bg-transparent border-0 p-0 cursor-pointer w-full overflow-hidden"
            @click="openFancybox(carouselIndex)"
          >
            <img
              :src="getMediaUrlCompat(visibleMediaItems[carouselIndex]!)"
              :alt="`预览图片${carouselIndex + 1}`"
              loading="lazy"
              class="echoimg w-full h-auto"
            />
          </button>
          
          <!-- 视频 -->
          <button
            v-else
            class="video-preview carousel-slide bg-transparent border-0 p-0 cursor-pointer w-full relative"
            @click="openFancybox(carouselIndex)"
          >
            <video
              :src="getMediaUrlCompat(visibleMediaItems[carouselIndex]!) + '#t=0.1'"
              preload="metadata"
              muted
              playsinline
              class="echoimg video-thumb w-full h-auto"
            ></video>
            <div class="play-overlay">
              <Play class="play-icon" color="#ffffff" />
            </div>
          </button>
        </template>
      </div>

      <div
        v-if="visibleMediaItems.length > 1"
        class="carousel-nav mt-3 flex items-center justify-center gap-3 text-[var(--text-color-500)]"
      >
        <button
          class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
          @click="prevCarousel"
          :disabled="carouselIndex === 0"
        >
          <Prev class="w-5 h-5 text-[var(--text-color-600)]" />
        </button>
        <span class="text-sm"> {{ carouselIndex + 1 }} / {{ visibleMediaItems.length }} </span>
        <button
          class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
          @click="nextCarousel"
          :disabled="carouselIndex === visibleMediaItems.length - 1"
        >
          <Next class="w-5 h-5 text-[var(--text-color-600)]" />
        </button>
      </div>
    </div>

    <!-- 水平轮播布局 -->
    <div v-if="layout === ImageLayout.HORIZONTAL" class="imgwidth mx-auto mb-4">
      <div class="horizontal-scroll-container">
        <div class="horizontal-scroll-wrapper">
          <template v-for="(item, idx) in visibleMediaItems" :key="idx">
            <!-- 实况照片 -->
            <button
              v-if="isLivePhoto(item)"
              class="livephoto-preview horizontal-item bg-transparent rounded-lg border-0 p-0 cursor-pointer shrink-0 relative"
              @click="openFancybox(idx)"
            >
              <img
                :src="getMediaUrlCompat(item)"
                :alt="`实况照片${idx + 1}`"
                loading="lazy"
                class="echoimg h-full w-auto object-contain"
              />
              <div class="livephoto-overlay livephoto-overlay-small">
                <LivePhotoIcon class="livephoto-icon-small" color="#ffffff" />
              </div>
            </button>
            
            <!-- 普通图片 -->
            <button
              v-else-if="!isVideo(item)"
              class="horizontal-item bg-transparent rounded-lg border-0 p-0 cursor-pointer shrink-0"
              @click="openFancybox(idx)"
            >
              <img
                :src="getMediaUrlCompat(item)"
                :alt="`预览图片${idx + 1}`"
                loading="lazy"
                class="echoimg h-full w-auto object-contain"
              />
            </button>
            
            <!-- 视频 -->
            <button
              v-else
              class="video-preview horizontal-item bg-transparent rounded-lg border-0 p-0 cursor-pointer shrink-0 relative"
              @click="openFancybox(idx)"
            >
              <video
                :src="getMediaUrlCompat(item) + '#t=0.1'"
                preload="metadata"
                muted
                playsinline
                class="echoimg video-thumb h-full w-auto object-contain"
              ></video>
              <div class="play-overlay">
                <Play class="play-icon" color="#ffffff" />
              </div>
            </button>
          </template>
        </div>
      </div>
      <div class="scroll-hint">← 左右滑动查看更多 →</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from 'vue'
import { getMediaUrl, getHubMediaUrl, getImageUrl, getHubImageUrl } from '@/utils/other'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { ImageLayout } from '@/enums/enums'
import Prev from '@/components/icons/prev.vue'
import Next from '@/components/icons/next.vue'
import Play from '@/components/icons/play.vue'
import LivePhotoIcon from '@/components/icons/livephoto.vue'

const props = defineProps<{
  media?: App.Api.Ech0.Media[]
  images?: App.Api.Ech0.Image[]  // 向后兼容
  baseUrl?: string
  layout?: ImageLayout | string | undefined
}>()

// 使用 media 或 images（向后兼容）
const mediaItems = computed(() => props.media || props.images || [])

const baseUrl = props.baseUrl

// 布局状态（来自 props.layout）
const layout = props.layout || ImageLayout.WATERFALL

// 辅助函数：获取媒体URL（兼容新旧格式）
const getMediaUrlCompat = (item: any) => {
  // 如果有 media_url 字段，说明是新格式（Media）
  if ('media_url' in item) {
    return baseUrl ? getHubMediaUrl(item, baseUrl) : getMediaUrl(item)
  }
  // 否则是旧格式（Image）
  return baseUrl ? getHubImageUrl(item, baseUrl) : getImageUrl(item)
}

// 检查是否为视频
const isVideo = (item: any) => {
  return item.media_type === 'video'
}

// 检查是否为实况照片（通过 live_video_id 判断）
const isLivePhoto = (item: any) => {
  return item.live_video_id !== undefined && item.live_video_id > 0
}

// 获取实况照片的视频URL
const getLiveVideoUrl = (item: any) => {
  if (!isLivePhoto(item)) return null
  const video = mediaItems.value.find((m: any) => m.id === item.live_video_id)
  return video ? getMediaUrlCompat(video) : null
}

// 检查是否为实况照片的视频部分（应该隐藏）
const isLivePhotoVideo = (item: any) => {
  return mediaItems.value.some((m: any) => m.live_video_id === item.id)
}

// 过滤掉实况照片的视频部分，只显示图片部分
const visibleMediaItems = computed(() => {
  return mediaItems.value.filter((item: any) => !isLivePhotoVideo(item))
})

// 轮播索引
const carouselIndex = ref(0)

// 只显示前 9 张（用于九宫格），第 9 张显示 "+N" 覆盖层
const displayedImages = computed(() => mediaItems.value.slice(0, 9))
const extraCount = computed(() =>
  mediaItems.value.length > 9 ? mediaItems.value.length - 9 : 0
)

// 可见媒体的九宫格显示
const displayedVisibleImages = computed(() => visibleMediaItems.value.slice(0, 9))
const extraVisibleCount = computed(() =>
  visibleMediaItems.value.length > 9 ? visibleMediaItems.value.length - 9 : 0
)

// 瀑布流布局：获取列跨度
const getColSpan = (idx: number, total: number) => {
  if (total === 1) return 'col-span-1 justify-self-center'
  if (idx === 0 && total % 2 !== 0) return 'col-span-2'
  return ''
}

// 轮播导航
const prevCarousel = () => {
  if (carouselIndex.value > 0) carouselIndex.value--
}
const nextCarousel = () => {
  if (carouselIndex.value < visibleMediaItems.value.length - 1) carouselIndex.value++
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
    
    const leave = () => {
      container!.classList.remove('zoom')
      video.pause()
    }
    
    const handleVideoEnded = () => {
      container!.classList.remove('zoom')
    }
    
    // 鼠标事件绑定到 icon，触摸事件绑定到 image
    icon.addEventListener('mouseenter', start)
    icon.addEventListener('mouseleave', leave)
    image.addEventListener('touchstart', start)
    image.addEventListener('touchend', leave)
    image.addEventListener('touchcancel', leave)
    video.addEventListener('ended', handleVideoEnded)
    
    slide.livePhotoCleanup = () => {
      icon.removeEventListener('mouseenter', start)
      icon.removeEventListener('mouseleave', leave)
      image.removeEventListener('touchstart', start)
      image.removeEventListener('touchend', leave)
      image.removeEventListener('touchcancel', leave)
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

// 创建实况照片HTML内容（用于Fancybox）
// 参考实现：视频在下层，图片在上层，通过 CSS class 控制切换
function createLivePhotoHTML(imageUrl: string, videoUrl: string): string {
  return `
    <div class="fancybox-livephoto-container">
      <div class="livephoto-content">
        <video class="livephoto-video" src="${videoUrl}" preload="metadata" playsinline></video>
        <img class="livephoto-image" src="${imageUrl}" alt="实况照片" />
        <div class="livephoto-icon">
          <svg class="livephoto-icon-svg" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5" fill="none"/>
            <circle cx="12" cy="12" r="6" stroke="currentColor" stroke-width="1.5" fill="none"/>
            <circle cx="12" cy="12" r="3" fill="currentColor"/>
          </svg>
          <span class="livephoto-label">LIVE</span>
        </div>
      </div>
    </div>
  `
}

function openFancybox(startIndex: number) {
  // 处理所有可见媒体类型（图片、视频、实况照片）
  const items = visibleMediaItems.value.map((item: any) => {
    const mediaUrl = getMediaUrlCompat(item)
    
    if (isLivePhoto(item)) {
      // 为实况照片创建HTML内容
      const videoUrl = getLiveVideoUrl(item)
      if (videoUrl) {
        return {
          html: createLivePhotoHTML(mediaUrl, videoUrl),
          thumb: mediaUrl,
        }
      }
      // 如果没有找到视频，作为普通图片处理
      return {
        src: mediaUrl,
        type: 'image',
        thumb: mediaUrl,
      }
    } else if (isVideo(item)) {
      // 视频使用 Fancybox 原生支持
      return {
        src: mediaUrl,
        thumb: mediaUrl,
      }
    } else {
      // 普通图片
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
      // Fancybox 准备就绪后初始化实况照片交互
      ready: (fancybox: any) => {
        const carousel = fancybox.getCarousel()
        if (!carousel) return
        
        // 初始化所有实况照片幻灯片
        carousel.getSlides().forEach((slide: any) => {
          if (!slide.html && !slide.htmlEl) return
          
          const slideEl = slide.el || slide.htmlEl
          const isLivePhoto = slide.htmlEl?.classList?.contains('fancybox-livephoto-container') ||
            slideEl?.querySelector('.fancybox-livephoto-container')
          
          if (isLivePhoto) {
            initLivePhotoInteraction(slide)
          }
        })
        
        // 监听 Carousel 的 change 事件来处理幻灯片切换
        carousel.on('change', (carousel: any, to: number, from?: number) => {
          // 清理前一个幻灯片（暂停视频，但不移除事件监听器）
          if (from !== undefined) {
            const slides = carousel.getSlides()
            const prevSlide = slides[from]
            if (prevSlide) {
              // 只暂停视频，不清理事件监听器
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
          
          // 初始化新幻灯片（如果是实况照片且未初始化）
          const currentSlide = carousel.getSlides()[to]
          if (currentSlide && (currentSlide.html || currentSlide.htmlEl)) {
            setTimeout(() => {
              const slideEl = currentSlide.el || currentSlide.htmlEl
              const isLivePhoto = currentSlide.htmlEl?.classList?.contains('fancybox-livephoto-container') ||
                slideEl?.querySelector('.fancybox-livephoto-container')
              
              // 如果是实况照片且还没有初始化过，则初始化
              if (isLivePhoto && !currentSlide.livePhotoCleanup) {
                initLivePhotoInteraction(currentSlide)
              }
            }, 50)
          }
        })
      },
      // 当Fancybox关闭时清理资源
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

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
})

onBeforeUnmount(() => {
  // 清理所有可能残留的Fancybox实例
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
})
</script>

<style scoped>
.image-gallery-container {
  position: relative;
}

.imgwidth {
  width: 88%;
}

.echoimg {
  border-radius: 8px;
  box-shadow:
    0 1px 2px rgba(0, 0, 0, 0.02),
    0 2px 4px rgba(0, 0, 0, 0.02),
    0 4px 8px rgba(0, 0, 0, 0.02),
    0 8px 16px rgba(0, 0, 0, 0.02);
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}

/* 图片按钮悬停效果 */
button:has(img.echoimg):hover img.echoimg {
  transform: translateY(-2px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
}

/* 视频预览容器样式 - 与图片保持一致 */
.video-preview {
  position: relative;
  display: block;
  overflow: hidden;
  cursor: pointer;
}

/* 视频预览悬停效果 - 与图片一致 */
.video-preview:hover .video-thumb {
  transform: translateY(-2px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
}

.video-preview:hover .play-overlay {
  background: rgba(0, 0, 0, 0.65);
  transform: translate(-50%, -50%) scale(1.1);
}

.video-preview:active .video-thumb {
  transform: translateY(0);
}

.video-thumb {
  display: block;
  pointer-events: none;
  width: 100%;
  height: auto;
}

/* 播放图标覆盖层 */
.play-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  pointer-events: none;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.play-overlay::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.2);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.video-preview:hover .play-overlay::before {
  opacity: 1;
  animation: pulse-ring 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse-ring {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.5;
  }
}

.play-icon {
  width: 32px;
  height: 32px;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
  transition: transform 0.3s ease;
}

.video-preview:hover .play-icon {
  transform: scale(1.15);
}

/* 确保九宫格中视频和图片大小一致 */
.grid .video-preview {
  width: 100%;
  height: 100%;
}

.grid .video-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 水平轮播中的视频和图片保持一致 */
.horizontal-item {
  position: relative;
}

.horizontal-item .video-thumb,
.horizontal-item img {
  height: 100%;
  width: auto;
  object-fit: contain;
  display: block;
}

button:hover .echoimg {
  transform: scale(1.02);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
}

/* carousel, horizontal, grid styles (copied/adapted from provided template) */
.carousel-container {
  position: relative;
  width: 100%;
}
.carousel-slide {
  position: relative;
  width: 100%;
  display: block;
}

.horizontal-scroll-container {
  position: relative;
  width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.1) transparent;
}
.horizontal-scroll-container::-webkit-scrollbar {
  height: 4px;
}
.horizontal-scroll-wrapper {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  align-items: center;
}
.horizontal-item {
  flex-shrink: 0;
  height: 200px;
  width: auto;
  overflow: hidden;
  border-radius: 8px;
}
.scroll-hint {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-top: 8px;
  animation: hint-pulse 2s infinite;
}
@keyframes hint-pulse {
  0%,
  100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
}

.more-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  border-radius: 8px;
}

/* 超小屏幕优化 */
@media (max-width: 480px) {
  .play-overlay {
    width: 56px;
    height: 56px;
  }
  
  .play-icon {
    width: 28px;
    height: 28px;
  }
}

/* 实况照片预览样式 */
.livephoto-preview {
  position: relative;
  display: block;
  overflow: hidden;
  cursor: pointer;
}

.livephoto-preview:hover .echoimg {
  transform: translateY(-2px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
}

/* 实况照片指示器覆盖层 */
.livephoto-overlay {
  position: absolute;
  top: 8px;
  left: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 12px;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  pointer-events: none;
  transition: all 0.3s ease;
}

.livephoto-overlay-small {
  padding: 2px 6px;
  top: 4px;
  left: 4px;
}

.livephoto-preview:hover .livephoto-overlay {
  background: rgba(0, 0, 0, 0.7);
}

.livephoto-icon {
  width: 16px;
  height: 16px;
}

.livephoto-icon-small {
  width: 12px;
  height: 12px;
}

.livephoto-text {
  font-size: 10px;
  font-weight: 600;
  color: #fff;
  letter-spacing: 0.5px;
}

@media (max-width: 480px) {
  .livephoto-overlay {
    padding: 2px 6px;
    top: 4px;
    left: 4px;
  }
  
  .livephoto-icon {
    width: 12px;
    height: 12px;
  }
  
  .livephoto-text {
    font-size: 8px;
  }
}
</style>

<!-- Fancybox 全局样式（非 scoped，因为 Fancybox 内容渲染在 body 下） -->
<style>
/* Fancybox 实况照片容器样式 */
.fancybox-livephoto-container {
  position: relative;
  display: inline-block;
  border-radius: 8px;
  margin: 0 auto;
}

.fancybox-livephoto-container .livephoto-content {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 8px;
}

.fancybox-livephoto-container .livephoto-content img,
.fancybox-livephoto-container .livephoto-content video {
  display: block;
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
}

.fancybox-livephoto-container .livephoto-content video {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  opacity: 0;
  transition: opacity 0.5s ease, transform 0.5s ease;
}

/* 播放状态：显示视频 */
.fancybox-livephoto-container.zoom .livephoto-content video {
  opacity: 1;
}

.fancybox-livephoto-container .livephoto-content img {
  position: relative;
  z-index: 1;
  transition: opacity 0.5s ease, transform 0.5s ease;
}

/* 播放状态：zoom class */
.fancybox-livephoto-container.zoom .livephoto-content img,
.fancybox-livephoto-container.zoom .livephoto-content video {
  transform: scale(1.05);
}

.fancybox-livephoto-container.zoom .livephoto-content img {
  opacity: 0;
}

.fancybox-livephoto-container.zoom .livephoto-icon-svg {
  animation: livephoto-spin 3s linear infinite;
}

@keyframes livephoto-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 实况图标 - 与预览时保持完全一致的样式 */
.fancybox-livephoto-container .livephoto-content .livephoto-icon {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 12px;
  position: absolute;
  left: 8px;
  top: 8px;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  cursor: pointer;
  user-select: none;
  z-index: 10;
  transition: all 0.3s ease;
  pointer-events: auto;
}

.fancybox-livephoto-container .livephoto-content .livephoto-icon:hover {
  background: rgba(0, 0, 0, 0.7);
}

.fancybox-livephoto-container .livephoto-content .livephoto-icon-svg {
  width: 16px;
  height: 16px;
  color: #fff;
  flex-shrink: 0;
}

.fancybox-livephoto-container .livephoto-content .livephoto-label {
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .fancybox-livephoto-container {
    border-radius: 0;
  }
  
  .fancybox-livephoto-container .livephoto-content img,
  .fancybox-livephoto-container .livephoto-content video {
    max-height: 60vh;
  }
  
  /* 移动端保持与预览时一致的样式 */
  .fancybox-livephoto-container .livephoto-content .livephoto-icon {
    padding: 3px 6px;
  }
  
  .fancybox-livephoto-container .livephoto-content .livephoto-label {
    display: none;
  }
  
  .fancybox-livephoto-container .livephoto-content .livephoto-icon-svg {
    width: 14px;
    height: 14px;
  }
}

/* 超小屏幕优化 */
@media (max-width: 480px) {
  .fancybox-livephoto-container .livephoto-content img,
  .fancybox-livephoto-container .livephoto-content video {
    max-height: 50vh;
  }
  
  /* 超小屏幕保持与预览时一致的样式 */
  .fancybox-livephoto-container .livephoto-content .livephoto-icon {
    padding: 2px 6px;
  }
  
  .fancybox-livephoto-container .livephoto-content .livephoto-icon-svg {
    width: 12px;
    height: 12px;
  }
}
</style>
