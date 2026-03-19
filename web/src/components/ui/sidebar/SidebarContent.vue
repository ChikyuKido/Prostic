<script setup lang="ts">
import type { HTMLAttributes } from 'vue'

import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <div data-slot="sidebar-content" :class="cn('flex min-h-0 flex-1 flex-col gap-2 overflow-x-hidden overflow-y-auto p-2', props.class)">
    <slot />
  </div>
</template>
