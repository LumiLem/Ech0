<template>
  <div v-show="shouldShowComment" id="comments" class="max-w-sm h-auto p-4 my-4 mx-auto">
    <div id="comment"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

declare global {
  // @ts-nocheck
  /* eslint-disable */
  interface Window {
    twikoo: any
  }
}

const { CommentSetting, loading } = storeToRefs(useSettingStore())
const isScriptLoaded = ref<boolean>(false)
const isInitialized = ref<boolean>(false)

// 计算是否应该显示评论
const shouldShowComment = computed(() => {
  return CommentSetting.value?.enable_comment
})

// 加载 twikoo 脚本
function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    // 检查脚本是否已经存在
    const existingScript = document.querySelector(`script[src="${src}"]`)
    if (existingScript) {
      resolve()
      return
    }

    const script = document.createElement('script')
    script.src = src
    script.onload = () => {
      isScriptLoaded.value = true
      resolve()
    }
    script.onerror = (error) => {
      console.error('Failed to load Twikoo script:', error)
      reject(error)
    }
    document.head.appendChild(script)
  })
}

// 初始化 Twikoo
async function initializeTwikoo() {
  if (isInitialized.value || !shouldShowComment.value) {
    return
  }

  try {
    // 确保脚本已加载
    if (!isScriptLoaded.value) {
      await loadScript('/others/scripts/twikoo.all.min.js')
    }

    // 等待 DOM 更新
    await nextTick()

    // 检查 Twikoo 是否可用
    if (!window.twikoo) {
      throw new Error('Twikoo is not available')
    }

    // 初始化 Twikoo
    window.twikoo.init({
      envId: CommentSetting.value.comment_api,
      el: '#comment',
    })

    isInitialized.value = true
  } catch (error) {
    console.error('Failed to initialize Twikoo:', error)
  }
}

// 监听加载状态和评论设置变化
watch(
  [() => loading.value, () => shouldShowComment.value],
  async ([isLoading, showComment]) => {
    // 当加载完成且应该显示评论时，初始化 Twikoo
    if (!isLoading && showComment) {
      await initializeTwikoo()
    }
  },
  { immediate: true },
)

// 当评论设置变化时重新初始化
watch(
  () => CommentSetting.value?.comment_api,
  async (newApiUrl, oldApiUrl) => {
    if (newApiUrl && newApiUrl !== oldApiUrl && shouldShowComment.value) {
      isInitialized.value = false
      await initializeTwikoo()
    }
  },
)
</script>

<style>
.twikoo {
  width: 100% !important;
  max-width: 500px !important;
}
.tk-meta-input {
  flex-direction: column !important;
  padding: 1px 3px !important;
}

.tk-meta-input .el-input {
  width: 100% !important;
  flex: 1 !important;
  margin: 4px 0px !important;
}

.tk-meta-input .el-input + .el-input {
  margin-left: 0 !important;
}

.el-textarea {
  position: relative;
  display: inline-block;
  width: 100%;
  vertical-align: bottom;
  font-size: 14px;
  padding: 0px 2px !important;
}

.tk-comments-container {
  width: 100% !important;
  max-width: 500px !important;
  padding: 0 0.4rem !important;
  min-height: 10rem;
  display: flex;
  flex-direction: column;
}
</style>
