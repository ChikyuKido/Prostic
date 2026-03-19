<script setup lang="ts">
import type { HTMLAttributes } from 'vue'

import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <div data-slot="sidebar-inset" :class="cn('bg-background flex min-w-0 flex-1 flex-col', props.class)">
    <slot />
  </div>
</template>
