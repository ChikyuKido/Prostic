<script setup lang="ts">
import { PanelLeft } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { useSidebar } from './sidebar'

const { toggle } = useSidebar()
</script>

<template>
  <Button size="icon" variant="ghost" @click="toggle">
    <PanelLeft class="size-4" />
    <span class="sr-only">Toggle Sidebar</span>
  </Button>
</template>
