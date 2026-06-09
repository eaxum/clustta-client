<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <div class="ignore-list-header">
        <p class="ignore-description">
          {{ $t('settings.ignoreListDescriptionPre') }}
          <Chip :label="'*.log'" :isStatic="true" :color="'hsl(var(--accent))'" />
          {{ $t('settings.ignoreListDescriptionMid') }}
          <Chip :label="'/FolderName'" :isStatic="true" :color="'hsl(var(--accent))'" />
          {{ $t('settings.ignoreListDescriptionPost') }}
        </p>

        <div class="template-actions">
          <DropDownBox :items="templateNames" :selectedItem="selectedTemplate" :onSelect="applyTemplate" :placeHolder="$t('settings.selectTemplate')" :fixedWidth="true" />

          <ActionButton :icon="getAppIcon('trash')" :label="$t('settings.clearAll')" @click="prepClearAll" v-tooltip="$t('settings.clearAll')" />
        </div>
      </div>

      <IgnoreListBox :placeholder="$t('placeholders.addIgnoreItem')" :selectedItems="ignoreList" @itemAdded="addIgnoredItem"
        @itemRemoved="removeIgnoredItem" />

    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ProjectService } from '@/services';
import { ignoreTemplates, ignoreTemplateNames } from '@/lib/ignoreTemplates';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Chip from '@/instances/common/components/Chip.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import IgnoreListBox from '@/instances/common/components/IgnoreListBox.vue';

// stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// store imports
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

// refs
const ignoreList = ref([]);
const selectedTemplate = ref('');
const templateNames = ignoreTemplateNames;

// methods
const addIgnoredItem = (item) => {
  if (!ignoreList.value.includes(item)) {
    ignoreList.value.push(item);
    saveIgnoreList();
  }
};

// Prompts the user to confirm applying a preset template to the ignore list.
const applyTemplate = (templateName) => {
  selectedTemplate.value = '';
  const entries = ignoreTemplates[templateName];
  if (!entries) return;
  trayStates.popUpModalTitle = t('settings.applyIgnoreTemplateTitle', { name: templateName });
  trayStates.popUpModalMessage = t('settings.applyIgnoreTemplateMessage');
  trayStates.popUpModalIcon = 'list';
  trayStates.popUpModalFunction = () => mergeTemplateEntries(entries);
  modals.setModalVisibility('popUpModal', true);
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

// Returns the project URI used for saving the ignore list.
const getProjectUri = () => {
  if (projectStore.activeProject.has_remote) {
    return projectStore.getActiveProjectUrl;
  }
  return projectStore.activeProject.uri;
};

// Merges template entries into the current ignore list, skipping duplicates.
const mergeTemplateEntries = (entries) => {
  const newItems = entries.filter((e) => !ignoreList.value.includes(e));
  if (!newItems.length) {
    modals.disableAllModals();
    return;
  }
  ignoreList.value.push(...newItems);
  saveIgnoreList();
  modals.disableAllModals();
};

// Clears all items from the ignore list after confirmation.
const prepClearAll = () => {
  if (!ignoreList.value.length) return;
  trayStates.popUpModalTitle = t('settings.clearIgnoreListTitle');
  trayStates.popUpModalMessage = t('settings.clearIgnoreListMessage');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = () => {
    ignoreList.value = [];
    saveIgnoreList();
    modals.disableAllModals();
  };
  modals.setModalVisibility('popUpModal', true);
};

const removeIgnoredItem = (item) => {
  ignoreList.value = ignoreList.value.filter((removedItem) => removedItem !== item);
  saveIgnoreList();
};

// Persists the current ignore list to the project and updates the store.
const saveIgnoreList = () => {
  const projectUri = getProjectUri();
  ProjectService.SetIgnoreList(projectUri, projectStore.selectedStudio.name, ignoreList.value)
    .then(() => {
      let project = projectStore.activeProject;
      let projectIndex = projectStore.projects.findIndex((p) => p.id === project.id);
      projectStore.projects[projectIndex].ignore_list = ignoreList.value;
      projectStore.activeProject.ignore_list = ignoreList.value;
      projectStore.refreshUntrackedItems();
    })
    .catch(() => {
      notificationStore.addNotification(t('notifications.failedToUpdateIgnoreList'), 'error');
    });
};

// lifecycle hooks
onMounted(async () => {
  ignoreList.value = projectStore.activeProject.ignore_list;
});
</script>


<style scoped>
.ignore-list-header {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
}

.ignore-description {
  font-size: 14px;
  color: hsl(var(--foreground));
  margin: 0;
  line-height: 1.4;
  padding: .5rem;
}

.template-actions {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  justify-content: space-between;
}

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: hsl(var(--foreground));
  justify-content: space-between;
  
  padding: 1rem;
  background-color: hsl(var(--destructive));
  background-color: hsl(var(--background));
  
}
</style>

