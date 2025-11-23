<template>
  <div class="px-9 md:px-11">
    <div
      class="widget rounded-md shadow-sm hover:shadow-md ring-1 ring-[var(--ring-color)] ring-inset p-4"
    >
      <h2 class="text-[var(--widget-title-color)] font-bold text-lg mb-1 flex items-center">
        <StatusInfo class="mr-2" />系统状态:
      </h2>

      <div v-if="!isLoading" class="text-[var(--text-color-next-500)] text-sm">
        <!-- 系统管理员 -->
        <!-- <div>
          <h1>
            当前系统管理员：
            <span class="ml-2">{{ status?.username }}</span>
          </h1>
        </div> -->
        <!-- 当前登录用户 -->
        <div>
          <p>
            <span class="">当前登录用户:</span>
            <span class="ml-2">
              {{ userStore?.user?.username ? userStore.user.username : '未登录' }}
            </span>
          </p>
        </div>
        <!-- 当前共有Ech0 -->
        <div>
          <p>
            <span class="">已发布Echos:</span>
            <span class="ml-2">{{ status?.total_echos }}</span>
            <span class="ml-2">条</span>
          </p>
        </div>
      </div>
      <div v-else>
        <div class="text-[var(--text-color-next-500)] text-sm">加载中...</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { fetchGetStatus } from '@/service/api'
import { onMounted, ref } from 'vue'
import { useUserStore } from '@/stores/user'
import StatusInfo from '../icons/statusinfo.vue'

const status = ref<App.Api.Ech0.Status>()
const userStore = useUserStore()
const isLoading = ref<boolean>(true)

onMounted(() => {
  fetchGetStatus().then((res) => {
    status.value = res.data
    isLoading.value = false
  })
})
</script>

<style scoped></style>
