import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { localStg } from '@/utils/storage'

export type ThemeMode = 'system' | 'light' | 'dark'
type ThemeType = 'light' | 'dark'

export const useThemeStore = defineStore('themeStore', () => {
  const savedThemeMode = localStg.getItem('themeMode')
  const savedTheme = localStg.getItem('theme')

  // 初始化 themeMode
  const mode = ref<ThemeMode>(
    savedThemeMode === 'system' || savedThemeMode === 'light' || savedThemeMode === 'dark'
      ? savedThemeMode
      : 'system',
  )
  const theme = ref<ThemeType>(
    savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : 'light',
  )

  const toggleTheme = () => {
    if (mode.value === 'system') {
      mode.value = 'light'
    } else if (mode.value === 'light') {
      mode.value = 'dark'
    } else {
      mode.value = 'system'
    }

    applyTheme()
    localStg.setItem('themeMode', mode.value)
  }

  const applyTheme = () => {
    switch (mode.value) {
      case 'system':
        theme.value = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
        break
      case 'light':
        theme.value = 'light'
        break
      case 'dark':
        theme.value = 'dark'
        break
    }

    document.documentElement.classList.remove('light', 'dark')
    document.documentElement.classList.add(theme.value)
    localStg.setItem('theme', theme.value)
  }

  const init = () => {
    applyTheme()
    // 监听系统主题变化
    watch(theme, applyTheme)
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme)
  }

  return {
    theme,
    mode,
    toggleTheme,
    applyTheme,
    init,
  }
})