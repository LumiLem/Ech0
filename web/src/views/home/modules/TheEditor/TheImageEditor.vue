<template>
  <div>
    <h2 class="text-[var(--text-color-500)] font-bold my-2">插入图片/视频（支持直链、本地、S3存储）</h2>
    <div v-if="!MediaUploading" class="flex items-center gap-2 mb-3">
      <div class="flex items-center gap-2">
        <span class="text-[var(--text-color-500)]">选择添加方式：</span>
        <!-- 直链 -->
        <BaseButton
          :icon="Url"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetMediaSource(ImageSource.URL)"
          title="插入图片/视频链接"
        />
        <!-- 上传本地 -->
        <BaseButton
          :icon="Upload"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetMediaSource(ImageSource.LOCAL)"
          title="上传本地图片/视频"
        />
        <!-- S3 存储 -->
        <BaseButton
          v-if="S3Setting.enable"
          :icon="Bucket"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetMediaSource(ImageSource.S3)"
          title="S3存储图片/视频"
        />
      </div>
      <div>
        <BaseButton
          v-if="mediaToAdd.media_url != ''"
          :icon="Addmore"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="editorStore.handleAddMoreMedia"
          title="添加更多图片/视频"
        />
      </div>
    </div>

    <!-- 布局方式选择 -->
    <div class="mb-2 flex items-center gap-2">
      <span class="text-[var(--text-color-500)]">选择布局方式：</span>
      <BaseSelect
        v-model="echoToAdd.layout"
        :options="layoutOptions"
        class="w-32 h-7"
        placeholder="请选择布局方式"
      />
      <!-- AI 布局推荐按钮（手动触发，仅在非自动模式时显示） -->
      <BaseButton
        v-if="mediaListToAdd.length > 0 && echoToAdd.layout !== ImageLayout.AUTO"
        :icon="Magic"
        :class="`w-7 h-7 sm:w-7 sm:h-7 rounded-md ${isRecommending ? 'animate-pulse' : ''}`"
        :disabled="isRecommending"
        @click="handleAIRecommend"
        :title="isRecommending ? 'AI 正在分析...' : 'AI 智能推荐布局'"
      />
    </div>

    <!-- 智能压缩 -->
    <div v-if="mediaToAdd.media_source !== ImageSource.URL" class="mb-3 flex items-center">
      <span class="text-[var(--text-color-500)]">智能压缩：</span>
      <BaseSwitch v-model="enableCompressor" />
    </div>

    <!-- 当前上传方式与状态 -->
    <div class="text-[var(--text-color-300)] text-sm mb-2">
      当前上传方式为
      <span class="font-bold">
        {{
          mediaToAdd.media_source === ImageSource.URL
            ? '直链'
            : mediaToAdd.media_source === ImageSource.LOCAL
              ? '本地存储'
              : 'S3存储'
        }}</span
      >
      {{ !MediaUploading ? '' : '，正在上传中...' }}
    </div>

    <div class="my-1">
      <!-- 媒体上传 -->
      <TheUppy
        v-if="mediaToAdd.media_source !== ImageSource.URL"
        :TheImageSource="mediaToAdd.media_source"
        :EnableCompressor="enableCompressor"
      />

      <!-- 媒体直链 -->
      <div v-if="mediaToAdd.media_source === ImageSource.URL" class="flex items-center gap-2">
        <BaseInput
          v-model="mediaToAdd.media_url"
          class="rounded-lg h-auto flex-1"
          placeholder="请输入图片或视频链接..."
        />
        <BaseButton
          v-if="mediaToAdd.media_url != ''"
          :icon="Addmore"
          class="w-8 h-8 sm:w-8 sm:h-8 rounded-md shrink-0"
          @click="editorStore.handleAddMoreMedia"
          title="添加更多图片/视频"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ImageSource, ImageLayout } from '@/enums/enums'
import Url from '@/components/icons/url.vue'
import Upload from '@/components/icons/upload.vue'
import Bucket from '@/components/icons/bucket.vue'
import Addmore from '@/components/icons/addmore.vue'
import Magic from '@/components/icons/magic.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import TheUppy from '@/components/advanced/TheUppy.vue'
import { localStg } from '@/utils/storage'

const editorStore = useEditorStore()
const { mediaToAdd, MediaUploading, echoToAdd, mediaListToAdd } = storeToRefs(editorStore)
const settingStore = useSettingStore()
const { S3Setting } = storeToRefs(settingStore)
const enableCompressor = ref<boolean>(false)

// AI 布局推荐状态
const isRecommending = ref(false)

const handleSetMediaSource = (source: ImageSource) => {
  mediaToAdd.value.media_source = source

  // 记忆上传方式
  localStg.setItem('image_source', source)
}

// AI 智能推荐布局（手动点击按钮）
const handleAIRecommend = async () => {
  isRecommending.value = true
  try {
    await editorStore.doRecommendLayout(true) // 显示 toast 提示
  } finally {
    isRecommending.value = false
  }
}

// 布局选择
const layoutOptions = [
  { label: '🪄 自动', value: ImageLayout.AUTO },
  { label: '瀑布流', value: ImageLayout.WATERFALL },
  { label: '九宫格', value: ImageLayout.GRID },
  { label: '单图轮播', value: ImageLayout.CAROUSEL },
  { label: '水平轮播', value: ImageLayout.HORIZONTAL },
]
</script>

<style scoped></style>
