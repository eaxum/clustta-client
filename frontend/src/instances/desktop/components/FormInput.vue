<template>
  <div class="form-group" :class="{ 'form-group-vertical': labelTop }">
    <label v-if="label" class="form-label">{{ label }}</label>
    <div class="form-input-wrapper">
      <div class="form-input-container">
        <input
          :type="inputType"
          :value="modelValue"
          @input="handleInput"
          :disabled="disabled"
          :placeholder="placeholder"
          class="form-input"
          :class="{ 'has-icon': showValidation || isSecret }"
        />
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
import { computed, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

const props = defineProps({
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
</script>

<style scoped>
.form-group {
  margin-bottom: .8rem;
  width: 100%;
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.form-group-vertical {
  flex-direction: column;
  gap: 0.5rem;
}

.form-group-vertical .form-label {
  min-width: unset;
  padding-top: 0;
  /* background-color: crimson; */
  padding-left: .5rem;
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
  box-sizing: border-box;
  border-radius: var(--large-radius);
  background-color: var(--midnight-steel);
  padding-right: .25rem;
}

.form-input {
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  box-sizing: border-box;
  font-size: 14px;
  padding: 0.75rem;
  border: 0px;
  outline: none;
  background-color: transparent;
  color: var(--white);
  width: 100%;
  height: 40px;
  transition: opacity 0.2s;
}

.form-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-input::placeholder {
  color: var(--white);
  opacity: .5;
}

.form-input-container:hover,
.form-input-container:focus-within {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.form-input-icon {
  height: 100%;
  width: min-content;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
