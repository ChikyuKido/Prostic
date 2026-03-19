<script setup lang="ts">
import { DialogClose as DialogClosePrimitive, type DialogCloseProps } from 'reka-ui'

const props = defineProps<DialogCloseProps>()
</script>

<template>
  <DialogClosePrimitive v-bind="props">
    <slot />
  </DialogClosePrimitive>
</template>
