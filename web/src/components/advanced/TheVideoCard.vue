<template>
  <div
    class="max-w-sm flex justify-center items-center bg-[var(--card-color)] rounded-lg shadow-sm ring-1 ring-inset ring-[var(--ring-color-100)] p-2 gap-2"
  >
    <!-- 使用移动端外链播放器（解决跳转） -->
    <iframe
      v-if="props.videoId.startsWith('BV')"
      :src="`https://www.bilibili.com/blackboard/html5mobileplayer.html?bvid=${props.videoId}&as_wide=1&high_quality=1&danmaku=0`"
      scrolling="no"
      border="0"
      frameborder="no"
      framespacing="0"
      allowfullscreen="true"
      loading="lazy"
      class="w-full aspect-video rounded"
    ></iframe>
    <iframe
      v-else
      :src="`https://www.youtube.com/embed/${props.videoId}`"
      frameborder="0"
      allow="accelerometer; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
      allowfullscreen
      loading="lazy"
      class="w-full aspect-video rounded"
    ></iframe>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  videoId: string
}>()
</script>

<style scoped></style>
