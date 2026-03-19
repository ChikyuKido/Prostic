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
    data-slot="card-title"
    :class="cn('leading-none font-semibold tracking-tight', props.class)"
  >
    <slot />
  </div>
</template>
