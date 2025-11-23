<template>
  <article
    class="relative flex flex-col gap-3 rounded-md bg-[var(--card-color)] px-5 py-5 shadow-sm transition duration-200 hover:-translate-y-0.5 hover:shadow-md"
  >
    <header class="flex items-center gap-3">
      <div
        class="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-full shadow-sm ring-inset ring-1 ring-gray-200 bg-amber-50"
      >
        <img
          v-if="avatarUrl"
          :src="avatarUrl"
          :alt="displayName"
          class="h-full w-full object-cover"
        />
        <span v-else class="text-lg font-semibold text-amber-600">{{ initials }}</span>
      </div>

      <div class="min-w-0 flex-1">
        <p class="truncate text-lg font-semibold text-stone-800">
          {{ displayName }}
        </p>
        <p v-if="username" class="truncate text-sm text-[var(--text-color-next-500)]">
          @{{ username }}
        </p>
      </div>

      <!-- 关注按钮 （等待接收、已关注、关注） -->
      <BaseButton
        class="rounded-md border border-dashed border-orange-400 bg-transparent! shadow-none! ring-0! text-sm text-orange-600! hover:text-orange-400!"
        @click="handleFollow"
      >
        <template v-if="followStatus === FollowStatus.NONE">关注</template>
        <template v-else-if="followStatus === FollowStatus.PENDING">等待接受</template>
        <template v-else-if="followStatus === FollowStatus.ACCEPTED">已关注</template>
        <template v-else-if="followStatus === FollowStatus.REJECTED">关注被拒</template>
      </BaseButton>
    </header>

    <section
      v-if="sanitizedSummary"
      class="prose prose-amber max-w-none text-sm text-[var(--text-color-next-600)]"
    >
      <div v-html="sanitizedSummary" />
    </section>

    <footer class="flex flex-wrap items-center gap-0.5 text-xs text-[var(--text-color-next-500)]">
      <span v-if="actor.type">类型: {{ actor.type }}</span>
      <span v-if="actor.inbox"
        >收件箱: <span class="text-amber-600">{{ actor.inbox }}</span></span
      >
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import BaseButton from '../common/BaseButton.vue'
import {
  fetchGetFollowStatus,
  fetchFollowFediverseActor,
  fetchUnfollowFediverseActor,
} from '@/service/api/fediverse'
import { FollowStatus } from '@/enums/enums'
import { theToast } from '@/utils/toast'

// ActivityPub Actor 和 Media 类型定义
type ActivityPubMedia = {
  type?: string
  mediaType?: string
  url?: string
}

// 简化的 ActivityPub Actor 类型
type ActivityPubActor = {
  id?: string
  type?: string
  name?: string
  preferredUsername?: string
  summary?: string
  icon?: ActivityPubMedia | ActivityPubMedia[]
  image?: ActivityPubMedia | ActivityPubMedia[]
  inbox?: string
}

// 组件 Props 定义
const props = defineProps<{
  actor: ActivityPubActor
  followLoading?: boolean
  followSuccess?: boolean
}>()

const actor = computed(() => props.actor ?? {})
const displayName = computed(() => actor.value.name || actor.value.preferredUsername || '未知用户')
const username = computed(() => actor.value.preferredUsername || actor.value.name || '')
const avatarUrl = computed(() => {
  const icon = actor.value.icon ?? actor.value.image
  if (!icon) return ''
  if (Array.isArray(icon)) {
    const candidate = icon.find((item) => item?.url)
    return candidate?.url ?? ''
  }
  return icon.url ?? ''
})
// 计算用户头像的首字母
const initials = computed(() => {
  const name = displayName.value.trim()
  if (!name) return 'AP'
  const words = name.split(/\s+/)
  if (words.length === 1) return words[0]?.charAt(0).toUpperCase()
  return ((words[0]?.charAt(0) ?? '') + (words[1]?.charAt(0) ?? '')).toUpperCase()
})
// 简单处理 summary，去除多余空白
const sanitizedSummary = computed(() => {
  const summary = actor.value.summary?.trim()
  if (!summary) return ''
  return summary
})

//================================================
// 关注相关
//================================================
const followStatus = ref<FollowStatus>(FollowStatus.NONE)

const handleFollow = () => {
  // 根据当前状态执行不同操作
  switch (followStatus.value) {
    case FollowStatus.NONE:
      // 调用关注接口
      if (actor.value && actor.value.id) {
        fetchFollowFediverseActor({
          targetActor: actor.value.id,
        }).then((res) => {
          if (res.code === 1) {
            followStatus.value = FollowStatus.PENDING
            theToast.info('已发送关注请求，等待对方接受')
          }
        })
      }
      break
    case FollowStatus.PENDING:
      // 取消关注请求
      if (actor.value && actor.value.id) {
        fetchUnfollowFediverseActor({
          targetActor: actor.value.id,
        }).then((res) => {
          if (res.code === 1) {
            followStatus.value = FollowStatus.NONE
            theToast.info('已取消关注请求')
          }
        })
      }
      break
    case FollowStatus.ACCEPTED:
      // 取消关注
      if (actor.value && actor.value.id) {
        fetchUnfollowFediverseActor({
          targetActor: actor.value.id,
        }).then((res) => {
          if (res.code === 1) {
            followStatus.value = FollowStatus.NONE
            theToast.info('已取消关注')
          }
        })
      }
      break
    case FollowStatus.REJECTED:
      // 重新发起关注请求
      if (actor.value && actor.value.id) {
        fetchFollowFediverseActor({
          targetActor: actor.value.id,
        }).then((res) => {
          if (res.code === 1) {
            followStatus.value = FollowStatus.PENDING
            theToast.info('已重新发送关注请求，等待对方接受')
          }
        })
      }
      break
    default:
      followStatus.value = FollowStatus.NONE
      break
  }
}

onMounted(async () => {
  // 组件挂载时，说明已经有 Actor 数据
  // 查看当前Actor的关注状态
  const currentActor = actor.value
  if (currentActor && currentActor.id) {
    const res = await fetchGetFollowStatus(currentActor.id)
    if (res.code === 1 && res.data) {
      if (res.data.length > 0) {
        switch (res.data) {
          case FollowStatus.NONE:
            followStatus.value = FollowStatus.NONE
            break
          case FollowStatus.PENDING:
            followStatus.value = FollowStatus.PENDING
            break
          case FollowStatus.ACCEPTED:
            followStatus.value = FollowStatus.ACCEPTED
            break
          case FollowStatus.REJECTED:
            followStatus.value = FollowStatus.REJECTED
            break
          default:
            followStatus.value = FollowStatus.NONE
            break
        }
      }
    }
  }
})
</script>
