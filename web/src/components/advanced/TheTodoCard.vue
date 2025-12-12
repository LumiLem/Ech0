<template>
  <div>
    <div
      class="widget flex flex-col gap-2 p-4 rounded-md ring-1 ring-[var(--ring-color)] ring-inset mx-auto shadow-sm hover:shadow-md"
    >
      <!-- 操作模式 -->
      <div v-if="props.operative">
        <!-- 顶部id + 按钮 -->
        <div class="flex justify-between items-center">
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
        
        <!-- 具体内容 -->
        <div v-if="!loading && props.todo" class="mt-3">
          <p class="text-[var(--text-color-next-500)] text-sm whitespace-pre-wrap">
            {{ props.todo.content }}
          </p>
        </div>
        <div v-if="loading" class="mt-3">
          <p class="text-[var(--text-color-next-500)] text-sm">加载中...</p>
        </div>
        <div v-if="!loading && !props.todo" class="mt-3">
          <p class="text-[var(--text-color-next-500)] text-sm">今日无事🎉</p>
        </div>
      </div>
      <div v-else>
        <p class="text-[var(--widget-title-color)] font-bold text-lg flex items-center mb-3">
          <Busy class="mr-1" /> 待办事项：
        </p>
        
        <!-- 加载状态 -->
        <div v-if="loading" class="text-[var(--text-color-next-500)] text-sm">
          加载中...
        </div>
        
        <!-- 待办列表 -->
        <div v-else-if="todos && todos.length > 0" class="space-y-1">
          <transition-group name="todo-item" tag="div">
            <div 
              v-for="todo in todos.filter(t => t.status === 0 && !completedButUndoableTodos.has(t.id))" 
              :key="todo.id"
              class="px-2 py-1 rounded hover:bg-[var(--hover-bg-color)] transition-all duration-300"
              :class="{
                'opacity-60 scale-95': completingTodos.has(todo.id)
              }"
            >
              <BaseCheckbox
                :id="`todo-${todo.id}`"
                :model-value="checkedTodos.has(todo.id)"
                @change="handleCompleteTodo(todo.id)"
                class="w-full"
              >
                <span 
                  class="text-[var(--text-color-next-500)] text-sm whitespace-pre-wrap transition-all duration-300"
                  :class="{
                    'line-through opacity-50': completingTodos.has(todo.id)
                  }"
                >
                  {{ todo.content }}
                </span>
              </BaseCheckbox>
            </div>
          </transition-group>
        </div>
        
        <!-- 无待办状态 -->
        <div v-else class="text-[var(--text-color-next-500)] text-sm">
          今日无事🎉
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Done from '../icons/done.vue'
import Busy from '../icons/busy.vue'
import Delete from '../icons/delete.vue'
import BaseButton from '../common/BaseButton.vue'
import BaseCheckbox from '../common/BaseCheckbox.vue'
import { fetchUpdateTodo, fetchDeleteTodo } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useTodoStore } from '@/stores/todo'
import { storeToRefs } from 'pinia'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { ref, onUnmounted } from 'vue'

const { openConfirm } = useBaseDialog()

const props = defineProps<{
  todo: App.Api.Todo.Todo | undefined
  index: number
  operative: boolean
}>()

const emit = defineEmits(['refresh'])

const todoStore = useTodoStore()
const { loading, todos } = storeToRefs(todoStore)

// 跟踪正在完成的代办（用于显示删除线和消失动画）
const completingTodos = ref<Set<number>>(new Set())

// 跟踪待撤销的操作
const pendingUndoTodos = ref<Map<number, { timeoutId: ReturnType<typeof setTimeout>; todoContent: string }>>(new Map())

// 跟踪复选框的选中状态（受控模式）
const checkedTodos = ref<Set<number>>(new Set())

// 跟踪正在执行API的代办（防止重复操作）
const executingTodos = ref<Set<number>>(new Set())

// 跟踪已完成但可撤销的代办（这些代办会从列表中隐藏）
const completedButUndoableTodos = ref<Set<number>>(new Set())

// 跟踪消失定时器
const disappearTimeouts = ref<Map<number, ReturnType<typeof setTimeout>>>(new Map())

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

const handleCompleteTodo = (todoId: number) => {
  // 找到对应的代办内容
  const todo = todos.value.find(t => t.id === todoId)
  if (!todo) return
  
  // 立即标记为选中和正在完成（显示删除线和选中状态）
  checkedTodos.value.add(todoId)
  completingTodos.value.add(todoId)
  
  // 短暂延迟后从列表中移除，让用户看到删除线效果
  const disappearTimeoutId = setTimeout(() => {
    completedButUndoableTodos.value.add(todoId)
  }, 500) // 500ms后消失
  
  disappearTimeouts.value.set(todoId, disappearTimeoutId)
  
  // 显示撤销提示
  const undoTimeoutId = setTimeout(() => {
    // 3秒后执行真正的完成操作
    executeCompleteTodo(todoId)
  }, 3000)
  
  // 记录待撤销的操作
  pendingUndoTodos.value.set(todoId, {
    timeoutId: undoTimeoutId,
    todoContent: todo.content
  })
  
  // 显示带撤销按钮的提示
  theToast.success('待办已完成！', {
    duration: 3000,
    action: {
      label: '撤销',
      onClick: () => handleUndoComplete(todoId)
    }
  })
}

const executeCompleteTodo = (todoId: number) => {
  // 检查是否还有待撤销的记录（可能已被用户撤销）
  if (!pendingUndoTodos.value.has(todoId)) {
    return // 已被撤销，不执行API调用
  }
  
  // 检查是否已在执行API（防止重复调用）
  if (executingTodos.value.has(todoId)) {
    return
  }
  
  // 标记为正在执行API
  executingTodos.value.add(todoId)
  // 立即清理待撤销记录，防止用户在API调用后还能撤销
  pendingUndoTodos.value.delete(todoId)
  
  fetchUpdateTodo(todoId).then((res) => {
    if (res.code === 1) {
      // API成功，刷新列表（代办已经从UI中移除）
      todoStore.getTodos()
      // 清理状态
      checkedTodos.value.delete(todoId)
      completingTodos.value.delete(todoId)
      completedButUndoableTodos.value.delete(todoId)
      executingTodos.value.delete(todoId)
      disappearTimeouts.value.delete(todoId)
    } else {
      // 如果失败，恢复状态（让代办重新出现在列表中）
      checkedTodos.value.delete(todoId)
      completingTodos.value.delete(todoId)
      completedButUndoableTodos.value.delete(todoId)
      executingTodos.value.delete(todoId)
      disappearTimeouts.value.delete(todoId)
      theToast.error('完成代办失败，请重试')
    }
  }).catch(() => {
    // 如果出错，恢复状态（让代办重新出现在列表中）
    checkedTodos.value.delete(todoId)
    completingTodos.value.delete(todoId)
    completedButUndoableTodos.value.delete(todoId)
    executingTodos.value.delete(todoId)
    disappearTimeouts.value.delete(todoId)
    theToast.error('完成代办失败，请重试')
  })
}

const handleUndoComplete = (todoId: number) => {
  // 如果已经在执行API，无法撤销
  if (executingTodos.value.has(todoId)) {
    theToast.warning('操作正在执行中，无法撤销')
    return
  }
  
  const pendingUndo = pendingUndoTodos.value.get(todoId)
  if (!pendingUndo) {
    // 没有待撤销的记录，说明已经执行完成或已经撤销过了
    theToast.warning('操作已完成，无法撤销')
    return
  }
  
  // 清除定时器
  clearTimeout(pendingUndo.timeoutId)
  // 移除待撤销记录
  pendingUndoTodos.value.delete(todoId)
  
  // 清除消失定时器
  const disappearTimeoutId = disappearTimeouts.value.get(todoId)
  if (disappearTimeoutId) {
    clearTimeout(disappearTimeoutId)
    disappearTimeouts.value.delete(todoId)
  }
  
  // 恢复UI状态
  checkedTodos.value.delete(todoId)  // 重置复选框状态
  completingTodos.value.delete(todoId)
  completedButUndoableTodos.value.delete(todoId)  // 重新显示在列表中
  
  // 显示撤销成功提示
  theToast.info('已撤销完成操作')
}

// 组件卸载时清理所有定时器
onUnmounted(() => {
  pendingUndoTodos.value.forEach((pendingUndo) => {
    clearTimeout(pendingUndo.timeoutId)
  })
  pendingUndoTodos.value.clear()
  
  disappearTimeouts.value.forEach((timeoutId) => {
    clearTimeout(timeoutId)
  })
  disappearTimeouts.value.clear()
})
</script>

<style scoped>
/* 代办项目的进入/离开动画 */
.todo-item-enter-active,
.todo-item-leave-active {
  transition: all 0.5s ease;
}

.todo-item-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.todo-item-leave-to {
  opacity: 0;
  transform: translateX(20px) scale(0.95);
}

.todo-item-move {
  transition: transform 0.5s ease;
}
</style>
