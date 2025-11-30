import { defineStore } from 'pinia'
import { ref } from 'vue'
import { localStg } from '@/utils/storage'

type ThemeType = 'light' | 'dark'

export const useThemeStore = defineStore('themeStore', () => {
  const savedTheme = localStg.getItem('theme')
  const systemPrefersDark =
    window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  const theme = ref<ThemeType>(
    savedTheme === 'light' || savedTheme === 'dark'
      ? savedTheme
      : systemPrefersDark
        ? 'dark'
        : 'light',
  )

  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    applyTheme(theme.value)
  }

  const applyTheme = (t: ThemeType) => {
    document.documentElement.classList.remove('light', 'dark')
    document.documentElement.classList.add(t)
    localStg.setItem('theme', t)
  }

  const init = () => {
    applyTheme(theme.value)
  }

  return {
    theme,
    toggleTheme,
    applyTheme,
    init,
  }
})
