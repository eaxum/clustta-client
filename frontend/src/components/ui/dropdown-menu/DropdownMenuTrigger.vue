<script setup>
import { DropdownMenuTrigger } from 'radix-vue'
</script>

<template>
  <DropdownMenuTrigger as-child>
    <slot />
  </DropdownMenuTrigger>
</template>
