<template>
  <!-- 网易云 / QQ 音乐使用 Meting JS来展示 -->
  <div
    v-if="musicInfo && musicInfo.server !== MusicProvider.APPLE && metingAPI.length > 0 && !loading"
  >
    <meting-js
      :api="metingAPI"
      :server="musicInfo.server"
      :type="musicInfo.type"
      :id="musicInfo.id"
      :auto="props.echo.extension"
    >
    </meting-js>
  </div>
  <!-- Apple Music 使用官方IFrame -->
  <div
    v-else-if="musicInfo && musicInfo.server === MusicProvider.APPLE && musicInfo.id"
    class="shadow-sm rounded-xl overflow-hidden"
  >
    <iframe
      allow="autoplay *; encrypted-media *; fullscreen *; clipboard-write"
      frameborder="0"
      height="175"
      style="width: 100%; max-width: 660px; overflow: hidden; border-radius: 10px"
      sandbox="allow-forms allow-popups allow-same-origin allow-scripts allow-storage-access-by-user-activation allow-top-navigation-by-user-activation"
      :src="`https://embed.music.apple.com/cn/${musicInfo.type}/${musicInfo.id}`"
    >
    </iframe>
  </div>
  <div
    v-else
    class="max-w-sm flex justify-center items-center bg-[var(--card-color)] rounded-lg shadow-sm ring-1 ring-inset ring-[var(--ring-color-100)] p-2 gap-2 text-[var(--text-color-400)]"
  >
    <Music />非常抱歉，该音乐播放源已失效😭
  </div>
</template>

<script setup lang="ts">
import Music from '@/components/icons/music.vue'
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { parseMusicURL } from '@/utils/other'
import { useSettingStore } from '@/stores/setting'
import { ExtensionType, MusicProvider } from '@/enums/enums'

const { SystemSetting, loading } = storeToRefs(useSettingStore())
type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()

const musicInfo = computed(() => {
  if (props.echo.extension_type !== ExtensionType.MUSIC || !props.echo.extension) return null
  return parseMusicURL(props.echo.extension)
})
const metingAPI = computed(() => {
  if (!loading.value && SystemSetting.value && SystemSetting.value.meting_api.length > 0) {
    return SystemSetting.value.meting_api + '?server=:server&type=:type&id=:id&auth=:auth&r=:r'
  } else {
    return 'https://meting.soopy.cn/api?server=:server&type=:type&id=:id&auth=:auth&r=:r'
  }
})
</script>

<style scoped>
:deep(.aplayer) {
  border-radius: 5px;
  box-shadow: 0 1px 3px rgba(27, 27, 27, 0.075);
  transition: box-shadow 0.2s;
}
:deep(.aplayer):hover {
  box-shadow: 0 2px 5px rgba(19, 19, 19, 0.075);
}
</style>
