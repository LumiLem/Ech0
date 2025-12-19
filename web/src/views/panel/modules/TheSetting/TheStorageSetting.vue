<template>
  <PanelCard>
    <!-- 存储设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">存储设置</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="storageEditMode" @click="handleUpdateS3Setting" title="编辑">
            <Saveupdate class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
          <button @click="storageEditMode = !storageEditMode" title="编辑">
            <Edit
              v-if="!storageEditMode"
              class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 开启S3 -->
      <div class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] h-10">
        <h2 class="font-semibold w-30 shrink-0">启用S3存储:</h2>
        <BaseSwitch v-model="S3Setting.enable" :disabled="!storageEditMode" />
      </div>

      <!-- 使用 SSL -->
      <div class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] h-10">
        <h2 class="font-semibold w-30 shrink-0">启用SSL:</h2>
        <BaseSwitch v-model="S3Setting.use_ssl" :disabled="!storageEditMode" />
      </div>

      <!-- S3 Service Provider -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">S3 服务:</h2>
        <BaseSelect
          v-model="S3Setting.provider"
          :options="S3ServiceOptions"
          :disabled="!storageEditMode"
          class="w-fit h-8"
        />
      </div>

      <!-- S3 Endpoint -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">S3 Endpoint:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.endpoint"
          style="vertical-align: middle"
        >
          {{ S3Setting.endpoint.length === 0 ? '暂无' : S3Setting.endpoint }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.endpoint"
          type="text"
          placeholder="S3 Endpoint地址"
          class="w-full py-1!"
        />
      </div>

      <!-- S3 Access Key -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">Access Key:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.access_key"
          style="vertical-align: middle"
        >
          {{ S3Setting.access_key.length === 0 ? '暂无' : S3Setting.access_key }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.access_key"
          type="text"
          placeholder="S3 Access Key"
          class="w-full py-1!"
        />
      </div>

      <!-- S3 Secret Key -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">Secret Key:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.secret_key"
          style="vertical-align: middle"
        >
          {{ S3Setting.secret_key.length === 0 ? '暂无' : S3Setting.secret_key }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.secret_key"
          type="text"
          placeholder="S3 Secret Key"
          class="w-full py-1!"
        />
      </div>

      <!-- S3 Bucket -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">S3 Bucket:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.bucket_name"
          style="vertical-align: middle"
        >
          {{ S3Setting.bucket_name.length === 0 ? '暂无' : S3Setting.bucket_name }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.bucket_name"
          type="text"
          placeholder="S3 Bucket Name"
          class="w-full py-1!"
        />
      </div>

      <!-- Path Prefix -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">Path Prefix:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.path_prefix"
          style="vertical-align: middle"
        >
          {{ S3Setting.path_prefix.length === 0 ? '暂无' : S3Setting.path_prefix }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.path_prefix"
          type="text"
          placeholder="S3 Path Prefix（可选）"
          class="w-full py-1!"
        />
      </div>

      <!-- S3 Region -->
      <div
        v-if="S3Setting.provider !== S3Provider.MINIO"
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">S3 Region:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.region"
          style="vertical-align: middle"
        >
          {{ S3Setting.region.length === 0 ? '暂无' : S3Setting.region }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.region"
          type="text"
          placeholder="S3 Region"
          class="w-full py-1!"
        />
      </div>

      <!-- CDN 加速域名（可选） -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-30 shrink-0">CDN 域名:</h2>
        <span
          v-if="!storageEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="S3Setting.cdn_url"
          style="vertical-align: middle"
        >
          {{ S3Setting.cdn_url.length === 0 ? '暂无' : S3Setting.cdn_url }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.cdn_url"
          type="text"
          placeholder="S3 CDN 域名（可选）"
          class="w-full py-1!"
        />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { S3Provider } from '@/enums/enums'
import { fetchUpdateS3Settings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getS3Setting } = settingStore
const { S3Setting } = storeToRefs(settingStore)

const storageEditMode = ref<boolean>(false)

const S3ServiceOptions = ref<{ label: string; value: S3Provider }[]>([
  { label: 'AWS', value: S3Provider.AWS },
  { label: 'MinIO', value: S3Provider.MINIO },
  { label: 'Cloudflare R2', value: S3Provider.R2 },
  // { label: '阿里OSS', value: S3Provider.ALIYUN },
  // { label: '腾讯COS', value: S3Provider.TENCENT },
  { label: 'Other', value: S3Provider.OTHER },
])

const handleUpdateS3Setting = async () => {
  await fetchUpdateS3Settings(settingStore.S3Setting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      storageEditMode.value = false
      // 重新获取S3设置
      getS3Setting()
    })
}

onMounted(() => {
  getS3Setting()
})
</script>
