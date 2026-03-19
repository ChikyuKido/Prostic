<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { DialogOverlay as DialogOverlayPrimitive, type DialogOverlayProps } from 'reka-ui'

import { cn } from '@/lib/utils'

interface Props extends DialogOverlayProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <DialogOverlayPrimitive
    v-bind="props"
    :class="cn('fixed inset-0 z-50 bg-black/55 backdrop-blur-[2px]', props.class)"
  />
</template>
