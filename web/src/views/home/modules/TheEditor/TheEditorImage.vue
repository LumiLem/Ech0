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
            draggable="false"
            class="delete-btn absolute top-1 right-1 bg-red-100 hover:bg-red-300 active:bg-red-400 text-[var(--text-color-600)] rounded-lg w-7 h-7 flex items-center justify-center shadow z-10 transition-all"
            title="移除媒体"
          >
            <Close class="w-3.5 h-3.5" />
          </button>

          <!-- 实况照片预览 -->
          <div
            v-if="isLivePhoto(item)"
            class="livephoto-preview w-full h-full cursor-pointer"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaToAddUrl(item)"
              alt="实况照片"
              class="w-full h-full object-cover pointer-events-none"
              loading="lazy"
              draggable="false"
            />
            <!-- 实况照片指示器 -->
            <div class="livephoto-overlay">
              <LivePhotoIcon class="livephoto-icon" color="#ffffff" />
            </div>
          </div>

          <!-- 普通图片预览 -->
          <div
            v-else-if="item.media_type === 'image'"
            class="w-full h-full cursor-pointer"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaToAddUrl(item)"
              alt="图片"
              class="w-full h-full object-cover pointer-events-none"
              loading="lazy"
              draggable="false"
            />
          </div>

          <!-- 视频预览 -->
          <div
            v-else-if="item.media_type === 'video'"
            class="video-preview w-full h-full cursor-pointer relative"
            @click="openFancybox(idx)"
          >
            <video
              :src="getMediaToAddUrl(item) + '#t=0.1'"
              preload="metadata"
              muted
              playsinline
              draggable="false"
              class="w-full h-full object-cover pointer-events-none"
            ></video>
            <div class="play-overlay">
              <Play class="play-icon" color="#ffffff" />
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import Close from '@/components/icons/close.vue'
import LivePhotoIcon from '@/components/icons/livephoto.vue'
import Play from '@/components/icons/play.vue'
import { getMediaToAddUrl } from '@/utils/other'
import { fetchDeleteMedia } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useEchoStore } from '@/stores/echo'
import { Mode, ImageSource } from '@/enums/enums'
import { useEditorStore } from '@/stores/editor'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useMediaFancybox } from '@/composables/useMediaFancybox'

const { openConfirm } = useBaseDialog()

const echoStore = useEchoStore()
const { echoToUpdate } = storeToRefs(echoStore)
const editorStore = useEditorStore()
const { mediaListToAdd: imagesToAdd, currentMode, isUpdateMode } = storeToRefs(editorStore)

// 使用通用的媒体 Fancybox composable
const {
  openFancybox: openFancyboxBase,
  getVisibleMediaItems,
  isLivePhoto: isLivePhotoBase,
  isLivePhotoVideo: isLivePhotoVideoBase,
} = useMediaFancybox({
  getMediaUrl: (item) => getMediaToAddUrl(item as App.Api.Ech0.MediaToAdd),
})

// 包装函数，自动传入当前媒体列表
const isLivePhoto = (item: App.Api.Ech0.MediaToAdd) => isLivePhotoBase(item, imagesToAdd.value)
const isLivePhotoVideo = (item: App.Api.Ech0.MediaToAdd) => isLivePhotoVideoBase(item, imagesToAdd.value)

// 过滤掉实况照片的视频部分，只显示图片
const visibleMediaItems = computed(() => 
  getVisibleMediaItems(imagesToAdd.value) as App.Api.Ech0.MediaToAdd[]
)

// 包装 openFancybox，传入当前媒体列表
const openFancybox = (startIndex: number) => {
  openFancyboxBase(imagesToAdd.value, startIndex)
}

// 拖放相关状态
const draggedIndex = ref<number | null>(null)
const touchStartIndex = ref<number | null>(null)
const touchStartX = ref<number>(0)
const touchStartY = ref<number>(0)
const isDragging = ref<boolean>(false) // 是否正在拖拽
const dragThreshold = 10 // 拖拽阈值（像素）

// 拖放处理函数
const handleDragStart = (index: number, event: DragEvent) => {
  draggedIndex.value = index
  draggedItemRef.value = visibleMediaItems.value[index] || null
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

// 获取实况照片关联的视频项
const getLivePhotoVideo = (item: App.Api.Ech0.MediaToAdd) => {
  if (!isLivePhoto(item)) return null
  
  // 情况1: 已保存的实况照片 - 通过 live_video_id 查找
  if (item.live_video_id) {
    return imagesToAdd.value.find(m => m.id === item.live_video_id) || null
  }
  
  // 情况2: 新上传的实况照片 - 通过 live_pair_id 查找
  if (item.live_pair_id) {
    return imagesToAdd.value.find((m: any) => 
      m.media_type === 'video' && m.live_pair_id === item.live_pair_id
    ) || null
  }
  
  return null
}

// 记录拖拽开始时的项引用，避免索引错乱
const draggedItemRef = ref<App.Api.Ech0.MediaToAdd | null>(null)

const handleDragOver = (index: number) => {
  if (draggedItemRef.value === null) return
  
  const draggedItem = draggedItemRef.value // 保存引用，确保类型收窄
  const targetItem = visibleMediaItems.value[index]
  if (!targetItem || targetItem === draggedItem) return
  
  // 在原始数组中找到对应的索引
  const draggedActualIndex = imagesToAdd.value.indexOf(draggedItem)
  const targetActualIndex = imagesToAdd.value.indexOf(targetItem)
  
  if (draggedActualIndex === -1 || targetActualIndex === -1) return
  
  // 获取实况照片关联的视频
  const draggedVideo = getLivePhotoVideo(draggedItem)
  const targetVideo = getLivePhotoVideo(targetItem)
  
  // 收集需要移动的项（保持图片和视频的相对顺序）
  const getItemsWithVideo = (item: App.Api.Ech0.MediaToAdd, video: App.Api.Ech0.MediaToAdd | null) => {
    if (!video) return [item]
    const itemIdx = imagesToAdd.value.indexOf(item)
    const videoIdx = imagesToAdd.value.indexOf(video)
    // 保持原有顺序
    return itemIdx < videoIdx ? [item, video] : [video, item]
  }
  
  const draggedItems = getItemsWithVideo(draggedItem, draggedVideo)
  const targetItems = getItemsWithVideo(targetItem, targetVideo)
  
  // 从数组中移除所有相关项
  const allItemsToRemove = new Set([...draggedItems, ...targetItems])
  const remainingItems = imagesToAdd.value.filter(item => !allItemsToRemove.has(item))
  
  // 找到目标位置在剩余数组中应该插入的位置
  // 通过找到目标项原本后面第一个不被移除的项来确定
  let insertIndex = remainingItems.length // 默认插入到末尾
  for (let i = targetActualIndex + 1; i < imagesToAdd.value.length; i++) {
    const item = imagesToAdd.value[i]
    if (item && !allItemsToRemove.has(item)) {
      insertIndex = remainingItems.indexOf(item)
      break
    }
  }
  
  // 根据拖拽方向决定插入顺序
  const isDraggingForward = draggedActualIndex < targetActualIndex
  
  // 重新构建数组
  const newArray = [...remainingItems]
  if (isDraggingForward) {
    // 向后拖：目标项在前，拖拽项在后
    newArray.splice(insertIndex, 0, ...targetItems, ...draggedItems)
  } else {
    // 向前拖：拖拽项在前，目标项在后
    newArray.splice(insertIndex, 0, ...draggedItems, ...targetItems)
  }
  
  // 更新数组
  imagesToAdd.value.splice(0, imagesToAdd.value.length, ...newArray)
  
  // 更新可见索引（用于样式）
  draggedIndex.value = visibleMediaItems.value.indexOf(draggedItem)
}

const handleDrop = (index: number) => {
  // 拖放完成，不需要额外操作
}

const handleDragEnd = () => {
  draggedIndex.value = null
  draggedItemRef.value = null
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
    draggedItemRef.value = visibleMediaItems.value[touchStartIndex.value] || null
  }
  
  // 只有在拖拽状态下才阻止默认行为和执行拖拽逻辑
  if (isDragging.value) {
    event.preventDefault() // 防止页面滚动和长按菜单
    
    const element = document.elementFromPoint(touch.clientX, touch.clientY)
    
    // 找到目标媒体项
    const mediaItem = element?.closest('[data-media-index]')
    if (mediaItem) {
      const targetIndex = parseInt(mediaItem.getAttribute('data-media-index') || '-1')
      if (targetIndex >= 0) {
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
  draggedItemRef.value = null
  isDragging.value = false
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
