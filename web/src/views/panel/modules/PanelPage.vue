<template>
  <div
    class="border border-[var(--border-color-300)] p-4 mx-auto flex flex-col max-w-screen-lg rounded-md mt-2 sm:mt-4 mb-12"
  >
    <h1
      class="text-4xl md:text-6xl italic font-bold font-serif text-center text-[var(--text-color-next-300)] mb-8 md:mb-12"
    >
      Ech0 Panel
    </h1>

    <!-- 移动端选择器 -->
    <div class="md:hidden mb-6 px-2 flex justify-between items-center mb-3">
      <div class="w-1/2">
        <BaseSelect
          class="!focus:ring-0 h-9"
          v-model="selectedRoute"
          :options="routeOptions"
          placeholder="选择页面"
          @change="handleRouteChange"
        />
      </div>

      <div class="flex gap-2 items-center">
        <!-- 返回首页 -->
        <BaseButton
          :icon="BackHand"
          @click="router.push('/')"
          class="w-9 h-9 rounded-md"
          title="返回首页"
        >
        </BaseButton>
        <!-- 退出登录 -->
        <BaseButton
          v-if="userStore.isLogin"
          :icon="Logout"
          @click="handleLogout"
          class="w-9 h-9 rounded-md"
          title="退出"
        >
        </BaseButton>
        <!-- 登录 / 注册 -->
        <BaseButton
          v-else
          :icon="Auth"
          @click="router.push('/auth')"
          class="w-9 h-9 rounded-md"
          title="登录 / 注册"
        >
        </BaseButton>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="mx-auto flex my-4 w-full max-w-screen-lg rounded-md">
      <!-- 桌面端侧边栏 -->
      <div class="hidden md:flex flex-col gap-2 w-48 pr-8 shrink-0">
        <!-- 返回首页 -->
        <BaseButton @click="router.push('/')" :class="getButtonClasses('', true)" title="返回首页">
          <Arrow
            class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
          />
        </BaseButton>

        <div class="h-px bg-[var(--bg-color-300)] mx-2" />

        <!-- 仪表盘 -->
        <BaseButton
          :icon="Dashboard"
          @click="router.push('/panel/dashboard')"
          :class="getButtonClasses('panel-dashboard')"
          title="仪表盘"
        >
          仪表盘
        </BaseButton>

        <!-- 偏好设置 -->
        <BaseButton
          :icon="Setting"
          @click="router.push('/panel/setting')"
          :class="getButtonClasses('panel-setting')"
          title="偏好设置"
        >
          偏好设置
        </BaseButton>

        <!-- 用户中心 -->
        <BaseButton
          :icon="User"
          @click="router.push('/panel/user')"
          :class="getButtonClasses('panel-user')"
          title="用户中心"
        >
          用户中心
        </BaseButton>

        <!-- 数据管理 -->
        <BaseButton
          :icon="Storage"
          @click="router.push('/panel/storage')"
          :class="getButtonClasses('panel-storage')"
          title="数据管理"
        >
          数据管理
        </BaseButton>

        <!-- 单点登录 -->
        <BaseButton
          :icon="Sso"
          @click="router.push('/panel/sso')"
          :class="getButtonClasses('panel-sso')"
          title="单点登录"
        >
          单点登录
        </BaseButton>

        <!-- 功能扩展 -->
        <BaseButton
          :icon="Extension"
          @click="router.push('/panel/extension')"
          :class="getButtonClasses('panel-extension')"
          title="功能扩展"
        >
          功能扩展
        </BaseButton>

        <!-- 外部集成 -->
        <BaseButton
          :icon="Others"
          @click="router.push('/panel/advance')"
          :class="getButtonClasses('panel-advance')"
          title="外部集成"
        >
          外部集成
        </BaseButton>

        <div class="h-px bg-[var(--bg-color-300)] mx-2" />

        <!-- 退出登录 -->
        <BaseButton
          :icon="Logout"
          @click="handleLogout"
          :class="getBottomButtonClasses()"
          title="退出登录"
        >
          登出
        </BaseButton>

        <!-- 登录 / 注册 -->
        <BaseButton
          :icon="Auth"
          @click="router.push('/auth')"
          :class="getBottomButtonClasses()"
          title="登录 / 注册"
        >
          登录
        </BaseButton>

        <div class="text-[var(--text-color-next-300)] font-serif my-2 ml-3">
          Version: {{ settingStore.hello?.version }}
        </div>
      </div>

      <!-- 路由内容 -->
      <div class="flex-1 min-w-0">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import Arrow from '@/components/icons/arrow.vue'
import User from '@/components/icons/user.vue'
import Auth from '@/components/icons/auth.vue'
import BackHand from '@/components/icons/backhand.vue'
import Extension from '@/components/icons/extension.vue'
import Dashboard from '@/components/icons/dashboard.vue'
import Others from '@/components/icons/theothers.vue'
import Setting from '@/components/icons/setting.vue'
import Storage from '@/components/icons/storage.vue'
import Sso from '@/components/icons/sso.vue'
import Logout from '@/components/icons/logout.vue'
import { computed, ref, watch } from 'vue'
import { useUserStore } from '@/stores/user'
import { useSettingStore } from '@/stores/setting'
import { useRouter, useRoute } from 'vue-router'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()

const userStore = useUserStore()
const settingStore = useSettingStore()
const router = useRouter()
const route = useRoute()

const currentRoute = computed(() => route.name as string)
const selectedRoute = ref(route.path)

// 统一的按钮样式计算函数
const getButtonClasses = (routeName: string, isBackButton = false) => {
  const baseClasses = isBackButton
    ? 'text-[var(--text-color-next-600)] rounded-md transition-all duration-300 border-none !shadow-none !ring-0 hover:opacity-75 p-2 group bg-transparent'
    : 'flex items-center gap-2 pl-3 py-1 rounded-md transition-all duration-300 border-none !shadow-none !ring-0 justify-start bg-transparent'

  const activeClasses =
    currentRoute.value === routeName
      ? 'text-stone-800 bg-orange-200'
      : 'text-[var(--text-color-next-600)] hover:opacity-75'

  return `${baseClasses} ${activeClasses}`
}

// 底部按钮样式
const getBottomButtonClasses = () => {
  return 'flex items-center gap-2 pl-3 py-1 rounded-md transition-all duration-300 border-none !shadow-none !ring-0 text-[var(--text-color-600)] hover:opacity-75 justify-start bg-transparent'
}

// 路由选项
const routeOptions = [
  { label: '仪表盘', value: '/panel/dashboard' },
  { label: '偏好设置', value: '/panel/setting' },
  { label: '用户中心', value: '/panel/user' },
  { label: '数据管理', value: '/panel/storage' },
  { label: '单点登录', value: '/panel/sso' },
  { label: '功能扩展', value: '/panel/extension' },
  { label: '外部集成', value: '/panel/advance' },
]

// 监听路由变化，更新选择器
watch(
  () => route.path,
  (newPath) => {
    selectedRoute.value = newPath
  },
)

// 处理选择器变化
const handleRouteChange = () => {
  router.push(selectedRoute.value)
}

const handleLogout = () => {
  // 检查是否登录
  if (!userStore.isLogin) {
    theToast.info('当前未登录')
    return
  }

  // 弹出浏览器确认框
  openConfirm({
    title: '确定要退出登录吗？',
    description: '',
    onConfirm: () => {
      // 清除用户信息
      userStore.logout()
      // 跳转到首页
      router.push('/')
      theToast.success('已退出登录')
    },
  })
}
</script>
