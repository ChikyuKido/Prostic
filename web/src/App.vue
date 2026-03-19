<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <main class="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(35,46,69,0.16),_transparent_45%),linear-gradient(180deg,_hsl(210_20%_98%),_hsl(210_25%_94%))] px-6 py-10 dark:bg-[radial-gradient(circle_at_top,_rgba(76,99,140,0.18),_transparent_40%),linear-gradient(180deg,_hsl(222_30%_10%),_hsl(222_30%_7%))]">
    <div class="mx-auto flex min-h-[calc(100vh-5rem)] w-full max-w-6xl items-center justify-center">
      <RouterView />
    </div>
  </main>
</template>
