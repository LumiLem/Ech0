<template>
  <PanelCard>
    <div class="flex flex-row items-center justify-between mb-3">
      <h1 class="text-[var(--text-color-600)] font-bold text-lg">用户管理</h1>
      <div class="flex flex-row items-center justify-end gap-2 w-14">
        <!-- <button @click="userEditMode = !userEditMode" title="编辑">
          <Edit v-if="!userEditMode" class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
        </button> -->
      </div>
    </div>

    <!-- 用户列表 -->
    <div v-if="loading" class="flex justify-center py-4 text-[var(--text-color-400)]">
      加载中...
    </div>

    <div v-else>
      <div v-if="allusers.length === 0" class="flex flex-col items-center justify-center mt-2">
        <span class="text-[var(--text-color-next-400)]">暂无其它用户...</span>
      </div>

      <div v-else class="mt-2 overflow-x-auto border border-[var(--border-color-300)] rounded-lg">
        <table class="min-w-full divide-y divide-[var(--divide-color-200)]">
          <thead>
            <tr class="bg-[var(--bg-color-50)] opacity-70">
              <th
                class="px-3 py-2 text-left text-sm font-semibold text-[var(--text-color-next-600)]"
              >
                #
              </th>
              <th
                class="px-3 min-w-18 py-2 text-left text-sm font-semibold text-[var(--text-color-next-600)]"
              >
                用户名
              </th>
              <th
                class="px-3 py-2 text-center text-sm font-semibold text-[var(--text-color-next-600)]"
              >
                权限更改
              </th>
              <th
                class="px-3 min-w-18 py-2 text-right text-sm font-semibold text-[var(--text-color-next-600)]"
              >
                操作
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--divide-color-100)] text-nowrap">
            <tr v-for="(user, index) in allusers" :key="user.id" class="">
              <td class="px-3 py-2 text-sm text-[var(--text-color-next-700)]">{{ index + 1 }}</td>
              <td class="px-3 py-2 text-sm text-[var(--text-color-next-700)] font-semibold">
                {{ user.username }}
              </td>
              <td class="px-3 py-2 text-center">
                <BaseSwitch v-model="user.is_admin" @click="handleUpdateUserPermission(user.id)" />
              </td>
              <td class="px-3 py-2 text-right">
                <button
                  class="p-1 hover:bg-[#fff6eb] rounded"
                  @click="handleDeleteUser(user.id)"
                  title="删除用户"
                >
                  <Deluser class="w-5 h-5 text-red-500" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
// import Edit from '@/components/icons/edit.vue'
// import Close from '@/components/icons/close.vue'
import { ref, onMounted } from 'vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import Deluser from '@/components/icons/deluser.vue'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'
const { openConfirm } = useBaseDialog()

const loading = ref<boolean>(true)

import { fetchGetAllUsers, fetchUpdateUserPermission, fetchDeleteUser } from '@/service/api'

const allusers = ref<App.Api.User.User[]>([])
// const userEditMode = ref<boolean>(false)

const handleDeleteUser = async (userId: number) => {
  openConfirm({
    title: '确定要删除该用户吗？',
    description: '删除后将无法恢复，请谨慎操作',
    onConfirm: () => {
      fetchDeleteUser(userId).then((res) => {
        if (res.code === 1) {
          getAllUsers()
        }
      })
    },
  })
}

const handleUpdateUserPermission = async (userId: number) => {
  fetchUpdateUserPermission(userId)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      // 重新获取设置
      getAllUsers()
    })
}

const getAllUsers = async () => {
  loading.value = true
  try {
    const res = await fetchGetAllUsers()
    if (res.code === 1) {
      allusers.value = res.data
    }
    loading.value = false
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  getAllUsers()
})
</script>

<style scoped></style>
