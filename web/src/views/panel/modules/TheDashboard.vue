<template>
  <div class="w-full px-2">
    <!-- 加载状态 -->
    <div v-if="!connected" class="flex items-center justify-center h-96">
      <div class="flex flex-col items-center gap-4">
        <div class="loader"></div>
        <p class="text-[var(--text-color-next-500)]">正在建立连接...</p>
      </div>
    </div>

    <!-- 主内容 -->
    <div v-else>
      <div class="md:flex items-center gap-2">
        <!-- CPU -->
        <MetricCard title="CPU 使用率" class="md:w-1/2 mb-2 md:mb-0">
          <VChart class="h-60" :option="cpuOption" autoresize />
        </MetricCard>

        <!-- Disk -->
        <MetricCard title="磁盘使用情况" class="md:w-1/2 mb-2 md:mb-0">
          <VChart class="h-60" :option="diskOption" autoresize />
        </MetricCard>
      </div>

      <div class="md:flex items-center mt-4 gap-2">
        <!-- Memory -->
        <MetricCard title="内存使用率" class="md:w-1/2 mb-2 md:mb-0">
          <VChart class="h-60" :option="memoryOption" autoresize />
        </MetricCard>

        <!-- System -->
        <MetricCard title="状态信息" class="md:w-1/2 mb-2 md:mb-0">
          <div class="text-md text-[var(--text-color-next-500)] font-bold h-60 p-2">
            <p>
              登录用户：<span class="text-sm font-normal">{{ userStore.user?.username }}</span>
            </p>
            <p>
              主机名：<span class="text-sm font-normal">{{ metrics.System?.Hostname }}</span>
            </p>
            <p>
              操作系统：<span class="text-sm font-normal">{{ metrics.System?.OsName }}</span>
            </p>
            <p>
              内核版本：<span class="text-sm font-normal">{{ metrics.System?.KernelVersion }}</span>
            </p>
            <p>
              运行时长：<span class="text-sm font-normal">{{ metrics.System?.Uptime }} Hours</span>
            </p>
            <p>
              当前时间：<span class="text-sm font-normal">{{ metrics.System?.Time }}</span>
            </p>
            <p>
              进程数：<span class="text-sm font-normal">{{ metrics.System?.ProcessCount }}</span>
            </p>
            <p>
              Go版本：<span class="text-sm font-normal">{{ metrics.System?.GolangVersion }}</span>
            </p>
            <p>
              协程数：<span class="text-sm font-normal">{{ metrics.System?.GoRoutineCount }}</span>
            </p>
          </div>
        </MetricCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useOWebSocket } from '@/service/request/websocket'
import MetricCard from '@/layout/MetricCard.vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
} from 'echarts/components'
import { GaugeChart, LineChart, BarChart, PieChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import { useUserStore } from '@/stores/user'
import { getWsUrl } from '@/service/request/shared'

const connected = ref<boolean>(false)

const userStore = useUserStore()

// 注册按需组件
use([
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  GaugeChart,
  LineChart,
  BarChart,
  PieChart,
  CanvasRenderer,
])

const metrics = ref<App.Api.Dashboard.Metrics>({
  CPU: {
    UsagePercent: 0,
    Cores: 0,
    FrequencyMHz: 0,
  },
  Memory: {
    Total: 0,
    Used: 0,
    Available: 0,
    Percentage: 0,
  },
  Disk: {
    Total: 0,
    Used: 0,
    Available: 0,
    Percentage: 0,
  },
  Network: {
    BytesSentPerSecond: 0,
    BytesReceivedPerSecond: 0,
    TotalBytesReceived: 0,
    TotalBytesSent: 0,
  },
  System: {
    Uptime: 0,
    Hostname: '',
    OsName: '',
    KernelVersion: '',
    Time: '',
    KernelArch: '',
    TimeZone: '',
    ProcessCount: 0,
    ThreadCount: 0,
    GolangVersion: '',
    GoRoutineCount: 0,
  },
})

const { onMessage, open } = useOWebSocket<App.Api.Response<App.Api.Dashboard.Metrics>>({
  url: getWsUrl('/ws/dashboard/metrics'),
  autoReconnect: true,
  heartbeat: true,
})

onMounted(() => {
  open() // 等组件挂载后再连接

  // 每次收到服务器推送的 metrics 更新 metrics
  onMessage((payload) => {
    if (payload.code === 1 && payload.data) {
      metrics.value = payload.data
      updateCharts() // 更新图表
      // 首次收到数据后显示内容
      if (!connected.value) {
        connected.value = true
      }
    }
  })
})

// ---- 图表配置 ----
const cpuOption = ref({})
const memoryOption = ref({})
const diskOption = ref({})
const networkOption = ref({})

// 实时更新函数
function updateCharts() {
  const mainColor =
    getComputedStyle(document.documentElement).getPropertyValue('--dashboard-main-color').trim() ||
    '#f5d2ae'
  const nextColor =
    getComputedStyle(document.documentElement).getPropertyValue('--dashboard-next-color').trim() ||
    '#fae2bf'

  cpuOption.value = {
    color: [mainColor],
    tooltip: {
      trigger: 'axis',
      formatter: '{a}: {c} MHz',
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      // 使用两个空点让同一个 y 值横跨整个 x 轴
      data: ['', ''],
      axisLabel: { show: false },
      axisTick: { show: false },
      axisLine: { show: false },
    },
    yAxis: {
      type: 'value',
      name: 'CPU(MHz)',
    },
    series: [
      {
        name: 'CPU 频率',
        type: 'line',
        smooth: true,
        // 不显示点
        symbol: 'none',
        // 两个相同的值使线贯穿整个 x 轴
        data: [metrics.value.CPU.FrequencyMHz, metrics.value.CPU.FrequencyMHz],
        areaStyle: {},
        lineStyle: {
          width: 2,
          color: mainColor,
        },
        itemStyle: {
          color: mainColor,
        },
      },
    ],
  }

  memoryOption.value = {
    color: [mainColor, nextColor],
    tooltip: { trigger: 'item' },
    legend: { bottom: '0%' },
    series: [
      {
        name: '内存',
        type: 'pie',
        radius: ['40%', '70%'],
        label: {
          show: false,
          position: 'center',
          formatter: () => {
            // 显示百分比，保留1位小数
            return `${metrics.value.Memory.Percentage.toFixed(1)}%`
          },
          fontSize: 12,
          fontWeight: 'bold',
          color: '#a67c52',
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 28,
            fontWeight: 'bold',
          },
        },
        data: [
          { value: metrics.value.Memory.Used, name: '已用内存(GB)' },
          { value: metrics.value.Memory.Available, name: '可用内存(GB)' },
        ],
      },
    ],
  }

  diskOption.value = {
    tooltip: { trigger: 'item' },
    color: [mainColor],
    grid: { left: '10%', right: '10%', bottom: '10%', top: '15%' },
    xAxis: { type: 'category', data: ['磁盘使用率'] },
    yAxis: { type: 'value', max: 100 },
    series: [
      {
        type: 'bar',
        data: [metrics.value.Disk.Percentage],
        label: {
          show: true,
          position: 'top',
          formatter: (params: { value: number }) => `${Number(params.value).toFixed(2)}%`,
        },
      },
    ],
  }

  networkOption.value = {
    color: [mainColor],
    tooltip: { trigger: 'axis' },
    legend: { data: ['上传(B/s)', '下载(B/s)'] },
    xAxis: { type: 'category', data: [metrics.value.System.Time] },
    yAxis: { type: 'value' },
    series: [
      {
        name: '上传(B/s)',
        type: 'line',
        data: [metrics.value.Network.BytesSentPerSecond],
      },
      {
        name: '下载(B/s)',
        type: 'line',
        data: [metrics.value.Network.BytesReceivedPerSecond],
      },
    ],
  }
}
</script>

<style scoped>
/* 加载圆圈动画 */
.loader {
  width: 48px;
  height: 48px;
  border: 4px solid var(--dashboard-main-color);
  border-bottom-color: transparent;
  border-radius: 50%;
  display: inline-block;
  box-sizing: border-box;
  animation: rotation 1s linear infinite;
}

@keyframes rotation {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
