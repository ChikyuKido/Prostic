<script setup lang="ts">
import { ref } from 'vue'
import { RouterView } from 'vue-router'

import AppSidebar from '@/components/AppSidebar.vue'

const activeView = ref<{ refresh?: () => void | Promise<void> } | null>(null)

function setActiveView(value: unknown) {
  activeView.value = value as { refresh?: () => void | Promise<void> } | null
}

function refreshCurrentView() {
  const refresh = activeView.value?.refresh
  if (refresh) {
    void refresh()
  }
}
</script>

<template>
  <div class="grid min-h-[calc(100vh-5rem)] w-full max-w-7xl gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
    <AppSidebar @refreshed="refreshCurrentView" />
    <div class="min-w-0">
      <RouterView v-slot="{ Component }">
        <component :is="Component" :ref="setActiveView" />
      </RouterView>
    </div>
  </div>
</template>
