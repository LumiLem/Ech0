<template>
  <div class="mx-auto mb-2">
    <div class="w-full flex justify-between items-center">
      <!-- 搜索与过滤 -->
      <div class="flex justify-start items-center gap-2">
        <BaseInput
          v-if="!isFilteringMode"
          title="搜索"
          type="text"
          v-model="searchContent"
          placeholder="搜索..."
          class="w-42 h-10 bg-[var(--input-bg-color)]"
          @keyup.enter="$event.target.blur()"
          @blur="handleSearch"
        />
        <!-- 过滤条件 -->
        <Filter v-if="isFilteringMode" class="w-7 h-7" />
        <!-- 标签筛选 -->
        <div
          v-if="isFilteringMode && filteredTag"
          @click="handleCancelFilter"
          class="w-34 text-nowrap flex items-center justify-between px-1 py-0.5 text-[var(--text-color-300)] border border-dashed border-[var(--border-color-400)] rounded-md hover:cursor-pointer hover:line-through hover:text-[var(--text-color-500)]"
        >
          <p class="text-nowrap truncate">{{ filteredTag.name }}</p>
          <Close class="inline w-4 h-4 ml-1" />
        </div>
        <!-- 日期筛选 -->
        <div
          v-if="isFilteringMode && (filteredDate || filteredYearMonth)"
          @click="handleCancelFilter"
          class="w-34 text-nowrap flex items-center justify-between px-1 py-0.5 text-[var(--text-color-300)] border border-dashed border-[var(--border-color-400)] rounded-md hover:cursor-pointer hover:line-through hover:text-[var(--text-color-500)]"
        >
          <p class="text-nowrap truncate">{{ filterDateLabel }}</p>
          <Close class="inline w-4 h-4 ml-1" />
        </div>
      </div>

      <!-- 右侧图标组 -->
      <div class="flex justify-end items-center gap-1">
        <!-- RSS -->
        <div>
          <a href="/rss" title="RSS">
            <Rss class="w-8 h-8 text-[var(--text-color-400)]" />
          </a>
        </div>
        <!-- Ech0 Hub -->
        <div class="relative">
          <a v-if="isInAppBrowser" href="/hub" :title="hubUpdateTooltip || 'Ech0 Hub'" @click="handleHubClick">
            <HubIcon class="w-8 h-8 text-[var(--text-color-400)]" />
          </a>
          <RouterLink v-else to="/hub" :title="hubUpdateTooltip || 'Ech0 Hub'" @click="handleHubClick">
            <HubIcon class="w-8 h-8 text-[var(--text-color-400)]" />
          </RouterLink>
          <!-- 更新提示红点 -->
          <div
            v-if="hubUpdateCount > 0"
            class="absolute -top-1 -right-1 min-w-[18px] h-[18px] bg-red-500 text-white text-xs rounded-full flex items-center justify-center px-1 font-bold"
          >
            {{ hubUpdateCount > 99 ? '99+' : hubUpdateCount }}
          </div>
          <!-- 移动端详情气泡 (短暂显示) -->
          <Transition name="hint">
            <div
              v-if="showMobileHint && hubUpdateCount > 0"
              class="absolute top-10 left-1/2 -translate-x-1/2 bg-gray-800/95 backdrop-blur-sm text-white text-[11px] px-3 py-1.5 rounded-lg shadow-2xl whitespace-nowrap z-50 pointer-events-none md:hidden border border-white/10"
            >
              {{ hubUpdateTooltip }}
              <!-- 小三角 (水平居中对齐图标) -->
              <div class="absolute -top-1 left-1/2 -translate-x-1/2 w-2 h-2 bg-gray-800 rotate-45 border-t border-l border-white/10"></div>
            </div>
          </Transition>
        </div>
        <!-- Ech0 Widget -->
        <div class="block xl:hidden">
          <a v-if="isInAppBrowser" href="/widget" title="Ech0 Widget">
            <Widget :class="widgetIconClass" />
          </a>
          <RouterLink v-else to="/widget" title="Ech0 Widget">
            <Widget :class="widgetIconClass" />
          </RouterLink>
        </div>
        <!-- PanelPage -->
        <div>
          <a v-if="isInAppBrowser" href="/panel" title="面板">
            <Panel class="w-8 h-8 text-[var(--text-color-400)]" />
          </a>
          <RouterLink v-else to="/panel" title="面板">
            <Panel class="w-8 h-8 text-[var(--text-color-400)]" />
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import Panel from '@/components/icons/panel.vue'
import Rss from '@/components/icons/rss.vue'
import HubIcon from '@/components/icons/hub.vue'
import { RouterLink } from 'vue-router'
import { useEchoStore } from '@/stores/echo'
import { useTodoStore } from '@/stores/todo'
import { useConnectStore } from '@/stores/connect'
import { storeToRefs } from 'pinia'
import { useWindowFocus } from '@vueuse/core'
import { ref, onMounted, computed, watch, onBeforeUnmount } from 'vue'

const isFocused = useWindowFocus()
import Close from '@/components/icons/close.vue'
import Filter from '@/components/icons/filter.vue'
import Widget from '@/components/icons/widget.vue'

// 检测是否在 QQ/微信等内置浏览器中
const isInAppBrowser = ref(false)

// 在客户端挂载后检测
onMounted(() => {
  const ua = navigator.userAgent.toLowerCase()
  isInAppBrowser.value = ua.includes('qq') || ua.includes('micromessenger') || ua.includes('weibo')
  
  // 开始轮询 Hub 更新
  startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
})

const echoStore = useEchoStore()
const todoStore = useTodoStore()
const connectStore = useConnectStore()
const { refreshForSearch, getEchosByPage } = echoStore
const { clearHubUpdates } = connectStore
const { searchingMode, filteredTag, isFilteringMode, filteredDate, filteredYearMonth, isDateFilteringMode } = storeToRefs(echoStore)
const { todos } = storeToRefs(todoStore)
const { hubUpdateCount, hubUpdateTooltip } = storeToRefs(connectStore)
const { startPolling, stopPolling, getConnectInfo } = connectStore

// 监听窗口焦点，切回来时尝试刷新（受缓存策略保护）
watch(isFocused, (focused) => {
  if (focused) {
    getConnectInfo()
  }
})

const showMobileHint = ref(false)

// 监听更新数量变化，显示移动端提示
watch(hubUpdateCount, (newCount) => {
  if (newCount > 0) {
    showMobileHint.value = true
    setTimeout(() => {
      showMobileHint.value = false
    }, 5000) // 5秒后消失
  }
}, { immediate: true })

const searchContent = ref<string>('')

// 检查是否有未完成的待办事项
const hasIncompleteTodos = computed(() => {
  return todos.value.some(todo => todo.status === 0)
})

// 日期筛选显示标签
const filterDateLabel = computed(() => {
  if (filteredDate.value) {
    return filteredDate.value
  }
  if (filteredYearMonth.value) {
    return `${filteredYearMonth.value.year}年${filteredYearMonth.value.month}月`
  }
  return ''
})

// 动态计算Widget图标的样式类
const widgetIconClass = computed(() => {
  const baseClass = 'w-8 h-8 text-[var(--text-color-400)]'
  if (hasIncompleteTodos.value) {
    return `${baseClass} animate-pulse`
  }
  return baseClass
})

// 处理Hub点击，清除更新提示
const handleHubClick = () => {
  clearHubUpdates()
}

const handleSearch = () => {
  console.log('搜索内容:', searchContent.value)

  // 设置搜索内容
  echoStore.searchValue = searchContent.value

  // 判断是否是搜索模式
  if (searchingMode.value) {
    // 初始化搜索
    refreshForSearch()
    // 开始搜索
    getEchosByPage()
  }
}

const handleCancelFilter = () => {
  echoStore.isFilteringMode = false
  echoStore.isDateFilteringMode = false
  echoStore.filteredTag = null
  echoStore.filteredDate = null
  echoStore.filteredYearMonth = null
  echoStore.refreshEchosForFilter()
}
</script>

<style scoped>
.hint-enter-active,
.hint-leave-active {
  transition: all 0.3s ease;
}

.hint-enter-from,
.hint-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
