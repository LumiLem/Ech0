<template>
  <MdPreview
    :id="previewOptions.proviewId"
    :modelValue="props.content"
    :theme="theme"
    :show-code-row-number="previewOptions.showCodeRowNumber"
    :preview-theme="previewOptions.previewTheme"
    :code-theme="previewOptions.codeTheme"
    :code-style-reverse="previewOptions.codeStyleReverse"
    :no-img-zoom-in="previewOptions.noImgZoomIn"
    :code-foldable="previewOptions.codeFoldable"
    :auto-fold-threshold="previewOptions.autoFoldThreshold"
  />
</template>
<script setup lang="ts">
import 'md-editor-v3/lib/preview.css'
import { MdPreview } from 'md-editor-v3'
import { computed } from 'vue'
import { useThemeStore } from '@/stores/theme'

const props = defineProps<{
  content: string
}>()
const themeStore = useThemeStore()

const theme = computed(() => (themeStore.theme === 'light' ? 'light' : 'dark'))
const previewOptions = {
  proviewId: 'preview-only',
  showCodeRowNumber: false,
  previewTheme: 'github',
  codeTheme: 'atom',
  codeStyleReverse: true,
  noImgZoomIn: false,
  codeFoldable: true,
  autoFoldThreshold: 15,
}
</script>
