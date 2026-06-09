<template>
  <div class="form-group" :class="{ 'form-group-vertical': labelTop }">
    <label v-if="label" class="form-label">{{ label }}</label>
    <div class="form-input-wrapper">
      <div class="form-input-container">
        <input ref="inputRef" :type="inputType" :value="modelValue" @input="handleInput" :disabled="disabled"
          :placeholder="placeholder" class="form-input" :class="{ 'has-icon': showValidation || isSecret }" />
        <div v-if="needsValidation && showValidation" class="form-input-icon">
          <ActionButton v-if="error" :icon="getAppIcon('alert')" :isInactive="true" useAlert :showLabel="false" />
          <ActionButton v-else-if="loading" :icon="getAppIcon('loading')" :isInactive="true" :isLoading="true" :showLabel="false" />
          <ActionButton v-else-if="valid" :icon="getAppIcon('circle-check')" :isInactive="true" useGo :showLabel="false" />
        </div>
        <div v-if="isSecret && modelValue" class="form-input-icon">
          <ActionButton
            v-tooltip="isSecretVisible ? $t('components.formInput.hide') : $t('components.formInput.show')"
            :icon="isSecretVisible ? getAppIcon('eye-cancel') : getAppIcon('eye')"
            :buttonFunction="toggleSecretVisibility"
            :showLabel="false"
          />
        </div>
      </div>
      <InputAlert v-if="error" :show="!!error" :message="error" type="error" />
      <InputAlert v-else-if="info" :show="!!info" :message="info" type="info" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

const props = defineProps({
  autofocus: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  },
  info: {
    type: String,
    default: ''
  },
  isSecret: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  },
  labelTop: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  },
  modelValue: {
    type: String,
    default: ''
  },
  needsValidation: {
    type: Boolean,
    default: false
  },
  placeholder: {
    type: String,
    default: ''
  },
  showValidation: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    default: 'text'
  },
  valid: {
    type: Boolean,
    default: false
  }
});

// refs
const inputRef = ref(null);
const isSecretVisible = ref(false);

// computed
const inputType = computed(() => {
  if (props.isSecret) {
    return isSecretVisible.value ? 'text' : 'password';
  }
  return props.type;
});

const emit = defineEmits(['update:modelValue', 'input']);

// methods

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles input events and emits the new value.
const handleInput = (event) => {
  emit('update:modelValue', event.target.value);
  emit('input', event.target.value);
};

// Toggles the visibility of secret input fields.
const toggleSecretVisibility = () => {
  isSecretVisible.value = !isSecretVisible.value;
};

// lifecycle hooks
onMounted(() => {
  if (props.autofocus && inputRef.value) {
    inputRef.value.focus();
  }
});
</script>

<style scoped>
.form-group {
  margin-bottom: .75rem;
  width: 100%;
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.form-group-vertical {
  flex-direction: column;
  gap: 0.375rem;
}

.form-group-vertical .form-label {
  min-width: unset;
  padding-top: 0;
  padding-left: 0;
}

.form-label {
  color: hsl(var(--foreground));
  font-weight: 500;
  font-size: 0.875rem;
  min-width: 120px;
  padding-top: 0.625rem;
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
  box-sizing: border-box;
  border-radius: calc(var(--radius) - 2px);
  border: 1px solid hsl(var(--input));
  background-color: transparent;
  padding-right: .25rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.form-input {
  font-family: 'Inter', sans-serif;
  font-weight: 400;
  box-sizing: border-box;
  font-size: 0.875rem;
  padding: 0.5rem 0.75rem;
  border: 0px;
  outline: none;
  background-color: transparent;
  color: hsl(var(--foreground));
  width: 100%;
  height: 36px;
  transition: opacity 0.2s;
}

.form-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-input::placeholder {
  color: hsl(var(--muted-foreground));
}

.form-input-container:hover {
  border-color: hsl(var(--ring));
}

.form-input-container:focus-within {
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 1px hsl(var(--ring));
}

.form-input-icon {
  height: 100%;
  width: min-content;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
