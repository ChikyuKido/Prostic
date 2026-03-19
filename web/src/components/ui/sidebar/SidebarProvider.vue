<script setup lang="ts">
import type { HTMLAttributes } from 'vue'

import { cn } from '@/lib/utils'
import { provideSidebar } from './sidebar'

interface Props {
  class?: HTMLAttributes['class']
  defaultOpen?: boolean
  storageKey?: string
}

const props = withDefaults(defineProps<Props>(), {
  defaultOpen: true,
  storageKey: 'sidebar',
})

provideSidebar({
  defaultOpen: props.defaultOpen,
  storageKey: props.storageKey,
})
</script>

<template>
  <div
    data-slot="sidebar-provider"
    class="group/sidebar-wrapper flex min-h-screen w-full"
    :class="cn(props.class)"
  >
    <slot />
  </div>
</template>
