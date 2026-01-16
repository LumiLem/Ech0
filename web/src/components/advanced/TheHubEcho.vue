<template>
  <div class="w-full max-w-sm bg-[var(--card-color)] h-auto p-5 shadow rounded-lg mx-auto">
    <!-- 顶部用户头像和信息 -->
    <div class="flex flex-row items-center gap-2 mt-2 mb-4">
      <div>
        <img
          :src="userAvatar"
          alt="用户头像"
          class="w-10 h-10 sm:w-12 sm:h-12 rounded-full ring-1 ring-gray-200 shadow-sm object-cover"
          @error="handleImageError"
        />
      </div>
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--text-color-700)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            {{ displayUsername }}
          </h2>

          <div>
            <Verified class="text-sky-500 w-5 h-5" />
          </div>
        </div>
        <span class="text-[#5b7083] font-serif flex items-center gap-1">
          <span>🌐</span>
          <a :href="echo.server_url" target="_blank">{{ echo.server_name }}</a>
        </span>
      </div>
    </div>

    <!-- 图片 && 内容 -->
    <div>
      <div class="py-4">
        <!-- grid 和 horizontal 时，文字在图片上；其他布局（waterfall/carousel/null/undefined）文字在图片下 -->
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID || props.echo.layout === ImageLayout.HORIZONTAL
          "
        >
          <!-- 文字在上 -->
          <div class="mx-auto w-11/12 pl-1 mb-3">
            <MdPreview
              :id="previewOptions.proviewId"
              :modelValue="props.echo.content"
              :theme="theme"
              :show-code-row-number="previewOptions.showCodeRowNumber"
              :preview-theme="previewOptions.previewTheme"
              :code-theme="previewOptions.codeTheme"
              :code-style-reverse="previewOptions.codeStyleReverse"
              :no-img-zoom-in="previewOptions.noImgZoomIn"
              :code-foldable="previewOptions.codeFoldable"
              :auto-fold-threshold="previewOptions.autoFoldThreshold"
            />
          </div>

          <TheImageGallery
            :media="props.echo.media"
            :baseUrl="echo.server_url"
            :layout="props.echo.layout"
          />
        </template>

        <template v-else>
          <!-- 图片在上，文字在下（瀑布流 / 单图轮播 等） -->
          <TheImageGallery
            :media="props.echo.media"
            :baseUrl="echo.server_url"
            :layout="props.echo.layout"
          />

          <div class="mx-auto w-11/12 pl-1 mt-3">
            <MdPreview
              :id="previewOptions.proviewId"
              :modelValue="props.echo.content"
              :theme="theme"
              :show-code-row-number="previewOptions.showCodeRowNumber"
              :preview-theme="previewOptions.previewTheme"
              :code-theme="previewOptions.codeTheme"
              :code-style-reverse="previewOptions.codeStyleReverse"
              :no-img-zoom-in="previewOptions.noImgZoomIn"
              :code-foldable="previewOptions.codeFoldable"
              :auto-fold-threshold="previewOptions.autoFoldThreshold"
            />
          </div>
        </template>

        <!-- 扩展内容 -->
        <div v-if="props.echo.extension" class="my-4">
          <div v-if="props.echo.extension_type === ExtensionType.MUSIC">
            <TheAPlayerCard :echo="props.echo" />
          </div>
          <div v-if="props.echo.extension_type === ExtensionType.VIDEO">
            <TheVideoCard :videoId="props.echo.extension" class="px-2 mx-auto hover:shadow-md" />
          </div>
          <TheGithubCard
            v-if="props.echo.extension_type === ExtensionType.GITHUBPROJ"
            :GithubURL="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
          <TheWebsiteCard
            v-if="props.echo.extension_type === ExtensionType.WEBSITE"
            :website="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
        </div>
      </div>
    </div>

    <!-- 日期时间 && 操作按钮 -->
    <div class="flex justify-between items-center">
      <!-- 日期时间 -->
      <div class="flex justify-start items-center h-auto">
        <div class="flex justify-start text-sm text-slate-500 mr-1">
          {{ formatDate(props.echo.created_at) }}
        </div>
        <!-- 标签 -->
        <div class="text-sm text-[var(--text-color-300)] w-18 truncate text-nowrap">
          <span>{{ props.echo.tags && props.echo.tags[0]?.name ? `#${props.echo.tags[0].name}` : '' }}</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div ref="menuRef" class="relative flex items-center justify-center gap-2 h-auto">
        <!-- 跳转 -->
        <a :href="`${server_url}/echo/${echo_id}`" target="_blank" title="跳转至该 Echo">
          <LinkTo class="w-4 h-4" />
        </a>

        <!-- 点赞 -->
        <div class="flex items-center justify-end" title="点赞">
          <div class="flex items-center gap-1">
            <!-- 点赞按钮   -->
            <button
              @click="handleLikeEcho()"
              title="点赞"
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </button>

            <!-- 点赞数量   -->
            <span class="text-sm text-[var(--text-color-400)]">
              <!-- 如果点赞数不超过99，则显示数字，否则显示99+ -->
              {{ fav_count > 99 ? '99+' : fav_count }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheGithubCard from './TheGithubCard.vue'
import TheVideoCard from './TheVideoCard.vue'
import Verified from '../icons/verified.vue'
import GrayLike from '../icons/graylike.vue'
import LinkTo from '../icons/linkto.vue'
import TheAPlayerCard from './TheAPlayerCard.vue'
import TheWebsiteCard from './TheWebsiteCard.vue'
import TheImageGallery from './TheImageGallery.vue'
import 'md-editor-v3/lib/preview.css'
import { MdPreview } from 'md-editor-v3'
import { onMounted, computed, ref } from 'vue'
import { ExtensionType, ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { useThemeStore } from '@/stores'
import { useFetch } from '@vueuse/core'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'

type Echo = App.Api.Hub.Echo

const props = defineProps<{
  echo: Echo
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

const fav_count = ref<number>(props.echo.fav_count)
const server_url = props.echo.server_url
const echo_id = props.echo.id
const isLikeAnimating = ref(false)
const LIKE_LIST_KEY = server_url + '_liked_echo_ids'

const handleLikeEcho = async () => {
  isLikeAnimating.value = true
  setTimeout(() => {
    isLikeAnimating.value = false
  }, 250)

  // 如果已经点赞过，不再重复点赞
  const likedEchoIds: number[] = localStg.getItem(LIKE_LIST_KEY) || []
  if (likedEchoIds.includes(echo_id)) {
    theToast.info('你已经点赞过')
    return
  }

  // 调用后端接口，点赞
  const { error, data } = await useFetch<App.Api.Response<null>>(
    `${server_url}/api/echo/like/${echo_id}`,
  )
    .put()
    .json()

  if (error.value || data.value?.code !== 1) {
    theToast.error('点赞失败')
  } else {
    fav_count.value += 1
    likedEchoIds.push(echo_id)
    localStg.setItem(LIKE_LIST_KEY, likedEchoIds)
    theToast.success('点赞成功')
  }
}

// 用户头像（优先使用user.avatar，如果不存在则使用logo作为fallback）
const userAvatar = computed(() => {
  // 优先使用关联查询的发布者头像（新版本服务器）
  if (props.echo.user?.avatar) {
    const avatar = props.echo.user.avatar
    // 如果是完整URL则直接使用，否则需要拼接服务器地址
    if (avatar.startsWith('http')) {
      return avatar
    }
    // 拼接远程服务器地址
    return `${props.echo.server_url}/api${avatar}`
  }
  
  // Fallback: 使用站点Logo（旧版本服务器或新版本服务器没有user字段时）
  const logo = props.echo.logo
  if (logo && logo.length > 0) {
    return logo
  }
  
  return '/favicon.svg'
})

// 显示用户名（优先使用user.username，如果不存在则使用echo.username）
const displayUsername = computed(() => {
  // 优先使用实时查询的发布者用户名（新版本服务器）
  // Fallback到echos表中保存的用户名快照（旧版本服务器或新版本服务器没有user字段时）
  return props.echo.user?.username || props.echo.username || '未知用户'
})

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src = '/favicon.svg'
}

onMounted(() => {})
</script>

<style scoped lang="css">
#preview-only {
  background-color: inherit;
}

.md-editor {
  font-family: var(--font-sans);
  /* font-family: 'LXGW WenKai Screen'; */
}

:deep(ul li) {
  list-style-type: disc;
}

:deep(ul li li) {
  list-style-type: circle;
}

:deep(ul li li li) {
  list-style-type: square;
}

:deep(ol li) {
  list-style-type: decimal;
}

:deep(p) {
  white-space: normal;
  /* 允许正常换行 */
  overflow-wrap: break-word;
  /* 单词太长时自动换行 */
  word-break: normal;
  /* 保持单词整体性，不随便拆开 */
}
</style>
