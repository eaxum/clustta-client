<template>
  <div class="rename-input">
    <input 
      spellcheck="false"
      v-model="localValue"
      @keydown.enter="handleEnterKey"
      @keydown.esc="handleEscKey"
      class="input-short"
      type="text"
      :placeholder="placeholder"
      v-focus
    />
    <ActionButton 
      :isDisabled="!isNameChanged" 
      :icon="getAppIcon('check')" 
      v-tooltip="{ text: $t('components.renameInput.confirm'), shortcut: 'confirm' }"
      @click="handleConfirm" 
    />
    <ActionButton 
      :icon="getAppIcon('close')" 
      v-tooltip="{ text: $t('components.renameInput.cancel'), shortcut: 'cancel' }"
      @click="handleCancel" 
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useIconStore } from '@/stores/icons';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

const iconStore = useIconStore();

const props = defineProps({
  modelValue: {
    type: String,
    required: true
  },
  originalValue: {
    type: String,
    required: true
  },
  placeholder: {
    type: String,
    default: 'Enter name'
  }
});

// Note: placeholder prop default not i18n'd since it's a prop default

const emit = defineEmits(['update:modelValue', 'confirm', 'cancel']);

const localValue = ref(props.modelValue);

// Watch for external changes to modelValue
watch(() => props.modelValue, (newValue) => {
  localValue.value = newValue;
});

// Emit changes back to parent
watch(localValue, (newValue) => {
  emit('update:modelValue', newValue);
});

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const isNameChanged = computed(() => {
  const restrictedEntries = [props.originalValue, ''];
  const lowerCaseEditableName = localValue.value.toLowerCase();
  const lowerCaseRestrictedEntries = restrictedEntries.map(entry =>
    typeof entry === 'string' ? entry.toLowerCase() : entry
  );
  return !lowerCaseRestrictedEntries.includes(lowerCaseEditableName);
});

const handleEnterKey = () => {
  if (!isNameChanged.value) {
    const inputElement = document.querySelector('.rename-input .input-short');
    if (inputElement) {
      inputElement.classList.add('shake');
      setTimeout(() => {
        inputElement.classList.remove('shake');
      }, 300);
    }
  } else {
    handleConfirm();
  }
};

const handleEscKey = () => {
  emit('cancel');
};

const handleConfirm = () => {
  if (isNameChanged.value) {
    emit('confirm', localValue.value);
  }
};

const handleCancel = () => {
  emit('cancel');
};
</script>

<style scoped>
.rename-input {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  width: 100%;
  box-sizing: border-box;
}

.rename-input .input-short {
  flex: 1;
  min-width: 0;
  padding: 0.4rem 0.5rem;
  box-sizing: border-box;
  background: var(--bg);
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--normal-radius);
  color: var(--text);
  font-size: 14px;
  /* max-height: 50% !important;
  height: 30px; */
  height: 100%;
  font-family: Inter, sans-serif;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  10%, 30%, 50%, 70%, 90% { transform: translateX(-2px); }
  20%, 40%, 60%, 80% { transform: translateX(2px); }
}

.shake {
  animation: shake 0.3s ease-in-out;
}
</style>
