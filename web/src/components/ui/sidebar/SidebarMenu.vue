<script setup lang="ts">
import type { HTMLAttributes } from 'vue'

import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <ul data-slot="sidebar-menu" :class="cn('flex w-full min-w-0 flex-col gap-1', props.class)">
    <slot />
  </ul>
</template>
