<template>
  <div class="px-4 pb-4 py-2 mx-auto flex flex-col w-sm md:w-sm mt-4 mb-12">
    <h1
      class="text-4xl md:text-6xl italic font-bold font-serif text-center text-[var(--text-color-next-300)] mb-8"
    >
      Ech0 Widget
    </h1>

    <div class="px-8">
      <!-- 返回首页 -->
      <BaseButton @click="router.push('/')" :class="getButtonClasses('', true)" title="返回首页">
        <Arrow
          class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
        />
      </BaseButton>
    </div>

    <TheStatus />
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import TheStatus from './TheStatus.vue'
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const currentRoute = computed(() => route.name as string)

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
</script>
