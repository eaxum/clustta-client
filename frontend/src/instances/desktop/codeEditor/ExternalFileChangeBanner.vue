<template>
  <div v-if="document?.externalChanged || document?.saveError" class="external-change-banner">
    <span>{{ message }}</span>
    <template v-if="document.externalChanged">
      <ActionButton label="Reload" :showLabel="true" :showIcon="false" :buttonFunction="reload" />
      <ActionButton label="Keep Local" :showLabel="true" :showIcon="false" :buttonFunction="keepLocal" />
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';

const codeEditor = useCodeEditorStore();
const document = computed(() => codeEditor.activeDocument);
const message = computed(() => document.value?.saveError || 'This file changed on disk.');
const reload = () => codeEditor.reloadDocument();
const keepLocal = () => codeEditor.keepLocalVersion();
</script>

<style scoped>
.external-change-banner {
  display: flex;
  align-items: center;
  gap: .5rem;
  min-height: 34px;
  padding: 0 .75rem;
  background-color: var(--warning);
  color: var(--text-inverse);
}
</style>
