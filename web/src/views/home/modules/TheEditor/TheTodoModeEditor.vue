<template>
  <!-- todoMode -->
  <BaseTextArea
    v-model="content"
    class="rounded-lg h-auto sm:min-h-[6rem] md:min-h-[9rem]"
    placeholder="请输入待办事项..."
    :rows="3"
  />
</template>

<script setup lang="ts">
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import { computed } from 'vue'
import { useEditorStore } from '@/stores'

const editorStore = useEditorStore()
const content = computed({
  get: () => editorStore.todoToAdd.content,
  set: (val) => {
    editorStore.todoToAdd.content = val
  },
})
</script>

<style scoped></style>
