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

const labelClass = computed(() =>
  cn(
    'text-sidebar-foreground/60 flex h-8 shrink-0 items-center rounded-md px-2 text-xs font-medium uppercase tracking-[0.16em]',
    !open.value && 'justify-center px-0 text-[0px]',
    props.class,
  ),
)
</script>

<template>
  <div data-slot="sidebar-group-label" :class="labelClass">
    <slot />
  </div>
</template>
