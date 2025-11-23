<template>
  <div class="flex justify-center items-center h-screen">
    <div class="h-1/2 max-w-sm sm:max-w-md md:max-w-lg">
      <h1
        class="text-6xl italic font-bold text-center text-[var(--text-color-next-300)] mb-4 font-serif"
      >
        Ech0
      </h1>
      <!-- 登录  -->
      <div v-if="AuthMode === 'login'">
        <!-- 模式切换 -->
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-bold text-[var(--text-color-next-400)] mb-3">登录</h2>
          <div class="mb-3">
            <button
              @click="AuthMode = 'register'"
              class="text-[var(--text-color-next-500)] hover:text-[var(--text-color-next-700)] transition duration-200"
            >
              <div class="flex flex-row gap-0 items-center">
                注册
                <Arrow class="text-2xl" />
              </div>
            </button>
          </div>
        </div>
        <!-- 账号密码输入 -->
        <BaseInput v-model="username" type="text" placeholder="请输入用户名" class="mb-4" />
        <BaseInput v-model="password" type="password" placeholder="请输入密码" class="mb-4" />
        <div class="flex justify-between items-center px-0.5">
          <BaseButton
            @click="router.push({ name: 'home' })"
            title="返回首页"
            :icon="Home"
            class="rounded-md w-9 h-9"
          />
          <div class="flex items-center">
            <!-- OAuth2 登录 -->
            <BaseButton
              v-if="oauth2Status && oauth2Status.enabled"
              :icon="
                oauth2Status.provider === OAuth2Provider.GITHUB
                  ? Github
                  : oauth2Status.provider === OAuth2Provider.GOOGLE
                    ? Google
                    : oauth2Status.provider === OAuth2Provider.QQ
                      ? QQ
                      : Customoauth
              "
              @click="gotoOAuth2URL"
              class="w-9 h-9 rounded-md mr-2"
            />
            <!-- 账号密码登录 -->
            <BaseButton @click="handleLogin" class="w-12 h-9 rounded-md">
              <span class="text-[var(--text-color-next-500)]">登录</span>
            </BaseButton>
          </div>
        </div>
      </div>
      <!-- 注册 -->
      <div v-else-if="AuthMode === 'register'">
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-bold text-[var(--text-color-next-400)] mb-3">注册</h2>
          <div class="mb-3">
            <button
              @click="AuthMode = 'login'"
              class="text-[var(--text-color-next-500)] hover:text-[var(--text-color-next-700)] transition duration-200"
            >
              <div class="flex flex-row gap-0 items-center">
                登录
                <Arrow class="text-2xl rotate-180" />
              </div>
            </button>
          </div>
        </div>
        <BaseInput v-model="username" type="text" placeholder="请输入用户名" class="mb-4" />
        <BaseInput v-model="password" type="password" placeholder="请输入密码" class="mb-4" />
        <div class="flex justify-between items-center px-0.5">
          <BaseButton
            @click="router.push({ name: 'home' })"
            title="返回首页"
            :icon="Home"
            class="rounded-md w-9 h-9"
          />
          <BaseButton @click="handleRegister" class="rounded-md">
            <span class="text-[var(--text-color-next-500)]">注册</span>
          </BaseButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'
import Arrow from '@/components/icons/arrow.vue'
import Home from '@/components/icons/home.vue'
import Github from '@/components/icons/github.vue'
import Google from '@/components/icons/google.vue'
import QQ from '@/components/icons/qq.vue'
import Customoauth from '@/components/icons/customoauth.vue'
import { fetchGetOAuth2Status } from '@/service/api'
import { OAuth2Provider } from '@/enums/enums'

const AuthMode = ref<'login' | 'register'>('login') // login / register
const username = ref<string>('')
const password = ref<string>('')
const userStore = useUserStore()

const oauth2Status = ref<App.Api.Setting.OAuth2Status | null>(null)
const baseURL =
  import.meta.env.VITE_SERVICE_BASE_URL === '/'
    ? window.location.origin
    : import.meta.env.VITE_SERVICE_BASE_URL
const oauthURL = ref<string>(`${baseURL}/oauth/github/login`)

const gotoOAuth2URL = () => {
  window.location.href = oauthURL.value
}

const getOAuth2Status = async () => {
  const res = await fetchGetOAuth2Status()
  if (res.code === 1) {
    oauth2Status.value = res.data
    oauthURL.value = res.data.provider
      ? `${baseURL}/oauth/${res.data.provider}/login?redirect_uri=${window.location.origin}/auth`
      : `${baseURL}/auth`
  }
}

const router = useRouter()

const handleLogin = async () => {
  // console.log('登录', username.value, password.value)
  await userStore.login({
    username: username.value,
    password: password.value,
  })
}

const handleRegister = async () => {
  // console.log('注册', username.value, password.value)
  if (
    await userStore.signup({
      username: username.value,
      password: password.value,
    })
  ) {
    // 注册成功，切换到登录模式
    AuthMode.value = 'login'
  }
}

onMounted(async () => {
  const url = new URL(window.location.href)
  const token = url.searchParams.get('token')
  if (token) {
    console.log('检测到 token，尝试使用 token 登录', token)
    // 有 token，直接登录
    await userStore.loginWithToken(token)
    return
  }
  getOAuth2Status()
})
</script>
