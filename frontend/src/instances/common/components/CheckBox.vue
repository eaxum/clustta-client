<template>
  <label class="checkbox-container">
    <input 
      v-model="checkboxValue" 
      class="checkbox-input" 
      type="checkbox" 
      @change="handleChange"
      :disabled="disabled"
      :indeterminate="indeterminate"
      :aria-label="ariaLabel"
    />
    <span class="checkbox-box" :class="{ 'checked': checkboxValue, 'indeterminate': indeterminate }">
      <span v-if="checkboxValue" class="checkbox-checkmark"></span>
      <span v-else-if="indeterminate" class="checkbox-indeterminate"></span>
    </span>
  </label>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  indeterminate: {
    type: Boolean,
    default: false
  },
  ariaLabel: {
    type: String,
    default: 'Select item'
  }
});

const emit = defineEmits(['update:modelValue', 'change']);

const checkboxValue = computed({
  get() {
    return props.modelValue;
  },
  set(value) {
    emit('update:modelValue', value);
  }
});

const handleChange = () => {
  emit('change', checkboxValue.value);
};
</script>

<style scoped>
.checkbox-container {
  position: relative;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
}

.checkbox-container:has(.checkbox-input:disabled) {
  cursor: not-allowed;
  opacity: 0.5;
}

.checkbox-input {
  pointer-events: none;
  position: absolute;
  width: 10px;
  height: 10px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

.checkbox-box {
  --checkbox-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  width: var(--checkbox-size);
  height: var(--checkbox-size);
  box-sizing: border-box;
  border-radius: 10px;
  background-color: var(--surface-1);
  transition: all 0.2s ease-in-out;
  border: 1.5px solid var(--border-strong);
}

.checkbox-box.checked,
.checkbox-box.indeterminate {
  background-color: var(--accent);
  border-color: var(--accent);
}

.checkbox-container:hover .checkbox-box {
  border-color: var(--accent);
}

.checkbox-checkmark {
  width: 5px;
  height: 9px;
  margin-top: -2px;
  border: solid var(--accent-fg);
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  border-bottom-right-radius: 3px;
  animation: checkmarkAppear 0.2s ease-in-out;
}

.checkbox-indeterminate {
  width: 9px;
  height: 2px;
  border-radius: 1px;
  background-color: var(--accent-fg);
}

@keyframes checkmarkAppear {
  from {
    transform: rotate(45deg) scale(0);
    opacity: 0;
  }
  to {
    transform: rotate(45deg) scale(1);
    opacity: 1;
  }
}
</style>
