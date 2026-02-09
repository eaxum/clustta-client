<template>
  <div v-if="show" class="horizontal-flex input-alert" :class="alertClass">
    {{ message }}
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  message: {
    type: String,
    required: true
  },
  show: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    default: 'error',
    validator: (value) => ['error', 'info'].includes(value)
  }
});

// computed
const alertClass = computed(() => {
  return {
    'input-alert-error': props.type === 'error',
    'input-alert-info': props.type === 'info'
  };
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
  margin-top: .5rem;
  font-weight: 400;
  font-size: .8rem;
  padding: 0.25rem .5rem;
}

.input-alert-error {
  color: var(--attention);
}

.input-alert-info {
  color: var(--white);
  opacity: 0.6;
}
</style>
