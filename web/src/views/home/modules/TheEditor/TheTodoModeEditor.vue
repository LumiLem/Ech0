<template>
  <div>
    <div>
      <h2 class="text-[var(--text-color-500)] font-bold mb-1">Todo 模式</h2>
    </div>
    <div class="text-[var(--text-color-400)] my-2 font-serif">
      <p>
        当前待办事项数量: <span class="font-bold">{{ todos.length }}</span>
      </p>
    </div>
    <!-- todoMode -->
    <BaseTextArea
      v-model="content"
      class="rounded-lg h-auto sm:min-h-[6rem] md:min-h-[9rem]"
      placeholder="请输入待办事项..."
      :rows="3"
    />
  </div>
</template>

<script setup lang="ts">
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import { useTodoStore } from '@/stores'
import { storeToRefs } from 'pinia'

const todoStore = useTodoStore()
const editorStore = useEditorStore()
const content = computed({
  get: () => editorStore.todoToAdd.content,
  set: (val) => {
    editorStore.todoToAdd.content = val
  },
})

const { todos } = storeToRefs(todoStore)
</script>

<style scoped></style>
