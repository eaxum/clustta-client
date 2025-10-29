<template>
  <div class="form-group">
    <label v-if="label" class="form-label">{{ label }}</label>
    <div class="form-input-wrapper">
      <div class="form-input-container">
        <input
          :type="type"
          :value="modelValue"
          @input="handleInput"
          :disabled="disabled"
          :placeholder="placeholder"
          class="form-input"
          :class="{ 'has-icon': showValidation }"
        />
        <div v-if="showValidation" class="form-input-icon">
          <img 
            v-if="error" 
            class="alert-icons" 
            :src="getAppIcon('alert')" 
            alt="Error"
          />
          <img 
            v-else-if="loading" 
            class="alert-icons loading-icon" 
            src="/icons/loading.svg"
            alt="Loading"
          />
          <img 
            v-else-if="valid" 
            class="alert-icons" 
            :src="getAppIcon('circle-check')"
            alt="Valid"
          />
        </div>
      </div>
      <InputAlert v-if="error" :show="!!error" :message="error" />
    </div>
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
import InputAlert from '@/instances/common/components/InputAlert.vue';

const iconStore = useIconStore();

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  type: {
    type: String,
    default: 'text'
  },
  placeholder: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  },
  loading: {
    type: Boolean,
    default: false
  },
  valid: {
    type: Boolean,
    default: false
  },
  showValidation: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:modelValue', 'input']);

const handleInput = (event) => {
  emit('update:modelValue', event.target.value);
  emit('input', event.target.value);
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};
</script>

<style scoped>
.form-group {
  margin-bottom: 1rem;
  width: 100%;
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.form-label {
  color: var(--white);
  font-weight: 400;
  font-size: 0.875rem;
  min-width: 120px;
  padding-top: 0.75rem;
  flex-shrink: 0;
}

.form-input-wrapper {
  flex: 1;
  width: 100%;
}

.form-input-container {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.form-input {
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  box-sizing: border-box;
  font-size: 16px;
  border-radius: var(--normal-radius);
  padding: 0.75rem;
  border: 0px;
  outline: none;
  background-color: var(--midnight-steel);
  color: var(--white);
  width: 100%;
  height: 50px;
  transition: opacity 0.2s;
}

.form-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-input:focus {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.form-input::placeholder {
  color: rgba(255, 255, 255, 0.4);
}

.form-input-icon {
  height: 100%;
  width: min-content;
  position: absolute;
  right: .75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.alert-icons {
  width: 20px;
  height: 20px;
  filter: brightness(0) invert(1);
}

.loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
