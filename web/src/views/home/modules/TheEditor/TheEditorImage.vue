<template>
  <!-- 媒体预览（九宫格布局） -->
  <div
    v-if="
      visibleMediaItems &&
      visibleMediaItems.length > 0 &&
      (currentMode === Mode.ECH0 || currentMode === Mode.Image)
    "
    class="w-5/6 mx-auto my-7"
  >
    <div class="grid grid-cols-3 gap-2">
      <template v-for="(item, idx) in visibleMediaItems" :key="item.media_url || idx">
        <div
          class="relative overflow-hidden aspect-square rounded-lg shadow-lg group media-item"
          :class="{ 'dragging': draggedIndex === idx }"
          :data-media-index="idx"
          :draggable="true"
          @dragstart="handleDragStart(idx, $event)"
          @dragover.prevent="handleDragOver(idx)"
          @drop="handleDrop(idx)"
          @dragend="handleDragEnd"
          @touchstart="handleTouchStart(idx, $event)"
          @touchmove="handleTouchMove"
          @touchend="handleTouchEnd"
        >
          <!-- 删除按钮 - 移动端始终显示，桌面端悬停显示 -->
          <button
            @click.stop="handleRemoveImage(idx)"
            class="delete-btn absolute top-1 right-1 bg-red-100 hover:bg-red-300 active:bg-red-400 text-[var(--text-color-600)] rounded-lg w-7 h-7 flex items-center justify-center shadow z-10 transition-all"
            title="移除媒体"
          >
            <Close class="w-3.5 h-3.5" />
          </button>

          <!-- 实况照片预览 -->
          <button
            v-if="isLivePhoto(item)"
            class="livephoto-preview w-full h-full bg-transparent border-0 p-0 cursor-pointer"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaToAddUrl(item)"
              alt="实况照片"
              class="w-full h-full object-cover"
              loading="lazy"
            />
            <!-- 实况照片指示器 -->
            <div class="livephoto-overlay">
              <LivePhotoIcon class="livephoto-icon" color="#ffffff" />
            </div>
          </button>

          <!-- 普通图片预览 -->
          <button
            v-else-if="item.media_type === 'image'"
            class="w-full h-full bg-transparent border-0 p-0 cursor-pointer"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaToAddUrl(item)"
              alt="图片"
              class="w-full h-full object-cover"
              loading="lazy"
            />
          </button>

          <!-- 视频预览 -->
          <button
            v-else-if="item.media_type === 'video'"
            class="video-preview w-full h-full bg-transparent border-0 p-0 cursor-pointer relative"
            @click="openFancybox(idx)"
          >
            <video
              :src="getMediaToAddUrl(item) + '#t=0.1'"
              preload="metadata"
              muted
              playsinline
              class="w-full h-full object-cover"
            ></video>
            <div class="play-overlay">
              <Play class="play-icon" color="#ffffff" />
            </div>
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { storeToRefs } from 'pinia'
import Close from '@/components/icons/close.vue'
import LivePhotoIcon from '@/components/icons/livephoto.vue'
import Play from '@/components/icons/play.vue'
import { getMediaToAddUrl } from '@/utils/other'
import { fetchDeleteMedia } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useEchoStore } from '@/stores/echo'
import { Mode } from '@/enums/enums'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { ImageSource } from '@/enums/enums'
import { useEditorStore } from '@/stores/editor'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()

const echoStore = useEchoStore()
const { echoToUpdate } = storeToRefs(echoStore)
const editorStore = useEditorStore()
const { mediaListToAdd: imagesToAdd, currentMode, isUpdateMode } = storeToRefs(editorStore)

// 拖放相关状态
const draggedIndex = ref<number | null>(null)
const touchStartIndex = ref<number | null>(null)
const touchStartX = ref<number>(0)
const touchStartY = ref<number>(0)
const isDragging = ref<boolean>(false) // 是否正在拖拽
const dragThreshold = 10 // 拖拽阈值（像素）

// 检查是否为实况照片
const isLivePhoto = (item: App.Api.Ech0.MediaToAdd) => {
  // 情况1: 已保存的实况照片 - 有 live_video_id
  if (item.live_video_id !== undefined && item.live_video_id > 0) {
    return true
  }
  
  // 情况2: 新上传的实况照片 - 图片有 live_pair_id，且存在相同 live_pair_id 的视频
  if (item.media_type === 'image' && item.live_pair_id) {
    return imagesToAdd.value.some((m: any) => 
      m.media_type === 'video' && m.live_pair_id === item.live_pair_id
    )
  }
  
  return false
}

// 检查是否为实况照片的视频部分（应该隐藏）
const isLivePhotoVideo = (item: App.Api.Ech0.MediaToAdd) => {
  // 只检查视频类型
  if (item.media_type !== 'video') {
    return false
  }
  
  // 情况1: 新上传的实况照片的视频 - 通过 live_pair_id 判断
  if (item.live_pair_id) {
    return imagesToAdd.value.some((m: any) => 
      m.media_type === 'image' && m.live_pair_id === item.live_pair_id
    )
  }
  
  // 情况2: 已保存的实况照片的视频 - 通过 id 和 live_video_id 判断
  if (item.id) {
    return imagesToAdd.value.some((m: any) => m.live_video_id === item.id)
  }
  
  return false
}

// 过滤掉实况照片的视频部分，只显示图片
const visibleMediaItems = computed(() => {
  return imagesToAdd.value.filter((item: any) => !isLivePhotoVideo(item))
})

// 拖放处理函数
const handleDragStart = (index: number, event: DragEvent) => {
  draggedIndex.value = index
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

const handleDragOver = (index: number) => {
  if (draggedIndex.value === null || draggedIndex.value === index) return
  
  // 交换位置
  const draggedItem = visibleMediaItems.value[draggedIndex.value]
  const targetItem = visibleMediaItems.value[index]
  
  if (!draggedItem || !targetItem) return
  
  // 在原始数组中找到对应的索引
  const draggedActualIndex = imagesToAdd.value.indexOf(draggedItem)
  const targetActualIndex = imagesToAdd.value.indexOf(targetItem)
  
  if (draggedActualIndex === -1 || targetActualIndex === -1) return
  
  // 交换数组中的元素
  const temp = imagesToAdd.value[draggedActualIndex]
  const targetValue = imagesToAdd.value[targetActualIndex]
  if (temp && targetValue) {
    imagesToAdd.value[draggedActualIndex] = targetValue
    imagesToAdd.value[targetActualIndex] = temp
  }
  
  // 更新拖拽索引
  draggedIndex.value = index
}

const handleDrop = (index: number) => {
  // 拖放完成，不需要额外操作
}

const handleDragEnd = () => {
  draggedIndex.value = null
}

// 触摸拖放处理（移动端）
const handleTouchStart = (index: number, event: TouchEvent) => {
  // 不立即阻止默认行为，让点击事件可以正常触发
  touchStartIndex.value = index
  isDragging.value = false
  
  const touch = event.touches[0]
  if (touch) {
    touchStartX.value = touch.clientX
    touchStartY.value = touch.clientY
  }
}

const handleTouchMove = (event: TouchEvent) => {
  if (touchStartIndex.value === null) return
  
  const touch = event.touches[0]
  if (!touch) return
  
  // 计算移动距离
  const deltaX = Math.abs(touch.clientX - touchStartX.value)
  const deltaY = Math.abs(touch.clientY - touchStartY.value)
  const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY)
  
  // 只有移动距离超过阈值时才认为是拖拽
  if (!isDragging.value && distance > dragThreshold) {
    isDragging.value = true
    draggedIndex.value = touchStartIndex.value
  }
  
  // 只有在拖拽状态下才阻止默认行为和执行拖拽逻辑
  if (isDragging.value) {
    event.preventDefault() // 防止页面滚动和长按菜单
    
    const element = document.elementFromPoint(touch.clientX, touch.clientY)
    
    // 找到目标媒体项
    const mediaItem = element?.closest('[data-media-index]')
    if (mediaItem) {
      const targetIndex = parseInt(mediaItem.getAttribute('data-media-index') || '-1')
      if (targetIndex >= 0 && targetIndex !== draggedIndex.value) {
        handleDragOver(targetIndex)
      }
    }
  }
}

const handleTouchEnd = (event: TouchEvent) => {
  // 只有在拖拽状态下才阻止默认行为
  if (isDragging.value) {
    event.preventDefault()
  }
  
  touchStartIndex.value = null
  draggedIndex.value = null
  isDragging.value = false
}

// 获取实况照片的视频URL
const getLiveVideoUrl = (item: App.Api.Ech0.MediaToAdd) => {
  if (!isLivePhoto(item)) return null
  
  // 情况1: 已保存的实况照片 - 通过 live_video_id 查找
  if (item.live_video_id) {
    const video = imagesToAdd.value.find(m => m.id === item.live_video_id)
    return video ? getMediaToAddUrl(video) : null
  }
  
  // 情况2: 新上传的实况照片 - 通过 live_pair_id 查找
  if (item.live_pair_id) {
    const video = imagesToAdd.value.find((m: any) => 
      m.media_type === 'video' && m.live_pair_id === item.live_pair_id
    )
    return video ? getMediaToAddUrl(video) : null
  }
  
  return null
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

// 打开 Fancybox 查看大图
const openFancybox = (startIndex: number) => {
  // 处理所有可见媒体类型（图片、视频、实况照片）
  const items = visibleMediaItems.value.map((item: any) => {
    const mediaUrl = getMediaToAddUrl(item)
    
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
    } else if (item.media_type === 'video') {
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
          const isLivePhotoSlide = slide.htmlEl?.classList?.contains('fancybox-livephoto-container') ||
            slideEl?.querySelector('.fancybox-livephoto-container')
          
          if (isLivePhotoSlide) {
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
              const isLivePhotoSlide = currentSlide.htmlEl?.classList?.contains('fancybox-livephoto-container') ||
                slideEl?.querySelector('.fancybox-livephoto-container')
              
              // 如果是实况照片且还没有初始化过，则初始化
              if (isLivePhotoSlide && !currentSlide.livePhotoCleanup) {
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

const handleRemoveImage = (visibleIndex: number) => {
  if (
    visibleIndex < 0 ||
    visibleIndex >= visibleMediaItems.value.length ||
    visibleMediaItems.value.length === 0
  ) {
    theToast.error('当前媒体索引无效，无法删除！')
    return
  }
  
  const currentItem = visibleMediaItems.value[visibleIndex]
  
  if (!currentItem) {
    theToast.error('无法找到要删除的媒体！')
    return
  }
  
  const actualIndex = imagesToAdd.value.indexOf(currentItem)
  
  if (actualIndex === -1) {
    theToast.error('无法找到要删除的媒体！')
    return
  }
  
  const isLive = isLivePhoto(currentItem)
  const mediaType = isLive ? '实况照片' : (currentItem?.media_type === 'video' ? '视频' : '图片')

  openConfirm({
    title: `确定要移除${mediaType}吗？`,
    description: isLive ? '删除实况照片将同时删除图片和视频' : '',
    onConfirm: () => {
      const deleteMediaFile = (item: App.Api.Ech0.MediaToAdd) => {
        const mediaToDel: App.Api.Ech0.MediaToDelete = {
          url: String(item.media_url),
          source: String(item.media_source),
          object_key: item.object_key,
        }

        if (mediaToDel.source === ImageSource.LOCAL || mediaToDel.source === ImageSource.S3) {
          return fetchDeleteMedia({
            url: mediaToDel.url,
            source: mediaToDel.source,
            object_key: mediaToDel.object_key,
          })
        }
        return Promise.resolve()
      }

      // 如果是实况照片，找到并删除关联的视频
      if (isLive) {
        let videoItem = null
        let videoIndex = -1
        
        // 情况1: 已保存的实况照片 - 通过 live_video_id 查找
        if (currentItem.live_video_id) {
          videoIndex = imagesToAdd.value.findIndex(m => m.id === currentItem.live_video_id)
          if (videoIndex >= 0) {
            videoItem = imagesToAdd.value[videoIndex]
          }
        }
        
        // 情况2: 新上传的实况照片 - 通过 live_pair_id 查找
        if (!videoItem && currentItem.live_pair_id) {
          videoIndex = imagesToAdd.value.findIndex((m: any) => 
            m.media_type === 'video' && m.live_pair_id === currentItem.live_pair_id
          )
          if (videoIndex >= 0) {
            videoItem = imagesToAdd.value[videoIndex]
          }
        }
        
        // 删除找到的视频
        if (videoItem && videoIndex >= 0) {
          deleteMediaFile(videoItem).then(() => {
            // 从数组中删除视频
            imagesToAdd.value.splice(videoIndex, 1)
          }).catch((err) => {
            console.error('删除实况照片视频失败:', err)
          })
        }
      }

      // 删除图片/视频文件
      deleteMediaFile(currentItem).then(() => {
        // 从数组中删除当前项（需要重新查找索引，因为可能已经删除了视频）
        const newIndex = imagesToAdd.value.indexOf(currentItem)
        if (newIndex >= 0) {
          imagesToAdd.value.splice(newIndex, 1)
        }

        // 如果删除成功且当前处于Echo更新模式，则需要立马执行更新（图片删除操作不可逆，需要立马更新确保后端数据同步）
        if (isUpdateMode.value && echoToUpdate.value) {
          editorStore.handleAddOrUpdateEcho(true)
        }
      }).catch((err) => {
        console.error('删除媒体文件失败:', err)
        // 即使删除失败也从数组中移除
        const newIndex = imagesToAdd.value.indexOf(currentItem)
        if (newIndex >= 0) {
          imagesToAdd.value.splice(newIndex, 1)
        }
      })
    },
  })
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
/* 实况照片指示器覆盖层 */
.livephoto-overlay {
  position: absolute;
  top: 4px;
  left: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 12px;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  pointer-events: none;
  z-index: 5;
}

.livephoto-icon {
  width: 12px;
  height: 12px;
}

/* 视频元素样式 */
.video-preview video {
  pointer-events: none; /* 禁用视频的鼠标事件，防止显示画中画等浏览器控制 */
}

/* 视频播放按钮覆盖层 */
.play-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  transition: all 0.3s ease;
  pointer-events: none;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.video-preview:hover .play-overlay {
  background: rgba(0, 0, 0, 0.7);
  transform: translate(-50%, -50%) scale(1.1);
}

.play-icon {
  width: 24px;
  height: 24px;
}

/* 媒体项样式 */
.media-item {
  cursor: grab;
  touch-action: none; /* 禁用浏览器默认的触摸行为 */
  user-select: none;
  -webkit-user-select: none;
  -webkit-touch-callout: none; /* 禁用 iOS 长按菜单 */
  -webkit-tap-highlight-color: transparent; /* 移除点击高亮 */
}

.media-item:active {
  cursor: grabbing;
}

/* 禁用媒体项内所有元素的长按菜单 */
.media-item * {
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

/* 拖拽中的样式 */
.media-item.dragging {
  opacity: 0.5;
  transform: scale(0.95);
  transition: all 0.2s ease;
}

/* 删除按钮 - 桌面端悬停显示 */
.delete-btn {
  opacity: 0;
  transform: scale(0.8);
}

.group:hover .delete-btn {
  opacity: 1;
  transform: scale(1);
}

/* 移动端：删除按钮始终显示 */
@media (max-width: 768px) {
  .delete-btn {
    opacity: 1;
    transform: scale(1);
  }
}

@media (max-width: 480px) {
  .play-overlay {
    width: 40px;
    height: 40px;
  }
  
  .play-icon {
    width: 20px;
    height: 20px;
  }
  
  .delete-btn {
    width: 8vw;
    height: 8vw;
    max-width: 32px;
    max-height: 32px;
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
