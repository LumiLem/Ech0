<template>
  <div class="w-full px-2">
    <!-- 一个月内的热力图 -->
    <div class="flex justify-center items-center">
      <TheHeatMap />
    </div>

    <!-- 系统状态 -->
    <div class="justify-center my-1">
      <TheStatusCard />
    </div>

    <!-- 当前在忙 -->
    <div v-if="isLogin" class="justify-center my-2 px-9 md:px-11">
      <TheTodoCard :todo="todos[0]" :index="0" :operative="false" @refresh="getTodos" />
    </div>

    <!-- Ech0 Connect -->
    <div class="justify-center my-2">
      <TheConnects />
    </div>

    <!-- 音乐播放器 -->
    <div class="justify-center my-1">
      <TheAudioCard />
    </div>

    <!-- 近况总结 -->
    <div class="justify-center my-2">
      <TheRecentCard />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheHeatMap from '@/components/advanced/TheHeatMap.vue'
import TheConnects from '@/views/connect/modules/TheConnects.vue'
import TheStatusCard from '@/components/advanced/TheStatusCard.vue'
import TheTodoCard from '@/components/advanced/TheTodoCard.vue'
import TheAudioCard from '@/components/advanced/TheAudioCard.vue'
import TheRecentCard from '@/components/advanced/TheRecentCard.vue'

import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { useTodoStore } from '@/stores/todo'
import { onMounted } from 'vue'
const todoStore = useTodoStore()
const userStore = useUserStore()
const { isLogin } = storeToRefs(userStore)
const { getTodos } = useTodoStore()
const { todos } = storeToRefs(todoStore)

onMounted(() => {
  if (isLogin.value) {
    getTodos()
  }
})
</script>
