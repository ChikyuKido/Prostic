<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppSidebar from '@/components/AppSidebar.vue'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
</script>

<template>
  <SidebarProvider class="min-h-screen">
    <AppSidebar />
    <SidebarInset class="min-w-0">
      <header class="flex h-14 shrink-0 items-center border-b border-border/60 bg-background/85 px-4 backdrop-blur-sm">
        <SidebarTrigger class="-ml-1" />
      </header>
      <div class="flex flex-1 flex-col p-4 md:p-6">
        <RouterView />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
