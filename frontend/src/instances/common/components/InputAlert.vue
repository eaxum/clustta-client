<template>
  <div v-if="show" class="horizontal-flex input-alert">
    {{ message }}
  </div>
</template>

<script setup>

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  message: {
    type: String,
    required: true
  }
});
</script>

<style scoped>
.horizontal-flex {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  width: 100%;
}

.input-alert {
  color: var(--attention);
  font-weight: 400;
  font-size: 14px;
  padding: 0.25rem .5rem;
}
</style>
