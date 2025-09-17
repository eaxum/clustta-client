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
import { ref, onMounted } from 'vue';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import AppsGrid from '@/instances/common/components/AppsGrid.vue';
import { useTemplateStore } from '@/stores/template';

// states
const trayStates = useTrayStates();

// stores
const modals = useDesktopModalStore();
const templateStore = useTemplateStore();

// refs
const showSearch = false;
const selectedTemplate = ref('');
const modalContainer = ref(null);

// computed properties
const title = 'Select Asset Template';
const icon = '/icons/new_task.svg';


// onMounted hook
onMounted(() => {
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];

  if (templateStore.lastUsedTemplate) {
    selectedTemplate.value = templateStore.lastUsedTemplate;
  } else {
    selectedTemplate.value = templateStore.templates[0]?.name;
  };

});


</script>


<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  background-color: red;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 100px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1.5rem;
}

.input-short {
  width: 100%;
}

.listbox-short {
  width: 130px;
}

.input-label {
  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 1rem;

}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: space-between;
}


.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>

