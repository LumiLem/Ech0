<template>
  <div
    class="bg-[var(--editor-bg-color)] rounded-md ring-1 ring-[var(--ring-color)] ring-inset mx-auto shadow-xs hover:shadow-sm"
  >
    <div class="mx-auto w-full px-3 py-4">
      <!-- The Title && Nav -->
      <TheTitleAndNav />

      <div class="mt-8 px-3 gap-2 font-serif flex justify-between items-baseline">
        <div class="text-[var(--theboard-border-color)] text-lg">
          <span class="font-bold">欢迎光临!</span>
        </div>
        <div class="text-[var(--text-color-300)] text-xs italic">
          Powered by <a href="https://github.com/lin-snow/Ech0">Ech0</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheTitleAndNav from './TheEditor/TheTitleAndNav.vue'
</script>
