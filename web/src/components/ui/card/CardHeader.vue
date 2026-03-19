<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <div
    data-slot="card-header"
    :class="cn('grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6', props.class)"
  >
    <slot />
  </div>
</template>
