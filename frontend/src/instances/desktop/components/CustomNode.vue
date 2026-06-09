<!-- CustomNode.vue -->
<template>
  <div class="custom-node-root">
    <Handle type="source" :position="Position.Left" />
    <Handle type="target" :position="Position.Right" />
  </div>
</template>

<script setup>

import { Handle } from '@vue-flow/core';
import { Position } from '@vue-flow/core';

const props = defineProps({
  data: Object,
});

</script>

<style scoped>
.custom-node-root {
  text-align: center;
  position: relative;
  height: min-content;
  width: min-content;
  /* padding: .2rem; */
  border-radius: var(--normal-radius);
  /* background: #fff; */
  color: hsl(var(--background));
  border: 1px solid hsl(var(--border));
  text-align: center;
}


.custom-node {
  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: hsl(var(--destructive)); */
  font-size: 12px;
  overflow: hidden;
  text-align: center;
  position: relative;
  width: 150px;
  height: 70px;
  padding: 10px;
  color: hsl(var(--background));
}
</style>

