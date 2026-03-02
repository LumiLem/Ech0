<template>
  <div class="flex justify-between items-center py-1 px-3">
    <!-- 未登录状态：显示站点Logo和服务名称 -->
    <div v-if="!isLogin" class="flex flex-row items-center gap-2 justify-between">
      <button
        type="button"
        class="inline-flex rounded-full transition-transform duration-200 hover:scale-105 active:scale-95"
        :title="isZenMode ? '退出 Zen Mode' : '进入 Zen Mode'"
        :aria-pressed="isZenMode"
        @click="handleZenModeToggle"
      >
        <img
          :src="siteLogo"
          alt="站点Logo"
          loading="lazy"
          class="w-6 sm:w-7 h-6 sm:h-7 rounded-lg ring-1 ring-[var(--ring-color)] shadow-sm object-cover"
          @error="handleImageError"
        />
      </button>
      <h1 class="text-[var(--editor-title-color)] font-bold sm:text-xl">
        {{ SystemSetting.server_name }}
      </h1>
    </div>

    <!-- 已登录状态：显示用户头像和用户名 -->
    <div v-else class="flex flex-row items-center gap-2 justify-between">
      <button
        type="button"
        class="inline-flex rounded-full transition-transform duration-200 hover:scale-105 active:scale-95"
        :title="isZenMode ? '退出 Zen Mode' : '进入 Zen Mode'"
        :aria-pressed="isZenMode"
        @click="handleZenModeToggle"
      >
        <img
          :src="userAvatar"
          alt="用户头像"
          class="w-6 sm:w-7 h-6 sm:h-7 rounded-full ring-1 ring-gray-200 shadow-sm object-cover"
          @error="handleImageError"
        />
      </button>
      <h1 class="text-[var(--editor-title-color)] font-bold sm:text-xl">
        {{ user?.username }}
      </h1>
    </div>

    <div class="flex flex-row items-center gap-2">
      <!-- Hello -->
      <div
        class="p-1 ring-1 ring-inset ring-[var(--ring-color)] rounded-full transition-colors duration-200 cursor-pointer"
      >
        <Hello @click="handleHello" class="w-6 h-6" />
      </div>
      <!-- Github -->
      <!--
      <div>
        <a href="https://github.com/LumiLem/Ech0" target="_blank" title="Github">
          <Github class="w-6 sm:w-7 h-6 sm:h-7 text-[var(--text-color-400)]" />
        </a>
      </div>
      -->
    </div>
  </div>
</template>

<script setup lang="ts">
import Hello from '@/components/icons/hello.vue'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref } from 'vue'
import { fetchHelloEch0 } from '@/service/api'
import { useSettingStore, useThemeStore, useUserStore, useZenStore } from '@/stores'
import { getApiUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'

const settingStore = useSettingStore()
const userStore = useUserStore()
const themeStore = useThemeStore()
const zenStore = useZenStore()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)
const { isZenMode } = storeToRefs(zenStore)

const apiUrl = getApiUrl()

// 站点Logo（未登录时显示）
const siteLogo = computed(() => {
  const logo = SystemSetting.value.server_logo
  if (!logo || logo.length === 0) {
    return '/Ech0.svg'
  }
  return logo.startsWith('http') ? logo : `${apiUrl}${logo}`
})

// 用户头像（已登录时显示）
const userAvatar = computed(() => {
  const avatar = user.value?.avatar
  if (!avatar || avatar.length === 0) {
    return '/Ech0.svg'
  }
  return avatar.startsWith('http') ? avatar : `${apiUrl}${avatar}`
})

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src = '/Ech0.svg'
}

const handleHello = async (event: MouseEvent) => {
  await themeStore.toggleTheme(event)

  // 在主题切换完成后获取正确的模式
  const modeText =
    themeStore.mode === 'system' ? 'Auto' : themeStore.mode === 'light' ? 'Light' : 'Dark'

  const hello = ref<App.Api.Ech0.HelloEch0>()

  fetchHelloEch0().then((res) => {
    if (res.code === 1) {
      hello.value = res.data
      theToast.success('你好呀！ 👋', {
        description: `当前版本：v${hello.value.version} | ${modeText}`,
        duration: 2000,
        action: {
          label: 'Github',
          onClick: () => {
            window.open(hello.value?.github, '_blank')
          },
        },
      })
    }
  })
}

const handleZenModeToggle = async () => {
  await zenStore.toggleZenMode()
}

onMounted(() => {
  // 获取系统设置
  settingStore.getSystemSetting()
})
</script>

<style scoped></style>
