<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="'stall'" :showSearch="showSearch" />
    <div class="general-container">

      <div class="studio-info-text">
        <p>Studios are spaces for your projects and teams to collaborate. Select how you would like to create your studio</p>
      </div>

      <div class="studio-types-container">
        <div 
          v-for="studioType in studioTypes" 
          :key="studioType.type" 
          :class="['studio-type-card', { 'studio-type-card-selected': selectedStudioType === studioType.type }]"
          @click="selectStudioType(studioType.type)"
        >
          <div class="studio-icon">
            <img class="small-icons" :src="getAppIcon(studioType.icon)" :alt="studioType.title" />
          </div>
          <div class="studio-details">
            <div class="studio-type-title">
              {{ studioType.title }}
              <span v-if="studioType.beta" class="beta-badge">BETA</span>
            </div>
            <div class="studio-type-description">
              {{ studioType.message }}
            </div>
          </div>
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Next'" :fullWidth="true" :buttonFunction="handleNext" :isActive="true" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';

// constants
const showSearch = false;
const title = 'New Studio';

// studio types data
const studioTypes = ref([
  {
    type: 'clustta-cloud',
    title: 'ClusttaCloud',
    message: 'Get started instantly with our managed cloud service. No setup required, automatic updates and enterprise-grade security.',
    icon: 'clustta',
    beta: true
  },
  {
    type: 'self-managed',
    title: 'Self Managed',
    message: 'Host and manage your own Clustta studio instance with full control over data, security, and customization.',
    icon: 'two-drives'
  }
]);

// refs
const modalContainer = ref(null);
const selectedStudioType = ref('clustta-cloud');
const selectedTemplate = ref('');

// methods

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles the next button click to open the appropriate config modal.
const handleNext = () => {
  if (selectedStudioType.value === 'self-managed') {
    modals.setModalVisibility('configSelfManagedStudioModal', true);
  } else if (selectedStudioType.value === 'clustta-cloud') {
    modals.setModalVisibility('configClusttaCloudStudioModal', true);
  }
};

// Selects a studio type.
const selectStudioType = (type) => {
  selectedStudioType.value = type;
};

// lifecycle
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

.studio-info-text {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  color: var(--white);
  font-size: 14px;
  padding: .5rem 0;
  box-sizing: border-box;
}

.studio-info-text p {
  margin: 0;
}

.studio-types-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  box-sizing: border-box;
  overflow: hidden;
}

/* Studio type card styles */
.studio-type-card {
  width: 100%;
  box-sizing: border-box;
  background-color: var(--dark-steel);
  display: flex;
  align-items: flex-start;
  justify-content: flex-start;
  border-radius: 8px;
  gap: .5rem;
  padding: 1rem .8rem;
  outline: var(--transparent-line);
  outline-offset: -1px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.studio-type-card:hover {
  background-color: #ffffff15;
  background-color: var(--steel);
  /* outline: var(--solid-line); */
}

.studio-type-card-selected {
  background-color: var(--steel);
  outline: var(--solid-line);
}

.studio-type-card-selected:hover {
  outline: var(--solid-line);
}


.studio-icon {
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.studio-icon img {
  width: 20px;
  height: 20px;
}

.studio-details {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  width: 100%;
}

.studio-type-title {
  font-weight: 400;
  color: var(--white);
  font-size: 16px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.beta-badge {
  display: inline-block;
  padding: 0.15rem 0.4rem;
  background-color: #2D9CDB;
  color: var(--white);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
  border-radius: 4px;
  text-transform: uppercase;
}

.studio-type-description {
  font-size: 13px;
  color: var(--white);
  line-height: 1.4;
}

.general-container{
  gap: 1rem;
  /* padding: 1rem 0; */
  /* padding-bottom: 1rem; */
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>

