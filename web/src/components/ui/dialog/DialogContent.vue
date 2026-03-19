<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import {
  DialogContent as DialogContentPrimitive,
  DialogPortal as DialogPortalPrimitive,
  type DialogContentProps,
  DialogOverlay as DialogOverlayPrimitive,
} from 'reka-ui'

import { cn } from '@/lib/utils'

interface Props extends DialogContentProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <DialogPortalPrimitive>
    <DialogOverlayPrimitive class="fixed inset-0 z-50 bg-black/55 backdrop-blur-[2px]" />
    <DialogContentPrimitive
      v-bind="props"
      :class="cn('fixed left-1/2 top-1/2 z-50 grid w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 gap-4 rounded-xl border border-border bg-background p-6 shadow-2xl', props.class)"
    >
      <slot />
    </DialogContentPrimitive>
  </DialogPortalPrimitive>
</template>
