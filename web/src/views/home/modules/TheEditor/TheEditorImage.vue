<template>
  <!-- 媒体预览（图片和视频） -->
  <div
    v-if="
      visibleMediaItems &&
      visibleMediaItems.length > 0 &&
      (currentMode === Mode.ECH0 || currentMode === Mode.Image)
    "
    class="relative rounded-lg shadow-lg w-5/6 mx-auto my-7"
  >
    <button
      @click="handleRemoveImage"
      class="absolute -top-3 -right-4 bg-red-100 hover:bg-red-300 text-[var(--text-color-600)] rounded-lg w-7 h-7 flex items-center justify-center shadow"
      title="移除媒体"
    >
      <Close class="w-4 h-4" />
    </button>
    <div class="rounded-lg overflow-hidden relative">
      <template v-for="(item, idx) in visibleMediaItems" :key="idx">
        <!-- 实况照片预览 -->
        <a
          v-if="isLivePhoto(item)"
          :href="getMediaToAddUrl(item)"
          data-fancybox="gallery"
          :data-thumb="getMediaToAddUrl(item)"
          :class="{ hidden: idx !== imageIndex }"
          class="relative block"
        >
          <img
            :src="getMediaToAddUrl(item)"
            alt="实况照片"
            class="max-w-full object-cover"
            loading="lazy"
          />
          <!-- 实况照片指示器 -->
          <div class="livephoto-overlay">
            <LivePhotoIcon class="livephoto-icon" color="#ffffff" />
          </div>
        </a>
        
        <!-- 普通图片预览 -->
        <a
          v-else-if="item.media_type === 'image'"
          :href="getMediaToAddUrl(item)"
          data-fancybox="gallery"
          :data-thumb="getMediaToAddUrl(item)"
          :class="{ hidden: idx !== imageIndex }"
        >
          <img
            :src="getMediaToAddUrl(item)"
            alt="Image"
            class="max-w-full object-cover"
            loading="lazy"
          />
        </a>
        
        <!-- 视频预览 -->
        <div
          v-else-if="item.media_type === 'video'"
          :class="{ hidden: idx !== imageIndex }"
        >
          <video
            :src="getMediaToAddUrl(item)"
            controls
            playsinline
            preload="metadata"
            class="max-w-full object-cover"
          >
            您的浏览器不支持视频播放
          </video>
        </div>
      </template>
    </div>
  </div>
  <!-- 媒体切换 -->
  <div v-if="visibleMediaItems.length > 1" class="flex items-center justify-center">
    <button @click="imageIndex = Math.max(imageIndex - 1, 0)">
      <Prev class="w-7 h-7" />
    </button>
    <span class="text-[var(--text-color-500)] text-sm mx-2">
      {{ imageIndex + 1 }} / {{ visibleMediaItems.length }}
    </span>
    <button @click="imageIndex = Math.min(imageIndex + 1, visibleMediaItems.length - 1)">
      <Next class="w-7 h-7" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { storeToRefs } from 'pinia'
import Next from '@/components/icons/next.vue'
import Prev from '@/components/icons/prev.vue'
import Close from '@/components/icons/close.vue'
import LivePhotoIcon from '@/components/icons/livephoto.vue'
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

// const images = defineModel<App.Api.Ech0.ImageToAdd[]>('imagesToAdd', { required: true })

// const { currentMode } = defineProps<{
//   currentMode: Mode
// }>()

// const emit = defineEmits(['handleAddorUpdateEcho'])

const imageIndex = ref<number>(0) // 临时图片索引变量
const echoStore = useEchoStore()
const { echoToUpdate } = storeToRefs(echoStore)
const editorStore = useEditorStore()
const { mediaListToAdd: imagesToAdd, currentMode, isUpdateMode } = storeToRefs(editorStore)

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

const handleRemoveImage = () => {
  if (
    imageIndex.value < 0 ||
    imageIndex.value >= visibleMediaItems.value.length ||
    visibleMediaItems.value.length === 0
  ) {
    theToast.error('当前媒体索引无效，无法删除！')
    return
  }
  
  const visibleIndex = imageIndex.value
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
      if (isLive && currentItem.live_video_id) {
        const videoIndex = imagesToAdd.value.findIndex(m => m.id === currentItem.live_video_id)
        if (videoIndex >= 0) {
          const videoItem = imagesToAdd.value[videoIndex]
          if (videoItem) {
            // 删除视频文件
            deleteMediaFile(videoItem).then(() => {
              // 从数组中删除视频
              imagesToAdd.value.splice(videoIndex, 1)
            }).catch((err) => {
              console.error('删除实况照片视频失败:', err)
            })
          }
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

      imageIndex.value = 0
    },
  })
}

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
})
</script>

<style scoped>
/* 实况照片指示器覆盖层 - 与 TheImageGallery 保持一致 */
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
  z-index: 10;
}

.livephoto-icon {
  width: 16px;
  height: 16px;
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
}
</style>
