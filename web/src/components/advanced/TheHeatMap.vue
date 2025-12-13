<template>
  <div class="flex justify-center items-center p-2">
    <div class="px-7 md:px-9">
      <!-- 日历模式的年月切换器 -->
      <div v-if="isCalendarMode" class="flex items-center justify-between mb-2">
        <button @click="previousMonth" class="p-1 rounded transition-colors text-[var(--text-color-400)] hover:text-[var(--text-color-600)] hover:bg-[var(--bg-color-100)]">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
          </svg>
        </button>
        <span 
          class="text-lg font-medium cursor-pointer transition-colors select-none text-[var(--text-color-500)] hover:text-[var(--text-color-700)]" 
          @click="handleFilterByYearMonth"
          @mousedown="startLongPress"
          @mouseup="cancelLongPress"
          @mouseleave="cancelLongPress"
          @touchstart.prevent="startLongPress"
          @touchend="cancelLongPress"
          @touchcancel="cancelLongPress"
          title="单击筛选本月内容，长按切换到30天视图"
        >
          {{ currentYear }}年{{ currentMonth }}月
        </span>
        <button @click="nextMonth" class="p-1 rounded transition-colors text-[var(--text-color-400)] hover:text-[var(--text-color-600)] hover:bg-[var(--bg-color-100)]">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
          </svg>
        </button>
      </div>



      <!-- 日历热力图 -->
      <div v-if="isCalendarMode" class="grid grid-cols-7 gap-[6px]">
        <!-- 星期标题 -->
        <div v-for="day in weekDays" :key="day" class="w-7 h-6 flex items-center justify-center text-xs text-gray-500">
          {{ day }}
        </div>
        
        <!-- 日期方块 -->
        <div
          v-for="(day, index) in calendarDays"
          :key="index"
          class="relative w-7 h-7 rounded-[6px] transition-colors duration-300 ease ring-1 ring-[var(--heatmap-ring-color)] hover:ring-[var(--ring-color-300)] hover:shadow-sm cursor-pointer"
          :class="{
            'opacity-30': !day.isCurrentMonth
          }"
          :style="{ backgroundColor: getColor(day.count) }"
          @click="handleFilterByDate(day.date)"
          @mouseenter="showCalendarTooltip(day, $event)"
          @mouseleave="hideTooltip"
        ></div>
      </div>

      <!-- 原始30天热力图 -->
      <div 
        v-else 
        class="flex" 
        @mousedown="startLongPress"
        @mouseup="cancelLongPress"
        @mouseleave="cancelLongPress"
        @touchstart.prevent="startLongPress"
        @touchend="cancelLongPress"
        @touchcancel="cancelLongPress"
        title="长按切换到日历视图"
      >
        <div v-for="col in 10" :key="col" class="flex flex-col gap-1 mr-1">
          <div
            v-for="row in 3"
            :key="row"
            class="relative w-5 h-5 rounded-[6px] transition-colors duration-300 ease ring-1 ring-[var(--heatmap-ring-color)] hover:ring-[var(--ring-color-300)] hover:shadow-sm cursor-pointer"
            :style="{ backgroundColor: getColor(getCell(row - 1, col - 1)?.count ?? 0) }"
            @click.stop="handleFilterByOriginalDate(row - 1, col - 1)"
            @mouseenter="showOriginalTooltip(row - 1, col - 1, $event)"
            @mouseleave="hideTooltip"
          ></div>
        </div>
      </div>

      <!-- 自定义 tooltip -->
      <div
        v-if="tooltip.visible"
        class="absolute z-50 px-2 py-1 bg-orange-500 text-white text-xs rounded shadow"
        :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
      >
        {{ tooltip.text }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { fetchGetHeatMap } from '@/service/api'
import { useEchoStore } from '@/stores/echo'

const echoStore = useEchoStore()

interface CalendarDay {
  date: string
  count: number
  isCurrentMonth: boolean
  day: number
}

const HEATMAP_MODE_KEY = 'heatmap-mode'

const originalHeatmapData = ref<App.Api.Ech0.HeatMap>([]) // 近30天数据
const calendarHeatmapData = ref<App.Api.Ech0.HeatMap>([]) // 日历月份数据
const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth() + 1)
const isCalendarMode = ref(localStorage.getItem(HEATMAP_MODE_KEY) === 'calendar')

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

// 创建日历数据的映射
const dataMap = computed(() => {
  const map = new Map<string, number>()
  calendarHeatmapData.value.forEach(item => {
    if (item.date) {
      map.set(item.date, item.count)
    }
  })
  return map
})

// 生成日历天数
const calendarDays = computed(() => {
  const year = currentYear.value
  const month = currentMonth.value
  const firstDay = new Date(year, month - 1, 1)
  const lastDay = new Date(year, month, 0)
  const daysInMonth = lastDay.getDate()
  const startWeekDay = firstDay.getDay()
  
  const days: CalendarDay[] = []
  
  // 添加上个月的尾部天数
  const prevMonth = month === 1 ? 12 : month - 1
  const prevYear = month === 1 ? year - 1 : year
  const prevMonthLastDay = new Date(prevYear, prevMonth, 0).getDate()
  
  for (let i = startWeekDay - 1; i >= 0; i--) {
    const day = prevMonthLastDay - i
    const dateStr = `${prevYear}-${String(prevMonth).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    days.push({
      date: dateStr,
      count: dataMap.value.get(dateStr) || 0,
      isCurrentMonth: false,
      day
    })
  }
  
  // 添加当前月的天数
  for (let day = 1; day <= daysInMonth; day++) {
    const dateStr = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    days.push({
      date: dateStr,
      count: dataMap.value.get(dateStr) || 0,
      isCurrentMonth: true,
      day
    })
  }
  
  // 计算需要的总行数，只补齐到完整的周
  const totalCells = startWeekDay + daysInMonth
  const neededRows = Math.ceil(totalCells / 7)
  const totalNeededCells = neededRows * 7
  
  // 添加下个月的开头天数，只补齐到需要的行数
  const nextMonth = month === 12 ? 1 : month + 1
  const nextYear = month === 12 ? year + 1 : year
  let nextDay = 1
  
  while (days.length < totalNeededCells) {
    const dateStr = `${nextYear}-${String(nextMonth).padStart(2, '0')}-${String(nextDay).padStart(2, '0')}`
    days.push({
      date: dateStr,
      count: dataMap.value.get(dateStr) || 0,
      isCurrentMonth: false,
      day: nextDay
    })
    nextDay++
  }
  
  return days
})

const getColor = (count: number): string => {
  if (count >= 4) return 'var(--heatmap-bg-color-4)'
  if (count >= 3) return 'var(--heatmap-bg-color-3)'
  if (count >= 2) return 'var(--heatmap-bg-color-2)'
  if (count >= 1) return 'var(--heatmap-bg-color-1)'
  return 'var(--heatmap-bg-color-0)'
}

// 月份切换
const previousMonth = () => {
  if (currentMonth.value === 1) {
    currentMonth.value = 12
    currentYear.value--
  } else {
    currentMonth.value--
  }
}

const nextMonth = () => {
  if (currentMonth.value === 12) {
    currentMonth.value = 1
    currentYear.value++
  } else {
    currentMonth.value++
  }
}

// 原始30天热力图逻辑
const grid = computed(() => {
  const cells = [...originalHeatmapData.value]
  const total = 3 * 10
  while (cells.length < total) cells.push({ date: '', count: 0 } as App.Api.Ech0.HeatMap[0])
  const result: (App.Api.Ech0.HeatMap[0] | null)[][] = []
  for (let row = 0; row < 3; row++) {
    result.push(cells.slice(row * 10, (row + 1) * 10))
  }
  return result
})

const getCell = (row: number, col: number) => {
  return grid.value[row]?.[col] ?? null
}

// 模式切换
const toggleMode = () => {
  isCalendarMode.value = !isCalendarMode.value
  localStorage.setItem(HEATMAP_MODE_KEY, isCalendarMode.value ? 'calendar' : 'original')
  // 切换模式时加载对应数据
  if (isCalendarMode.value) {
    loadCalendarHeatMapData()
  } else {
    loadOriginalHeatMapData()
  }
}

// 长按切换视图
const longPressTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const isLongPress = ref(false)
const LONG_PRESS_DURATION = 500 // 长按时间阈值（毫秒）

function startLongPress() {
  isLongPress.value = false
  longPressTimer.value = setTimeout(() => {
    isLongPress.value = true
    toggleMode()
  }, LONG_PRESS_DURATION)
}

function cancelLongPress() {
  if (longPressTimer.value) {
    clearTimeout(longPressTimer.value)
    longPressTimer.value = null
  }
}

// Tooltip 相关
const tooltip = ref({
  visible: false,
  text: '',
  x: 0,
  y: 0,
})

function showCalendarTooltip(day: CalendarDay, event: MouseEvent) {
  tooltip.value.text = `${day.date}: ${day.count} 条`
  tooltip.value.visible = true

  const target = event.target as HTMLElement
  const rect = target.getBoundingClientRect()

  tooltip.value.x = rect.left + rect.width / 2 - 50
  tooltip.value.y = rect.top - 35
}

function showOriginalTooltip(row: number, col: number, event: MouseEvent) {
  const cell = getCell(row, col)
  if (cell) {
    tooltip.value.text = `${cell.date ?? ''}: ${cell.count ?? 0} 条`
    tooltip.value.visible = true

    const target = event.target as HTMLElement
    const rect = target.getBoundingClientRect()

    tooltip.value.x = rect.left
    tooltip.value.y = rect.top - 30
  }
}

function hideTooltip() {
  tooltip.value.visible = false
}

// 点击日期方块筛选
function handleFilterByDate(date: string) {
  if (!date) return
  // 先重置筛选状态
  echoStore.refreshEchosForFilter()
  // 设置日期筛选参数
  echoStore.filteredDate = date
  echoStore.filteredYearMonth = null
  echoStore.filteredTag = null
  echoStore.isDateFilteringMode = true
  // 设置为筛选模式，触发 TheFilteredEchos 组件显示
  echoStore.isFilteringMode = true
}

// 单击年月筛选整月
function handleFilterByYearMonth() {
  // 如果是长按触发的，不执行筛选
  if (isLongPress.value) {
    isLongPress.value = false
    return
  }
  // 先重置筛选状态
  echoStore.refreshEchosForFilter()
  // 设置年月筛选参数
  echoStore.filteredYearMonth = { year: currentYear.value, month: currentMonth.value }
  echoStore.filteredDate = null
  echoStore.filteredTag = null
  echoStore.isDateFilteringMode = true
  // 设置为筛选模式，触发 TheFilteredEchos 组件显示
  echoStore.isFilteringMode = true
}

// 点击30天视图的方块筛选
function handleFilterByOriginalDate(row: number, col: number) {
  const cell = getCell(row, col)
  if (cell && cell.date) {
    handleFilterByDate(cell.date)
  }
}

// 监听年月变化，重新获取日历数据
watch([currentYear, currentMonth], () => {
  if (isCalendarMode.value) {
    loadCalendarHeatMapData()
  }
})

// 加载近30天数据
const loadOriginalHeatMapData = async () => {
  try {
    const res = await fetchGetHeatMap()
    originalHeatmapData.value = res.data
  } catch (error) {
    console.error('Failed to load heatmap data:', error)
  }
}

// 加载指定月份数据
const loadCalendarHeatMapData = async () => {
  try {
    const res = await fetchGetHeatMap(currentYear.value, currentMonth.value)
    calendarHeatmapData.value = res.data
  } catch (error) {
    console.error('Failed to load calendar heatmap data:', error)
  }
}

onMounted(() => {
  if (isCalendarMode.value) {
    loadCalendarHeatMapData()
  } else {
    loadOriginalHeatMapData()
  }
})
</script>
