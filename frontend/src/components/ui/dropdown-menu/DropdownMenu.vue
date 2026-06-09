<script setup>
import { DropdownMenuContent, DropdownMenuItem, DropdownMenuPortal, DropdownMenuRoot, DropdownMenuSeparator, DropdownMenuSub, DropdownMenuSubContent, DropdownMenuSubTrigger, DropdownMenuTrigger, DropdownMenuLabel, DropdownMenuGroup } from 'radix-vue'
</script>

<template>
  <DropdownMenuRoot>
    <slot />
  </DropdownMenuRoot>
</template>
