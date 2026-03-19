<script setup lang="ts">
import type { HTMLAttributes } from 'vue'

import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <div data-slot="sidebar-header" :class="cn('shrink-0 flex flex-col gap-2 p-2', props.class)">
    <slot />
  </div>
</template>
