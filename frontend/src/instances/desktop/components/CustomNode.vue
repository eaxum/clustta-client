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
  border-radius: 8px;
  /* background: #fff; */
  color: black;
  border: 1px solid #1a192b;
  text-align: center;
}


.custom-node {
  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: tomato; */
  font-size: 12px;
  overflow: hidden;
  text-align: center;
  position: relative;
  width: 150px;
  height: 70px;
  padding: 10px;
  color: black;
}
</style>