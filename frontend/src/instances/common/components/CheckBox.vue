<template>
  <label class="checkbox-container">
    <input 
      v-model="checkboxValue" 
      class="checkbox-input" 
      type="checkbox" 
      @change="handleChange"
      :disabled="disabled"
    />
    <span class="checkbox-box" :class="{ 'checked': checkboxValue }">
      <span v-if="checkboxValue" class="checkbox-checkmark"></span>
    </span>
  </label>
</template>

<script setup>
import { ref, computed, watch } from "vue";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
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
  --primaryDarkestColor: rgba(0, 0, 0, 0.384);
  --checkedColor: var(--task-item-selected);
  height: max-content;
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
  border-radius: 6px;
  /* background-color: var(--primaryDarkestColor); */
  transition: all 0.2s ease-in-out;
  border: 2px solid transparent;
}

.checkbox-box.checked {
  background-color: var(--checkedColor);
  border-color: var(--checkedColor);
}

.checkbox-input:focus + .checkbox-box {
  outline: 2px solid var(--checkedColor);
  outline-offset: 2px;
}

.checkbox-checkmark {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  background-color: white;
  /* background-color: var(--checkedColor); */
  animation: checkmarkAppear 0.2s ease-in-out;
}

@keyframes checkmarkAppear {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
</style>
