<template>
  <div v-if="visibleMediaItems.length" class="image-gallery-container">
    <!-- 瀑布流布局 -->
    <div
      v-if="layout === ImageLayout.WATERFALL"
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
            <LivePhotoIcon class="livephoto-preview-icon" color="#ffffff" />
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
      <div :class="['grid gap-2', gridColsClass]">
        <template v-for="(item, idx) in displayedVisibleImages" :key="idx">
          <!-- 实况照片 -->
          <button
            v-if="isLivePhoto(item)"
            :class="[
              'livephoto-preview bg-transparent border-0 p-0 cursor-pointer overflow-hidden relative',
              isSingleItem ? singleItemColSpan : 'aspect-square'
            ]"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaUrlCompat(item)"
              :alt="`实况照片${idx + 1}`"
              loading="lazy"
              :class="['echoimg', isSingleItem ? 'w-full h-auto' : 'w-full h-full object-cover']"
            />
            <div :class="['livephoto-overlay', isSingleItem ? '' : 'livephoto-overlay-small']">
              <LivePhotoIcon :class="isSingleItem ? 'livephoto-preview-icon' : 'livephoto-preview-icon-small'" color="#ffffff" />
              <span v-if="isSingleItem" class="livephoto-text">LIVE</span>
            </div>
            <div v-if="extraVisibleCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraVisibleCount }}
            </div>
          </button>
          
          <!-- 普通图片 -->
          <button
            v-else-if="!isVideo(item)"
            :class="[
              'bg-transparent border-0 p-0 cursor-pointer overflow-hidden relative',
              isSingleItem ? singleItemColSpan : 'aspect-square'
            ]"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaUrlCompat(item)"
              :alt="`预览图片${idx + 1}`"
              loading="lazy"
              :class="['echoimg', isSingleItem ? 'w-full h-auto' : 'w-full h-full object-cover']"
            />
            <div v-if="extraVisibleCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraVisibleCount }}
            </div>
          </button>
          
          <!-- 视频 -->
          <button
            v-else
            :class="[
              'video-preview bg-transparent border-0 p-0 cursor-pointer overflow-hidden relative',
              isSingleItem ? singleItemColSpan : 'aspect-square'
            ]"
            @click="openFancybox(idx)"
          >
            <video
              :src="getMediaUrlCompat(item) + '#t=0.1'"
              preload="metadata"
              muted
              playsinline
              :class="['echoimg video-thumb', isSingleItem ? 'w-full h-auto' : 'w-full h-full object-cover']"
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
              <LivePhotoIcon class="livephoto-preview-icon" color="#ffffff" />
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
                <LivePhotoIcon class="livephoto-preview-icon-small" color="#ffffff" />
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
import { ref, computed } from 'vue'
import { getMediaUrl, getHubMediaUrl, getImageUrl, getHubImageUrl } from '@/utils/other'
import { ImageLayout } from '@/enums/enums'
import Prev from '@/components/icons/prev.vue'
import Next from '@/components/icons/next.vue'
import Play from '@/components/icons/play.vue'
import LivePhotoIcon from '@/components/icons/livephoto.vue'
import { useMediaFancybox, type MediaItem } from '@/composables/useMediaFancybox'

const props = defineProps<{
  media?: App.Api.Ech0.Media[]
  images?: App.Api.Ech0.Image[]  // 向后兼容
  baseUrl?: string
  layout?: ImageLayout | string | undefined
}>()

// 使用 media 或 images（向后兼容）
const mediaItems = computed(() => (props.media || props.images || []) as MediaItem[])

const baseUrl = props.baseUrl

// 布局状态（来自 props.layout）
const layout = props.layout || ImageLayout.GRID

// 辅助函数：获取媒体URL（兼容新旧格式）
const getMediaUrlCompat = (item: any) => {
  // 如果有 media_url 字段，说明是新格式（Media）
  if ('media_url' in item) {
    return baseUrl ? getHubMediaUrl(item, baseUrl) : getMediaUrl(item)
  }
  // 否则是旧格式（Image）
  return baseUrl ? getHubImageUrl(item, baseUrl) : getImageUrl(item)
}

// 使用通用的媒体 Fancybox composable
const {
  openFancybox: openFancyboxBase,
  getVisibleMediaItems,
  isLivePhoto: isLivePhotoBase,
  isVideo: isVideoBase,
} = useMediaFancybox({
  getMediaUrl: getMediaUrlCompat,
})

// 包装函数，自动传入当前媒体列表
const isLivePhoto = (item: MediaItem) => isLivePhotoBase(item, mediaItems.value)
const isVideo = (item: MediaItem) => isVideoBase(item)

// 过滤掉实况照片的视频部分，只显示图片部分
const visibleMediaItems = computed(() => getVisibleMediaItems(mediaItems.value))

// 包装 openFancybox，传入当前媒体列表
const openFancybox = (startIndex: number) => {
  openFancyboxBase(mediaItems.value, startIndex)
}

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

// 九宫格布局：根据媒体数量动态调整列数
// 1张: 根据宽高比决定占几列
// 2张: 两列
// 3张: 三列
// 4张: 2x2
// 5-6张: 三列
// 7-9张: 三列
const gridColsClass = computed(() => {
  const count = displayedVisibleImages.value.length
  if (count === 1) return 'grid-cols-3'
  if (count === 2) return 'grid-cols-2'
  if (count === 4) return 'grid-cols-2'
  return 'grid-cols-3'
})

// 是否为单张媒体（用于九宫格布局中完整显示单张图片）
const isSingleItem = computed(() => displayedVisibleImages.value.length === 1)

// 获取单张媒体的宽高比类型
const singleItemAspectType = computed(() => {
  if (!isSingleItem.value) return 'square'
  const item = displayedVisibleImages.value[0] as any
  const width = item?.width || 0
  const height = item?.height || 0
  if (!width || !height) return 'square' // 无宽高信息时默认方图
  const ratio = width / height
  if (ratio > 1.2) return 'landscape' // 横图
  if (ratio < 0.8) return 'portrait'  // 竖图
  return 'square' // 方图
})

// 单张媒体的列跨度类
const singleItemColSpan = computed(() => {
  switch (singleItemAspectType.value) {
    case 'landscape': return 'col-span-3' // 横图占满
    case 'portrait': return 'col-span-1'  // 竖图占 1/3
    default: return 'col-span-2'          // 方图占 2/3
  }
})

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

.livephoto-preview-icon {
  width: 16px;
  height: 16px;
}

.livephoto-preview-icon-small {
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
  
  .livephoto-preview-icon {
    width: 12px;
    height: 12px;
  }
  
  .livephoto-text {
    font-size: 8px;
  }
}
</style>
