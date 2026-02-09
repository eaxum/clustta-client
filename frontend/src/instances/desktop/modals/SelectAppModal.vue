<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="'file-plus'" :showSearch="showSearch" />
    <div class="general-container">
      <AppsGrid />
    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';

// components
import AppsGrid from '@/instances/common/components/AppsGrid.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';

const modals = useDesktopModalStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();

// refs
const modalContainer = ref(null);
const selectedTemplate = ref('');

// constants
const showSearch = false;
const title = 'Select Asset Template';

// lifecycle hooks
onMounted(() => {
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];
  if (templateStore.lastUsedTemplate) {
    selectedTemplate.value = templateStore.lastUsedTemplate;
  } else {
    selectedTemplate.value = templateStore.templates[0]?.name;
  }
});
</script>


<style scoped>
@import "@/assets/desktop.css";
</style>

