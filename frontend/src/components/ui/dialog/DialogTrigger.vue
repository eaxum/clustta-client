<script setup>
import { DialogTrigger } from 'radix-vue'
</script>

<template>
  <DialogTrigger as-child>
    <slot />
  </DialogTrigger>
</template>
