<template>
  <PanelCard>
    <!-- 数据管理 -->
    <div>
      <div class="flex items-center justify-start mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">数据管理</h1>
      </div>

      <div class="flex flex-col gap-4 font-semibold">
        <!-- 备份数据 -->
        <div class="flex flex-start items-center gap-2">
          <p class="text-[var(--text-color-next-500)]">创建快照:</p>
          <BaseButton
            :icon="CreateBackup"
            @click="handleBackup"
            class="rounded-lg text-[var(--text-color-600)]!"
            title="创建快照"
          />
        </div>
        <!-- 导出快照 -->
        <div class="flex flex-start items-center gap-2">
          <p class="text-[var(--text-color-next-500)]">导出快照:</p>
          <BaseButton
            :icon="ExportBackup"
            @click="handleBackupExport"
            class="rounded-lg text-[var(--text-color-600)]!"
            title="导出快照"
          />
        </div>
        <!-- 恢复数据 -->
        <div class="flex flex-start items-center gap-2">
          <p class="text-[var(--text-color-next-500)]">恢复快照:</p>
          <BaseButton
            :icon="RestoreBackup"
            @click="handleBackupRestore"
            class="rounded-lg text-[var(--text-color-600)]!"
            title="恢复快照"
          />
        </div>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import CreateBackup from '@/components/icons/createbackup.vue'
import ExportBackup from '@/components/icons/exportbackup.vue'
import RestoreBackup from '@/components/icons/restorebackup.vue'
import { fetchBackup, fetchImportBackup } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'

const userStore = useUserStore()
const { isLogin } = storeToRefs(userStore)

const handleBackup = async () => {
  await theToast.promise(fetchBackup(), {
    loading: '备份中...',
    success: (res) => (res.code === 1 ? '备份成功' : '备份失败'),
    error: '备份失败',
  })
}

const handleBackupExport = async () => {
  if (!isLogin.value) {
    theToast.info('请登录后使用', { duration: 3000 })
    return
  }

  try {
    theToast.info('导出中...请稍等', {
      duration: 4000,
    })

    // 1. 获取 token
    const token = localStorage.getItem('token')
    const baseURL =
      import.meta.env.VITE_SERVICE_BASE_URL === '/'
        ? window.location.origin
        : import.meta.env.VITE_SERVICE_BASE_URL
    const downloadUrl = `${baseURL}/api/backup/export?token=${token}`

    // 创建隐藏的 a 标签触发下载
    const link = document.createElement('a')
    link.href = downloadUrl
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    theToast.success('备份导出开始！')
  } catch (error) {
    theToast.error('导出失败')
    console.error('导出备份失败:', error)
  }
}

const handleBackupRestore = async () => {
  if (!isLogin.value) {
    theToast.info('请登录后使用', { duration: 3000 })
    return
  }

  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.zip'
  input.onchange = async (event: Event) => {
    const target = event.target as HTMLInputElement
    if (target.files && target.files.length > 0) {
      const file = target.files[0]

      if (file) {
        await theToast.promise(
          fetchImportBackup(file),
          {
            loading: '导入中,请不要关闭页面...',
            success: (res) => (res.code === 1 ? '快照恢复成功🎉' : `导入失败: ${res.msg}`),
            error: '导入失败,请尝试重新导入或使用TUI模式进行恢复',
          },
          {
            duration: 5000,
          },
        )
      }
    }
  }
  input.click()
}
</script>
