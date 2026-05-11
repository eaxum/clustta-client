<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <div class="ignore-list-header">
        <p class="ignore-description">
          {{ $t('settings.ignoreListDescriptionPre') }}
          <Chip :label="'*.log'" :isStatic="true" :color="'var(--steel)'" />
          {{ $t('settings.ignoreListDescriptionMid') }}
          <Chip :label="'/FolderName'" :isStatic="true" :color="'var(--steel)'" />
          {{ $t('settings.ignoreListDescriptionPost') }}
        </p>

        <div class="template-actions">
          <DropDownBox :items="templateNames" :selectedItem="selectedTemplate" :onSelect="applyTemplate" :placeHolder="$t('settings.selectTemplate')" :fixedWidth="true">
            <template #itemAction="{ value, close }">
              <img v-if="isUserPreset(value)" :src="getAppIcon('trash')" class="preset-delete-icon" v-tooltip="$t('settings.deleteIgnorePresetTooltip')" @click="confirmDeletePreset(value, close)" />
            </template>
          </DropDownBox>

          <div class="template-actions-right">
            <ActionButton :icon="getAppIcon('floppy-disk')" :label="$t('settings.saveAsPreset')" @click="openSavePresetModal" v-tooltip="$t('settings.saveAsPreset')" :isDisabled="!ignoreList.length" />

            <ActionButton :icon="getAppIcon('broom')" :label="$t('settings.clearAll')" @click="prepClearAll" v-tooltip="$t('settings.clearAll')" />
          </div>
        </div>
      </div>

      <IgnoreListBox :placeholder="$t('placeholders.addIgnoreItem')" :selectedItems="ignoreList" @itemAdded="addIgnoredItem"
        @itemRemoved="removeIgnoredItem" />

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { ProjectService, SettingsService } from '@/services';
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
const userPresets = ref({});

// computed
// Combined list of built-in template names followed by user-saved preset names.
const templateNames = computed(() => {
  const userNames = Object.keys(userPresets.value);
  return [...ignoreTemplateNames, ...userNames];
});

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
  const entries = ignoreTemplates[templateName] ?? userPresets.value[templateName];
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

// Confirms and deletes a user-saved ignore list preset by name.
const confirmDeletePreset = (name, closeDropdown) => {
  if (!isUserPreset(name)) return;
  if (typeof closeDropdown === 'function') closeDropdown();
  trayStates.popUpModalTitle = t('settings.deleteIgnorePresetTitle', { name });
  trayStates.popUpModalMessage = t('settings.deleteIgnorePresetMessage');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = async () => {
    try {
      await SettingsService.RemoveIgnoreListPreset(name);
      delete userPresets.value[name];
      userPresets.value = { ...userPresets.value };
      notificationStore.addNotification(t('notifications.ignorePresetDeleted'), '', 'success');
    } catch (error) {
      notificationStore.errorNotification(t('notifications.errorDeletingIgnorePreset'), error);
    } finally {
      modals.disableAllModals();
    }
  };
  modals.setModalVisibility('popUpModal', true);
};

// Returns true if the given preset name belongs to the user's saved presets.
const isUserPreset = (name) => {
  return Object.prototype.hasOwnProperty.call(userPresets.value, name);
};

// Loads the user's saved ignore list presets from settings.
const loadUserPresets = async () => {
  try {
    const presets = await SettingsService.GetIgnoreListPresets();
    userPresets.value = presets || {};
  } catch (error) {
    userPresets.value = {};
  }
};

// Opens the modal to save the current ignore list as a named preset.
const openSavePresetModal = () => {
  if (!ignoreList.value.length) return;
  modals.setModalVisibility('saveIgnorePresetModal', true);
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
  await loadUserPresets();
  emitter.on('ignore-preset-added', loadUserPresets);
});

onUnmounted(() => {
  emitter.off('ignore-preset-added', loadUserPresets);
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
  color: var(--white);
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

.template-actions-right {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.preset-delete-icon {
  width: 18px;
  height: 18px;
  opacity: .6;
  cursor: pointer;
  transition: opacity .15s ease;
}

.preset-delete-icon:hover {
  opacity: 1;
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
  color: white;
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
  border-radius: var(--very-large-radius);
}
</style>

