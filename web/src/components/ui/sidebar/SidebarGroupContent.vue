<script setup lang="ts">
import type { HTMLAttributes } from 'vue'

import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <div data-slot="sidebar-group-content" :class="cn('w-full text-sm', props.class)">
    <slot />
  </div>
</template>
