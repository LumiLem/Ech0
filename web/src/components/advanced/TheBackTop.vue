<template>
  <button
    @click="scrollToTop"
    id="backToTop"
    class="cursor-pointer rounded-full shadow hover:shadow-md bg-[var(--back-to-top-bg-color)] ring-1 ring-inset ring-[var(--ring-color)] z-50"
  >
    <Arrowup class="w-full h-full" />
  </button>
</template>

<script setup lang="ts">
import Arrowup from '../icons/arrowup.vue'

const props = defineProps<{
  target?: HTMLElement | null
}>()

const scrollToTop = () => {
  // 如果传了 target 且 target 确实是滚动容器（scrollTop > 0），则滚动 target
  if (props.target && props.target.scrollTop > 0) {
    props.target.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }
  // 否则（移动端情况），滚动 window
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>
