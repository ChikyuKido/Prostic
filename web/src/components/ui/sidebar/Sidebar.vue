<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { computed } from 'vue'

import { cn } from '@/lib/utils'
import { useSidebar } from './sidebar'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
const { open } = useSidebar()

const sidebarClass = computed(() =>
  cn(
    'bg-sidebar text-sidebar-foreground sticky top-0 hidden h-screen shrink-0 overflow-hidden border-r border-sidebar-border md:flex md:flex-col transition-[width] duration-200 ease-linear',
    open.value ? 'w-72' : 'w-16',
    props.class,
  ),
)
</script>

<template>
  <aside data-slot="sidebar" :data-state="open ? 'expanded' : 'collapsed'" :class="sidebarClass">
    <slot />
  </aside>
</template>
