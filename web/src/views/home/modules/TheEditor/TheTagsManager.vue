<template>
  <div class="py-4">
    <h2 class="text-[var(--text-color-500)] font-bold mb-2">标签管理</h2>
    <p class="text-xs text-[var(--text-color-next-300)] mb-3">Tip: 点击标签可以按标签过滤或删除</p>
    <div class="flex flex-wrap gap-2">
      <Popover v-for="tag in tagList" :key="tag.id" class="relative" v-slot="{ close }">
        <PopoverButton
          class="flex items-center gap-1 border rounded-sm border-gray-300 border-dashed py-0.5 px-1 mb-1 hover:bg-[var(--bg-color-50)] outline-none"
          style="white-space: nowrap"
        >
          <div
            class="hover:cursor-pointer text-[var(--text-color-400)] flex items-center justify-start gap-2"
          >
            <div>#</div>
            {{ tag.name }}
          </div>
        </PopoverButton>

        <transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="translate-y-1 opacity-0"
          enter-to-class="translate-y-0 opacity-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="translate-y-0 opacity-100"
          leave-to-class="translate-y-1 opacity-0"
        >
          <PopoverPanel class="absolute left-1/2 z-10 mt-1 -translate-x-1/2 transform">
            <div class="overflow-hidden rounded-lg shadow-lg ring-1 ring-black/5">
              <div
                class="relative flex justify-around gap-2 bg-[var(--bg-color-50)] p-1 text-[var(--text-color-500)]"
              >
                <button
                  @click="
                    () => {
                      handleFilterByTag(tag)
                      close()
                    }
                  "
                  title="按标签过滤内容"
                  class="flex items-center justify-center rounded-md p-1 transition duration-150 ease-in-out hover:bg-[var(--bg-color-100)] focus:outline-none focus-visible:ring focus-visible:ring-orange-500/50"
                >
                  <Filter class="w-5 h-5" />
                </button>
                <div class="w-px bg-[var(--bg-color-300)]"></div>
                <button
                  @click="
                    () => {
                      handleDeleteTag(tag.id)
                      close()
                    }
                  "
                  title="删除该标签"
                  class="flex items-center justify-center rounded-md p-1 transition duration-150 ease-in-out hover:bg-red-50 focus:outline-none focus-visible:ring focus-visible:ring-orange-500/50"
                >
                  <Trashbin class="w-5 h-5" />
                </button>
              </div>
            </div>
          </PopoverPanel>
        </transition>
      </Popover>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useEchoStore } from '@/stores'
import { fetchDeleteTagById } from '@/service/api'
import { storeToRefs } from 'pinia'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'
import Trashbin from '@/components/icons/trashbin.vue'
import Filter from '@/components/icons/filter.vue'

const echoStore = useEchoStore()
const { tagList } = storeToRefs(echoStore)

const { openConfirm } = useBaseDialog()

// 按标签过滤内容
const handleFilterByTag = (tag: App.Api.Ech0.Tag) => {
  const wasFiltering = echoStore.isFilteringMode

  // 重置状态
  echoStore.filteredTag = null
  echoStore.refreshEchosForFilter()

  if (tag) {
    // 开始过滤
    echoStore.filteredTag = tag
    echoStore.isFilteringMode = true

    // 如果已经在过滤模式下（组件已挂载），需要手动触发获取数据
    // 如果不是在过滤模式下，切换会导致组件挂载，onMounted会自动获取数据
    if (wasFiltering) {
      echoStore.getEchosByPageForFilter()
    }
  }
}

// 删除标签
const handleDeleteTag = (tagId: number) => {
  openConfirm({
    title: '确认删除该标签吗？',
    description: '删除后，所有使用该标签的内容将不再关联此标签',
    onConfirm: () => {
      fetchDeleteTagById(tagId).then((res) => {
        if (res.code === 1) {
          echoStore.getTags()
        }
      })
    },
  })
}
</script>

<style scoped></style>
