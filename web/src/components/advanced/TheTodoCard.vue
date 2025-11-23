<template>
  <div>
    <div
      class="widget flex flex-col gap-2 p-4 rounded-md ring-1 ring-[var(--ring-color)] ring-inset mx-auto shadow-sm hover:shadow-md"
    >
      <!-- 顶部id + 按钮 -->
      <div v-if="props.operative" class="flex justify-between items-center">
        <!-- id -->
        <div class="flex justify-start gap-1 items-center h-auto font-bold text-2xl">
          <span class="italic text-[var(--text-color-next-300)]">#</span>
          <span class="text-[var(--text-color-next-400)]">{{ props.index }}</span>
        </div>
        <!-- 按钮 -->
        <div class="flex gap-2">
          <BaseButton
            :icon="Delete"
            @click="handleDeleteTodo"
            class="w-7 h-7 rounded-md text-red-200!"
            title="删除待办"
          />
          <BaseButton
            :icon="Done"
            @click="handleChangeTodoStatus"
            class="w-7 h-7 rounded-md"
            title="切换待办状态"
          />
        </div>
      </div>
      <div v-else>
        <p class="text-[var(--widget-title-color)] font-bold text-lg flex items-center">
          <Busy class="mr-1" /> 待办事项：
        </p>
      </div>
      <!-- 具体内容 -->
      <div v-if="!loading && props.todo">
        <p class="text-[var(--text-color-next-500)] text-sm whitespace-pre-wrap">
          {{ props.todo.content }}
        </p>
      </div>
      <div v-if="loading">
        <p class="text-[var(--text-color-next-500)] text-sm">加载中...</p>
      </div>
      <div v-if="!loading && !props.todo">
        <p class="text-[var(--text-color-next-500)] text-sm">今日无事🎉</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Done from '../icons/done.vue'
import Busy from '../icons/busy.vue'
import Delete from '../icons/delete.vue'
import BaseButton from '../common/BaseButton.vue'
import { fetchUpdateTodo, fetchDeleteTodo } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useTodoStore } from '@/stores/todo'
import { storeToRefs } from 'pinia'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()

const props = defineProps<{
  todo: App.Api.Todo.Todo | undefined
  index: number
  operative: boolean
}>()

const emit = defineEmits(['refresh'])

const todoStore = useTodoStore()
const { loading } = storeToRefs(todoStore)

const handleDeleteTodo = () => {
  openConfirm({
    title: '确定要删除待办吗？',
    description: '删除后将无法恢复，请谨慎操作',
    onConfirm: () => {
      if (props.todo?.id !== undefined) {
        fetchDeleteTodo(props.todo.id).then((res) => {
          if (res.code === 1) {
            theToast.success('待办已删除！')
            emit('refresh')
          }
        })
      }
    },
  })
}

const handleChangeTodoStatus = () => {
  if (props.todo?.id === undefined) {
    return
  }

  openConfirm({
    title: '确定要切换待办状态吗？',
    description: '切换后待办状态将标记为已完成',
    onConfirm: () => {
      fetchUpdateTodo(props.todo!.id).then((res) => {
        if (res.code === 1) {
          theToast.success('待办已完成！')
          emit('refresh')
        }
      })
    },
  })
}
</script>

<style scoped></style>
